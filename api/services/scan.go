package services

import (
	"database/sql"
	"fmt"
	"meli-challenge/api/classifiers"
	"meli-challenge/api/models"
	"meli-challenge/api/repositories"
)

type ScanService interface {
	ExecuteScan(databaseID int64, externalDB *sql.DB) (int64, error)
	UpdateScanStatus(scanID int64, status string) error
	GetScanResults(scanID int64) (map[string][]models.ScanResult, error)
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
	// Crear historial
	scanID, err = s.repoScan.CreateHistory(databaseID)
	if err != nil {
		return 0, err
	}

	// Ensure we update the status to success or failed depending on outcome
	defer func() {
		if err != nil {
			_ = s.repoScan.UpdateHistoryStatus(scanID, "failed")
			return
		}
		_ = s.repoScan.UpdateHistoryStatus(scanID, "success")
	}()

	// Obtener reglas desde DB interna
	rules, err := s.repoRule.GetAllRules()
	if err != nil {
		return scanID, err
	}

	// Construir clasificadores dinámicos
	classifiersList, err := classifiers.BuildClassifiers(rules)
	if err != nil {
		return scanID, err
	}

	// Listar tablas
	tables, err := externalDB.Query("SHOW TABLES")
	if err != nil {
		return scanID, err
	}
	defer tables.Close()

	for tables.Next() {
		var tableName string
		if err := tables.Scan(&tableName); err != nil {
			return scanID, err
		}
		fmt.Printf("Scanning table: %s\n", tableName)
		// Listar columnas
		cols, err := externalDB.Query(fmt.Sprintf("SHOW COLUMNS FROM %s", tableName))
		if err != nil {
			return 0, err
		}

		for cols.Next() {
			var field, colType, null, key, extra string
			var defaultVal sql.NullString

			if err := cols.Scan(&field, &colType, &null, &key, &defaultVal, &extra); err != nil {
				return scanID, err
			}

			infoType := "N/A"
			for _, c := range classifiersList {
				if c.Match(field) {
					infoType = c.InfoType()
					break
				}
			}

			result := models.ScanResult{
				TableName:  tableName,
				ColumnName: field,
				InfoType:   infoType,
			}

			if err := s.repoScan.SaveResult(scanID, result); err != nil {
				return scanID, err
			}
		}
		cols.Close()
	}

	return scanID, nil
}

func (s *scanService) GetScanResults(scanID int64) (map[string][]models.ScanResult, error) {
	results, err := s.repoScan.GetResultsByScanID(scanID)
	if err != nil {
		return nil, err
	}

	grouped := make(map[string][]models.ScanResult)
	for _, res := range results {
		grouped[res.TableName] = append(grouped[res.TableName], res)
	}
	return grouped, nil
}
