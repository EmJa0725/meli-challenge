package classifiers_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"meli-challenge/api/classifiers"
	"meli-challenge/api/models"
)

func TestBuildClassifiers_AllRules_PositiveAndNegative(t *testing.T) {
	rules := []models.ClassificationRule{
		{TypeName: "EMAIL_ADDRESS", Regex: "(?i)email"},
		{TypeName: "USERNAME", Regex: "(?i)^user(name)?$"},
		{TypeName: "CREDIT_CARD_NUMBER", Regex: "(?i)^(credit[_ ]?card(_?number)?|card[_ ]?number)$"},
		{TypeName: "FIRST_NAME", Regex: "(?i)^first(_?name)?$"},
		{TypeName: "LAST_NAME", Regex: "(?i)^last(_?name)?$"},
		{TypeName: "IP_ADDRESS", Regex: "(?i)^ip(_address)?$"},
		{TypeName: "PHONE_NUMBER", Regex: "(?i)^phone(_number)?$"},
		{TypeName: "DATE", Regex: "(?i)^(created_at|updated_at|.*_date)$"},
		{TypeName: "ADDRESS", Regex: "(?i)^address(_.*)?$"},
		{TypeName: "POSTAL_CODE", Regex: "(?i)^(postal|zip)_?code$"},
		{TypeName: "SSN", Regex: "(?i)^ssn$"},
		{TypeName: "PASSWORD", Regex: "(?i)^password$"},
	}

	classifiersList, err := classifiers.BuildClassifiers(rules)
	assert.NoError(t, err)
	assert.Len(t, classifiersList, len(rules))

	// Positive examples
	positives := map[string]string{
		"EMAIL_ADDRESS":      "useremail",
		"USERNAME":           "username",
		"CREDIT_CARD_NUMBER": "credit_card_number",
		"FIRST_NAME":         "first_name",
		"LAST_NAME":          "last_name",
		"IP_ADDRESS":         "ip_address",
		"PHONE_NUMBER":       "phone",
		"DATE":               "created_at",
		"ADDRESS":            "address_line1",
		"POSTAL_CODE":        "postal_code",
		"SSN":                "ssn",
		"PASSWORD":           "password",
	}

	// Negative examples (should not match)
	negatives := map[string]string{
		"EMAIL_ADDRESS":      "comment_text",  // not an email
		"USERNAME":           "user_id",       // should not match "id"
		"CREDIT_CARD_NUMBER": "cardboard_box", // should not match
		"FIRST_NAME":         "nickname",      // different semantic
		"LAST_NAME":          "lastname_hint", // too loose
		"IP_ADDRESS":         "description",   // false positive risk
		"PHONE_NUMBER":       "microphone",    // substring trap
		"DATE":               "validate",      // contains "date" but not a real date column
		"ADDRESS":            "email_address", // could wrongly overlap
		"POSTAL_CODE":        "code_review",   // generic "code"
		"SSN":                "session_id",    // "ssn" substring
		"PASSWORD":           "password_hint", // not the real password
	}

	for _, c := range classifiersList {
		// positive assertion
		if sample, ok := positives[c.InfoType()]; ok {
			assert.Truef(t, c.Match(sample), "expected %s to match rule %s", sample, c.InfoType())
		}
		// negative assertion
		if sample, ok := negatives[c.InfoType()]; ok {
			assert.Falsef(t, c.Match(sample), "expected %s NOT to match rule %s", sample, c.InfoType())
		}
	}
}
