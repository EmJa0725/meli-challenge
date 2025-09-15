package services

import (
	"meli-challenge/api/models"
	"meli-challenge/api/repositories"
)

type RuleService interface {
	GetAllRules() ([]models.ClassificationRule, error)
	CreateRule(rule models.ClassificationRule) (int64, error)
}

type ruleService struct {
	repo repositories.RuleRepository
}

func NewRuleService(repo repositories.RuleRepository) RuleService {
	return &ruleService{repo}
}

func (s *ruleService) GetAllRules() ([]models.ClassificationRule, error) {
	return s.repo.GetAllRules()
}

func (s *ruleService) CreateRule(rule models.ClassificationRule) (int64, error) {
	return s.repo.CreateRule(rule)
}
