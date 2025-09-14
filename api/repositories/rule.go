package repositories

import (
	"database/sql"
	"meli-challenge/api/models"
)

type RuleRepository interface {
	GetAllRules() ([]models.ClassificationRule, error)
}

type ruleRepository struct {
	conn *sql.DB
}

func NewRuleRepository(conn *sql.DB) RuleRepository {
	return &ruleRepository{conn}
}

func (r *ruleRepository) GetAllRules() ([]models.ClassificationRule, error) {
	rows, err := r.conn.Query("SELECT id, type_name, regex FROM classification_rules")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []models.ClassificationRule
	for rows.Next() {
		var rule models.ClassificationRule
		if err := rows.Scan(&rule.ID, &rule.TypeName, &rule.Regex); err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, nil
}
