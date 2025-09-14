package models

type ClassificationRule struct {
	ID       int64  `json:"id"`
	TypeName string `json:"type_name"`
	Regex    string `json:"regex"`
}
