package classifiers

import (
	"meli-challenge/api/models"
)

// BuildClassifiers creates a list of RegexClassifiers from the given classification rules.
func BuildClassifiers(rules []models.ClassificationRule) ([]*RegexClassifier, error) {
	var result []*RegexClassifier
	for _, r := range rules {
		rc, err := NewRegexClassifier(r)
		if err != nil {
			return nil, err
		}
		result = append(result, rc)
	}
	return result, nil
}
