package repositories

import (
	"database/sql"
	"meli-challenge/api/models"
)

type DatabaseRepository interface {
	Create(dbConfig models.Database) (int64, error)
}

type databaseRepository struct {
	conn *sql.DB
}

func NewDatabaseRepository(conn *sql.DB) DatabaseRepository {
	return &databaseRepository{conn: conn}
}

func (r *databaseRepository) Create(dbConfig models.Database) (int64, error) {
	stmt, err := r.conn.Prepare("INSERT INTO `external_databases` (`host`, `port`, `username`, `password`) VALUES(?, ?, ?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(dbConfig.Host, dbConfig.Port, dbConfig.Username, dbConfig.Password)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return 0, err
	}

	return id, nil
}
