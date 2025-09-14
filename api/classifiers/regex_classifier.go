package classifiers

import (
	"meli-challenge/api/models"
	"regexp"
)

type RegexClassifier struct {
	Pattern *regexp.Regexp
	Type    string
}

func NewRegexClassifier(rule models.ClassificationRule) (*RegexClassifier, error) {
	compiled, err := regexp.Compile(rule.Regex)
	if err != nil {
		return nil, err
	}
	return &RegexClassifier{
		Pattern: compiled,
		Type:    rule.TypeName,
	}, nil
}

func (rc *RegexClassifier) Match(column string) bool {
	return rc.Pattern.MatchString(column)
}

func (rc *RegexClassifier) InfoType() string {
	return rc.Type
}
