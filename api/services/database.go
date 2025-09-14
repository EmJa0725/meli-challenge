package services

import (
	"meli-challenge/api/models"
	"meli-challenge/api/repositories"
)

type DatabaseService interface {
	RegisterDatabase(dbConfig models.Database) (int64, error)
}

type databaseService struct {
	repo repositories.DatabaseRepository
}

func NewDatabaseService(repo repositories.DatabaseRepository) DatabaseService {
	return &databaseService{repo: repo}
}

func (s *databaseService) RegisterDatabase(dbConfig models.Database) (int64, error) {
	return s.repo.Create(dbConfig)
}
