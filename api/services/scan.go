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
}

type scanService struct {
	repoScan repositories.ScanRepository
	repoRule repositories.RuleRepository
}

func NewScanService(repoScan repositories.ScanRepository, repoRule repositories.RuleRepository) ScanService {
	return &scanService{repoScan: repoScan, repoRule: repoRule}
}

func (s *scanService) ExecuteScan(databaseID int64, externalDB *sql.DB) (int64, error) {
	// Crear historial
	scanID, err := s.repoScan.CreateHistory(databaseID)
	if err != nil {
		return 0, err
	}

	// Obtener reglas desde DB interna
	rules, err := s.repoRule.GetAllRules()
	if err != nil {
		return 0, err
	}

	// Construir clasificadores dinámicos
	classifiersList, err := classifiers.BuildClassifiers(rules)
	if err != nil {
		return 0, err
	}

	// Listar tablas
	tables, err := externalDB.Query("SHOW TABLES")
	if err != nil {
		return 0, err
	}
	defer tables.Close()

	for tables.Next() {
		var tableName string
		if err := tables.Scan(&tableName); err != nil {
			return 0, err
		}

		// Listar columnas
		cols, err := externalDB.Query(fmt.Sprintf("SHOW COLUMNS FROM %s", tableName))
		if err != nil {
			return 0, err
		}

		for cols.Next() {
			var field, colType, null, key, defaultVal, extra string
			if err := cols.Scan(&field, &colType, &null, &key, &defaultVal, &extra); err != nil {
				return 0, err
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
				return 0, err
			}
		}
		cols.Close()
	}

	return scanID, nil
}
