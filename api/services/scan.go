package services

import (
	"database/sql"
	"fmt"
	"meli-challenge/api/classifiers"
	"meli-challenge/api/models"
	"meli-challenge/api/repositories"
)

type ScanService interface {
	// ExecuteScan scans either a specific schema (targetDBName != "") or all non-system schemas (targetDBName == "" or "information_schema")
	ExecuteScan(databaseID int64, externalDB *sql.DB, targetDBName string) (int64, error)
	UpdateScanStatus(scanID int64, status string) error
	// GetScanResults returns a nested structure grouped by schema -> table -> columns
	GetScanResults(scanID int64) (models.DatabaseResult, error)
}

type scanService struct {
	repoScan repositories.ScanRepository
	repoRule repositories.RuleRepository
}

func NewScanService(repoScan repositories.ScanRepository, repoRule repositories.RuleRepository) ScanService {
	return &scanService{repoScan: repoScan, repoRule: repoRule}
}

func (s *scanService) UpdateScanStatus(scanID int64, status string) error {
	return s.repoScan.UpdateHistoryStatus(scanID, status)
}

func (s *scanService) ExecuteScan(databaseID int64, externalDB *sql.DB, targetDBName string) (scanID int64, err error) {
	// Create history record (status = running)
	scanID, err = s.repoScan.CreateHistory(databaseID)
	if err != nil {
		return 0, err
	}

	// Ensure history status is updated to 'success' or 'failed'
	defer func() {
		if err != nil {
			_ = s.repoScan.UpdateHistoryStatus(scanID, "failed")
			return
		}
		_ = s.repoScan.UpdateHistoryStatus(scanID, "success")
	}()

	// Load classification rules
	rules, err := s.repoRule.GetAllRules()
	if err != nil {
		return scanID, err
	}

	// Build dynamic classifiers from rules
	classifiersList, err := classifiers.BuildClassifiers(rules)
	if err != nil {
		return scanID, err
	}

	// Determine tables to scan:
	// - if targetDBName is empty or "information_schema" -> scan all non-system schemas
	// - otherwise scan only the provided schema
	var tablesRows *sql.Rows
	if targetDBName == "" || targetDBName == "information_schema" {
		tablesRows, err = externalDB.Query(`
			SELECT TABLE_SCHEMA, TABLE_NAME
			FROM information_schema.tables
			WHERE TABLE_TYPE='BASE TABLE'
			  AND TABLE_SCHEMA NOT IN ('mysql','sys','information_schema','performance_schema')
		`)
	} else {
		tablesRows, err = externalDB.Query(`
			SELECT TABLE_SCHEMA, TABLE_NAME
			FROM information_schema.tables
			WHERE TABLE_TYPE='BASE TABLE' AND TABLE_SCHEMA = ?
		`, targetDBName)
	}
	if err != nil {
		return scanID, err
	}
	defer tablesRows.Close()

	for tablesRows.Next() {
		var schemaName, tableName string
		if err := tablesRows.Scan(&schemaName, &tableName); err != nil {
			return scanID, err
		}
		fmt.Printf("Scanning: %s.%s\n", schemaName, tableName)

		// Get columns for the specific schema.table
		cols, err := externalDB.Query(`
			SELECT COLUMN_NAME
			FROM information_schema.columns
			WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
			ORDER BY ORDINAL_POSITION
		`, schemaName, tableName)
		if err != nil {
			return scanID, err
		}

		for cols.Next() {
			var columnName string
			if err := cols.Scan(&columnName); err != nil {
				cols.Close()
				return scanID, err
			}

			// Classify column name using dynamic regex-based classifiers
			infoType := "N/A"
			for _, c := range classifiersList {
				if c.Match(columnName) {
					infoType = c.InfoType()
					break
				}
			}

			// Persist result including schema_name
			result := models.ScanResult{
				SchemaName: schemaName,
				TableName:  tableName,
				ColumnName: columnName,
				InfoType:   infoType,
			}
			if err := s.repoScan.SaveResult(scanID, result); err != nil {
				cols.Close()
				return scanID, err
			}
		}
		cols.Close()
	}

	return scanID, nil
}

func (s *scanService) GetScanResults(scanID int64) (models.DatabaseResult, error) {
	results, err := s.repoScan.GetResultsByScanID(scanID)
	if err != nil {
		return models.DatabaseResult{}, err
	}

	// Build nested structure: schema -> table -> columns
	schemaMap := make(map[string]map[string][]models.ColumnView)

	for _, r := range results {
		schema := r.SchemaName
		if schema == "" {
			schema = "unknown"
		}
		if _, ok := schemaMap[schema]; !ok {
			schemaMap[schema] = make(map[string][]models.ColumnView)
		}
		schemaMap[schema][r.TableName] = append(schemaMap[schema][r.TableName], models.ColumnView{
			ColumnName: r.ColumnName,
			InfoType:   r.InfoType,
		})
	}

	var dbResult models.DatabaseResult
	for schema, tables := range schemaMap {
		var tableViews []models.TableView
		for tableName, cols := range tables {
			tableViews = append(tableViews, models.TableView{
				TableName: tableName,
				Columns:   cols,
			})
		}
		dbResult.Database = append(dbResult.Database, models.SchemaView{
			SchemaName:   schema,
			SchemaTables: tableViews,
		})
	}

	return dbResult, nil
}
