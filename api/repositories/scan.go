package repositories

import (
	"database/sql"
	"meli-challenge/api/models"
)

type ScanRepository interface {
	CreateHistory(databaseId int64) (int64, error)
	SaveResult(scanID int64, result models.ScanResult) error
}

type scanRepository struct {
	conn *sql.DB
}

func NewScanRepository(conn *sql.DB) ScanRepository {
	return &scanRepository{conn: conn}
}

func (r *scanRepository) CreateHistory(databaseId int64) (int64, error) {
	stmt, err := r.conn.Prepare("INSERT INTO scan_history(database_id) VALUES(?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(databaseId)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *scanRepository) SaveResult(scanID int64, result models.ScanResult) error {
	stmt, err := r.conn.Prepare("INSERT INTO scan_results(scan_id, table_name, column_name, info_type) VALUES (?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(scanID, result.TableName, result.ColumnName, result.InfoType)
	return err
}
