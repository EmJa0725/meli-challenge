package models

type ScanResult struct {
	ID         int64  `json:"id"`
	TableName  string `json:"table_name"`
	ColumnName string `json:"column_name"`
	InfoType   string `json:"info_type"`
}
