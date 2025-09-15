package repositories

import (
	"database/sql"
	"meli-challenge/api/models"
)

type RuleRepository interface {
	GetAllRules() ([]models.ClassificationRule, error)
	CreateRule(rule models.ClassificationRule) (int64, error)
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

func (r *ruleRepository) CreateRule(rule models.ClassificationRule) (int64, error) {
	stmt, err := r.conn.Prepare("INSERT INTO classification_rules(type_name, regex) VALUES (?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(rule.TypeName, rule.Regex)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}
