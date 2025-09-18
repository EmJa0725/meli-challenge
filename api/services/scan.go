package services

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"meli-challenge/api/classifiers"
	llm "meli-challenge/api/llm"
	"meli-challenge/api/models"
	"meli-challenge/api/repositories"
	"meli-challenge/logger"
)

type ScanService interface {
	// ExecuteScan scans all non-system schemas on the provided server instance
	ExecuteScan(databaseID int64, externalDB *sql.DB) (int64, error)
	// ExecuteScanV2 scans columns + samples data rows using LLM
	ExecuteScanV2(databaseID int64, externalDB *sql.DB) (int64, error)
	// Update scan history status
	UpdateScanStatus(scanID int64, status string) error
	// GetScanResults returns a nested structure grouped by schema -> table -> columns
	GetScanResults(scanID int64) (models.DatabaseResult, error)
}

type scanService struct {
	repoScan repositories.ScanRepository
	repoRule repositories.RuleRepository
}

func NewScanService(repoScan repositories.ScanRepository, repoRule repositories.RuleRepository) ScanService {
	return &scanService{repoScan: repoScan, repoRule: repoRule}
}

func (s *scanService) UpdateScanStatus(scanID int64, status string) error {
	return s.repoScan.UpdateHistoryStatus(scanID, status)
}

func (s *scanService) ExecuteScan(databaseID int64, externalDB *sql.DB) (scanID int64, err error) {
	// Create history record (status = running)
	scanID, err = s.repoScan.CreateHistory(databaseID)
	if err != nil {
		return 0, err
	}

	// Ensure history status is updated to 'success' or 'failed'
	defer func() {
		if err != nil {
			_ = s.repoScan.UpdateHistoryStatus(scanID, "failed")
			return
		}
		_ = s.repoScan.UpdateHistoryStatus(scanID, "success")
	}()

	// Load classification rules
	rules, err := s.repoRule.GetAllRules()
	if err != nil {
		return scanID, err
	}

	// Build dynamic classifiers from rules
	classifiersList, err := classifiers.BuildClassifiers(rules)
	if err != nil {
		return scanID, err
	}

	// Determine tables to scan: scan all non-system schemas
	tablesRows, err := externalDB.Query(`
		SELECT TABLE_SCHEMA, TABLE_NAME
		FROM information_schema.tables
		WHERE TABLE_TYPE='BASE TABLE'
		  AND TABLE_SCHEMA NOT IN ('mysql','sys','information_schema','performance_schema')
	`)
	if err != nil {
		return scanID, err
	}
	defer tablesRows.Close()

	for tablesRows.Next() {
		var schemaName, tableName string
		if err := tablesRows.Scan(&schemaName, &tableName); err != nil {
			return scanID, err
		}
		// log scanning progress
		// ...existing code...

		// Use package logger for structured logs
		// ...existing code...

		// log scanning progress
		// logger is imported at top

		logger.Infof("Scanning: %s.%s", schemaName, tableName)

		// Get columns for the specific schema.table
		cols, err := externalDB.Query(`
			SELECT COLUMN_NAME
			FROM information_schema.columns
			WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
			ORDER BY ORDINAL_POSITION
		`, schemaName, tableName)
		if err != nil {
			return scanID, err
		}

		for cols.Next() {
			var columnName string
			if err := cols.Scan(&columnName); err != nil {
				cols.Close()
				return scanID, err
			}

			// Classify column name using dynamic regex-based classifiers
			infoType := "N/A"
			for _, c := range classifiersList {
				if c.Match(columnName) {
					infoType = c.InfoType()
					break
				}
			}

			// Persist result including schema_name
			result := models.ScanResult{
				SchemaName: schemaName,
				TableName:  tableName,
				ColumnName: columnName,
				InfoType:   infoType,
			}
			if err := s.repoScan.SaveResult(scanID, result); err != nil {
				cols.Close()
				return scanID, err
			}
		}
		cols.Close()
	}

	return scanID, nil
}

func (s *scanService) GetScanResults(scanID int64) (models.DatabaseResult, error) {
	results, err := s.repoScan.GetResultsByScanID(scanID)
	if err != nil {
		return models.DatabaseResult{}, err
	}

	// Build nested structure: schema -> table -> columns
	schemaMap := make(map[string]map[string][]models.ColumnView)

	for _, r := range results {
		schema := r.SchemaName
		if schema == "" {
			schema = "unknown"
		}
		if _, ok := schemaMap[schema]; !ok {
			schemaMap[schema] = make(map[string][]models.ColumnView)
		}
		schemaMap[schema][r.TableName] = append(schemaMap[schema][r.TableName], models.ColumnView{
			ColumnName: r.ColumnName,
			InfoType:   r.InfoType,
		})
	}

	var dbResult models.DatabaseResult
	for schema, tables := range schemaMap {
		var tableViews []models.TableView
		for tableName, cols := range tables {
			tableViews = append(tableViews, models.TableView{
				TableName: tableName,
				Columns:   cols,
			})
		}
		dbResult.Database = append(dbResult.Database, models.SchemaView{
			SchemaName:   schema,
			SchemaTables: tableViews,
		})
	}

	return dbResult, nil
}

