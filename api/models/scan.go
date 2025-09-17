package models

// ScanResult represents a raw stored result row (no ID exposed in API responses)
type ScanResult struct {
	TableName  string `json:"table_name"`
	ColumnName string `json:"column_name"`
	InfoType   string `json:"info_type"`
	SchemaName string `json:"schema_name,omitempty"`
}

// ColumnView is used in API responses to describe a column and its detected type
type ColumnView struct {
	ColumnName string `json:"column_name"`
	InfoType   string `json:"info_type"`
}

// TableView groups columns under a table in the API response
type TableView struct {
	TableName string       `json:"table_name"`
	Columns   []ColumnView `json:"columns"`
}

// SchemaView groups tables under a schema in the API response
type SchemaView struct {
	SchemaName   string      `json:"schema_name"`
	SchemaTables []TableView `json:"schema_tables"`
}

// DatabaseResult is the top-level response model returned by GetScanResults
type DatabaseResult struct {
	Database []SchemaView `json:"database"`
}
