package services

import (
	"context"
	"database/sql"
	"fmt"
	"meli-challenge/api/classifiers"
	llm "meli-challenge/api/llm"
	"meli-challenge/api/models"
	"meli-challenge/api/repositories"
	"meli-challenge/logger"
	"strings"
)

type ScanService interface {
	// ExecuteScan scans all non-system schemas on the provided server instance
	ExecuteScan(databaseID int64, externalDB *sql.DB) (int64, error)
	// ExecuteScanV2 scans columns + samples data rows using LLM
	ExecuteScanV2(databaseID int64, externalDB *sql.DB) (int64, error)
	// Update scan history status
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

func (s *scanService) ExecuteScan(databaseID int64, externalDB *sql.DB) (scanID int64, err error) {
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

	// Determine tables to scan: scan all non-system schemas
	tablesRows, err := externalDB.Query(`
		SELECT TABLE_SCHEMA, TABLE_NAME
		FROM information_schema.tables
		WHERE TABLE_TYPE='BASE TABLE'
		  AND TABLE_SCHEMA NOT IN ('mysql','sys','information_schema','performance_schema')
	`)
	if err != nil {
		return scanID, err
	}
	defer tablesRows.Close()

	for tablesRows.Next() {
		var schemaName, tableName string
		if err := tablesRows.Scan(&schemaName, &tableName); err != nil {
			return scanID, err
		}
		// log scanning progress
		// ...existing code...

		// Use package logger for structured logs
		// ...existing code...

		// log scanning progress
		// logger is imported at top

		logger.Infof("Scanning: %s.%s", schemaName, tableName)

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

func (s *scanService) ExecuteScanV2(databaseID int64, externalDB *sql.DB) (scanID int64, err error) {
	// Create history record (status = running)
	scanID, err = s.repoScan.CreateHistory(databaseID)
	if err != nil {
		return 0, err
	}

	// Ensure history status is updated
	defer func() {
		if err != nil {
			_ = s.repoScan.UpdateHistoryStatus(scanID, "failed")
			return
		}
		_ = s.repoScan.UpdateHistoryStatus(scanID, "success")
	}()

	// Load classification rules (valid categories)
	rules, err := s.repoRule.GetAllRules()
	if err != nil {
		return scanID, err
	}
	var categories []string
	for _, r := range rules {
		categories = append(categories, r.TypeName)
	}

	// Init OpenAI client once
	llmClient := llm.NewOpenAIClient()
	ctx := context.Background()

	// Determine tables to scan
	tablesRows, err := externalDB.Query(`
		SELECT TABLE_SCHEMA, TABLE_NAME
		FROM information_schema.tables
		WHERE TABLE_TYPE='BASE TABLE'
		  AND TABLE_SCHEMA NOT IN ('mysql','sys','information_schema','performance_schema')
	`)
	if err != nil {
		return scanID, err
	}
	defer tablesRows.Close()

	for tablesRows.Next() {
		var schemaName, tableName string
		if err := tablesRows.Scan(&schemaName, &tableName); err != nil {
			return scanID, err
		}
		logger.Infof("Scanning (v2): %s.%s", schemaName, tableName)

		// Get columns
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

			// Sample up to 5 values from the column
			query := fmt.Sprintf("SELECT DISTINCT `%s` FROM `%s`.`%s` LIMIT 5", columnName, schemaName, tableName)
			sampleRows, err := externalDB.Query(query)
			if err != nil {
				// Some columns may not be selectable (e.g., blob), continue gracefully
				logger.Warnf("Skipping column %s.%s.%s: %v", schemaName, tableName, columnName, err)
				continue
			}

			var samples []string
			for sampleRows.Next() {
				var val sql.NullString
				if err := sampleRows.Scan(&val); err == nil && val.Valid {
					samples = append(samples, val.String)
				}
			}
			sampleRows.Close()

			// Combine samples into a single string (to give context to LLM)
			sampleText := fmt.Sprintf("Column: %s\nValues: %s", columnName, strings.Join(samples, ", "))

			// Call OpenAI for classification
			infoType, err := llmClient.ClassifySample(ctx, sampleText, categories)
			if err != nil {
				logger.Errorf("LLM classification failed for %s.%s.%s: %v", schemaName, tableName, columnName, err)
				infoType = "N/A"
			}

			// Save result once per column
			result := models.ScanResult{
				SchemaName: schemaName,
				TableName:  tableName,
				ColumnName: columnName,
				InfoType:   infoType,
			}
			if err := s.repoScan.SaveResult(scanID, result); err != nil {
				logger.Errorf("SaveResult exec failed for scanID=%d: %v", scanID, err)
				return scanID, err
			}
		}
		cols.Close()
	}

	return scanID, nil
}