func (s *scanService) ExecuteScanV2(databaseID int64, externalDB *sql.DB) (scanID int64, err error) {
	// Create history record (status = running)
	scanID, err = s.repoScan.CreateHistory(databaseID)
	if err != nil {
		return 0, err
	}

	// Ensure history status is updated
	defer func() {
		if err != nil {
			_ = s.repoScan.UpdateHistoryStatus(scanID, "failed")
			return
		}
		_ = s.repoScan.UpdateHistoryStatus(scanID, "success")
	}()

	// Load classification rules (valid categories)
	rules, err := s.repoRule.GetAllRules()
	if err != nil {
		return scanID, err
	}
	var categories []string
	for _, r := range rules {
		categories = append(categories, r.TypeName)
	}

	// Init OpenAI client once
	llmClient := llm.NewOpenAIClient()

	// Configurable concurrency/timeout/rate limiting for LLM calls
	maxConc := 4
	if v := os.Getenv("LLM_CONCURRENCY"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxConc = n
		}
	}
	timeoutMs := 8000
	if v := os.Getenv("LLM_TIMEOUT_MS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			timeoutMs = n
		}
	}
	ratePerSec := 0
	if v := os.Getenv("LLM_RATE_PER_SEC"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			ratePerSec = n
		}
	}

	// Determine tables to scan
	tablesRows, err := externalDB.Query(`
		SELECT TABLE_SCHEMA, TABLE_NAME
		FROM information_schema.tables
		WHERE TABLE_TYPE='BASE TABLE'
		  AND TABLE_SCHEMA NOT IN ('mysql','sys','information_schema','performance_schema')
	`)
	if err != nil {
		return scanID, err
	}
	defer tablesRows.Close()

	// We'll classify columns concurrently using a semaphore to limit parallel LLM calls.
	sem := make(chan struct{}, maxConc)
	var limiter <-chan time.Time
	var rateTicker *time.Ticker
	if ratePerSec > 0 {
		rateTicker = time.NewTicker(time.Second / time.Duration(ratePerSec))
		limiter = rateTicker.C
		defer rateTicker.Stop()
	}

	// Gather columns to process so we can run them concurrently and then persist results.
	type colWork struct {
		schema  string
		table   string
		column  string
		samples []string
	}
	var workItems []colWork

	for tablesRows.Next() {
		var schemaName, tableName string
		if err := tablesRows.Scan(&schemaName, &tableName); err != nil {
			return scanID, err
		}
		logger.Infof("Scanning (v2): %s.%s", schemaName, tableName)

		// Get columns
		cols, err := externalDB.Query(`
			SELECT COLUMN_NAME
			FROM information_schema.columns
			WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
			ORDER BY ORDINAL_POSITION
		`, schemaName, tableName)
		if err != nil {
			return scanID, err
		}

		for cols.Next() {
			var columnName string
			if err := cols.Scan(&columnName); err != nil {
				cols.Close()
				return scanID, err
			}

			// Sample up to 5 values from the column
			query := fmt.Sprintf("SELECT DISTINCT `%s` FROM `%s`.`%s` WHERE `%s` IS NOT NULL LIMIT 5", columnName, schemaName, tableName, columnName)
			sampleRows, err := externalDB.Query(query)
			if err != nil {
				// Some columns may not be selectable (e.g., blob), continue gracefully
				logger.Warnf("Skipping column %s.%s.%s: %v", schemaName, tableName, columnName, err)
				continue
			}

			var samples []string
			for sampleRows.Next() {
				var val sql.NullString
				if err := sampleRows.Scan(&val); err == nil && val.Valid {
					samples = append(samples, val.String)
				}
			}
			sampleRows.Close()

			workItems = append(workItems, colWork{schema: schemaName, table: tableName, column: columnName, samples: samples})
		}
		cols.Close()
	}

	// Process work items concurrently
	var wg sync.WaitGroup
	var mu sync.Mutex
	errs := make([]error, 0)

	for _, wi := range workItems {
		// capture
		wi := wi
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Acquire semaphore slot
			sem <- struct{}{}
			defer func() { <-sem }()

			// Rate limit if configured
			if limiter != nil {
				select {
				case <-limiter:
				}
			}

			// per-call timeout
			cctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutMs)*time.Millisecond)
			defer cancel()

			sampleText := fmt.Sprintf("Column: %s\nValues: %s", wi.column, strings.Join(wi.samples, ", "))
			infoType := "N/A"
			label, err := llmClient.ClassifySample(cctx, sampleText, categories)
			if err != nil {
				logger.Warnf("LLM classify failed for %s.%s.%s: %v", wi.schema, wi.table, wi.column, err)
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			} else if label != "" {
				infoType = label
			}

			result := models.ScanResult{
				SchemaName: wi.schema,
				TableName:  wi.table,
				ColumnName: wi.column,
				InfoType:   infoType,
			}
			if err := s.repoScan.SaveResult(scanID, result); err != nil {
				logger.Errorf("SaveResult exec failed for scanID=%d: %v", scanID, err)
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if len(errs) > 0 {
		// return first error but keep results persisted
		return scanID, errs[0]
	}

	return scanID, nil
}
