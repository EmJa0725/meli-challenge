package repositories

import (
	"database/sql"
	"meli-challenge/api/models"
	"meli-challenge/logger"
)

type ScanRepository interface {
	CreateHistory(databaseId int64) (int64, error)
	UpdateHistoryStatus(scanID int64, status string) error
	SaveResult(scanID int64, result models.ScanResult) error
	GetResultsByScanID(scanID int64) ([]models.ScanResult, error)
}

type scanRepository struct {
	conn *sql.DB
}

func NewScanRepository(conn *sql.DB) ScanRepository {
	return &scanRepository{conn: conn}
}

func (r *scanRepository) CreateHistory(databaseId int64) (int64, error) {
	stmt, err := r.conn.Prepare("INSERT INTO scan_history(database_id, status) VALUES(?, ?)")
	if err != nil {
		logger.Errorf("CreateHistory prepare failed: %v", err)
		return 0, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(databaseId, "running")
	if err != nil {
		logger.Errorf("CreateHistory exec failed for database_id=%d: %v", databaseId, err)
		return 0, err
	}

	return result.LastInsertId()
}

func (r *scanRepository) UpdateHistoryStatus(scanID int64, status string) error {
	stmt, err := r.conn.Prepare("UPDATE scan_history SET status = ? WHERE id = ?")
	if err != nil {
		logger.Errorf("UpdateHistoryStatus prepare failed: %v", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(status, scanID)
	if err != nil {
		logger.Errorf("UpdateHistoryStatus exec failed for scanID=%d status=%s: %v", scanID, status, err)
	}
	return err
}

func (r *scanRepository) SaveResult(scanID int64, result models.ScanResult) error {
	// Insert schema_name with the result
	stmt, err := r.conn.Prepare("INSERT INTO scan_results(scan_id, schema_name, table_name, column_name, info_type) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		logger.Errorf("SaveResult prepare failed: %v", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(scanID, result.SchemaName, result.TableName, result.ColumnName, result.InfoType)
	if err != nil {
		logger.Errorf("SaveResult exec failed for scanID=%d: %v", scanID, err)
	}
	return err
}

func (r *scanRepository) GetResultsByScanID(scanID int64) ([]models.ScanResult, error) {
	rows, err := r.conn.Query("SELECT schema_name, table_name, column_name, info_type FROM scan_results WHERE scan_id = ? ORDER BY schema_name, table_name, column_name", scanID)
	if err != nil {
		logger.Errorf("GetResultsByScanID query failed for scanID=%d: %v", scanID, err)
		return nil, err
	}
	defer rows.Close()

	var results []models.ScanResult
	for rows.Next() {
		var result models.ScanResult
		if err := rows.Scan(&result.SchemaName, &result.TableName, &result.ColumnName, &result.InfoType); err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}
