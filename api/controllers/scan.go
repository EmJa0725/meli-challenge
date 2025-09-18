package controllers

import (
	"database/sql"
	"fmt"
	"html/template"
	"meli-challenge/api/services"
	"meli-challenge/logger"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ScanController struct {
	Service services.ScanService
	DB      *sql.DB // Connection to internal Database
}

func NewScanController(service services.ScanService, db *sql.DB) *ScanController {
	return &ScanController{Service: service, DB: db}
}

func (ctrl *ScanController) ExecuteScan(c *gin.Context) {
	idParam := c.Param("id")
	dbID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid database ID"})
		return
	}

	// obtain database connection details from internal DB
	row := ctrl.DB.QueryRow("SELECT host, port, username, password FROM `external_databases` WHERE id = ?", dbID)
	var host, username, password string
	var port int
	if err := row.Scan(&host, &port, &username, &password); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Database not found"})
		return
	}

	// Always connect to information_schema so the service can query any schema on the server.
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/information_schema", username, password, host, port)
	externalDB, err := sql.Open("mysql", dsn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer externalDB.Close()

	logger.Infof("Starting scan for database id=%d host=%s port=%d", dbID, host, port)

	// Execute scan; service will scan all non-system schemas by connecting to information_schema
	scanID, err := ctrl.Service.ExecuteScan(dbID, externalDB)
	if err != nil {
		// Ensure scan history is marked as failed even if the error occurred before service updated it
		_ = ctrl.Service.UpdateScanStatus(scanID, "failed")
		logger.Errorf("Scan failed for database id=%d: %v", dbID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logger.Infof("Scan completed for database id=%d scan_id=%d", dbID, scanID)

	c.JSON(http.StatusCreated, gin.H{"scan_id": scanID})
}

func (ctrl *ScanController) GetScanResults(c *gin.Context) {
	idParam := c.Param("id")
	scanID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	dbResult, err := ctrl.Service.GetScanResults(scanID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(dbResult.Database) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "no results for this scan"})
		return
	}

	c.JSON(http.StatusOK, dbResult)
}

// RenderScanReport returns an HTML report summarizing a scan results with metrics.
func (ctrl *ScanController) RenderScanReport(c *gin.Context) {
	idParam := c.Param("id")
	scanID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid id")
		return
	}

	dbResult, err := ctrl.Service.GetScanResults(scanID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	// Fetch scan status from internal scan_history
	var scanStatus string
	row := ctrl.DB.QueryRow("SELECT status FROM scan_history WHERE id = ?", scanID)
	if err := row.Scan(&scanStatus); err != nil {
		// default when not found or error
		scanStatus = "unknown"
	}

	// Compute overall counts and per-table breakdown
	totalCols := 0
	typeCounts := make(map[string]int)
	typeOrder := make([]string, 0)

	type tableSummary struct {
		Schema     string
		Table      string
		Total      int
		TypeCounts map[string]int
	}

	var tables []tableSummary

	for _, schema := range dbResult.Database {
		for _, tbl := range schema.SchemaTables {
			ts := tableSummary{Schema: schema.SchemaName, Table: tbl.TableName, TypeCounts: make(map[string]int)}
			for _, col := range tbl.Columns {
				totalCols++
				ts.Total++
				ts.TypeCounts[col.InfoType]++
				if _, ok := typeCounts[col.InfoType]; !ok {
					typeOrder = append(typeOrder, col.InfoType)
				}
				typeCounts[col.InfoType]++
			}
			tables = append(tables, ts)
		}
	}

	// Build template
	const tpl = `<!doctype html>
<html>
<head><meta charset="utf-8"><title>Scan Report {{.ScanID}}</title>
<style>body{font-family:Arial,Helvetica,sans-serif}table{border-collapse:collapse;width:100%}th,td{border:1px solid #ddd;padding:8px}th{background:#f2f2f2;text-align:left}</style>
</head>
<body>
	<h1>Scan Report {{.ScanID}}</h1>
	<p>Scan status: <strong style="color:{{.StatusColor}}">{{.Status}}</strong></p>
	<p>Total columns scanned: {{.Total}}</p>

	<h2>By Info Type</h2>
	<table>
		<tr><th>Info Type</th><th>Count</th><th>Percentage</th></tr>
		{{range $t, $c := .TypeCounts}}
		<tr><td>{{$t}}</td><td>{{$c}}</td><td>{{printf "%.2f" (mul (div $c $.Total) 100) }}%</td></tr>
		{{end}}
	</table>

	<h2>Per Table</h2>
	{{range .Tables}}
		<h3>{{.Schema}}.{{.Table}} ({{.Total}} cols)</h3>
		<table>
			<tr><th>Info Type</th><th>Count</th></tr>
			{{range $k,$v := .TypeCounts}}
				<tr><td>{{$k}}</td><td>{{$v}}</td></tr>
			{{end}}
		</table>
	{{end}}

</body>
</html>`

	// prepare template with helper functions
	funcMap := template.FuncMap{
		"div": func(a, b int) float64 {
			if b == 0 {
				return 0
			}
			return float64(a) / float64(b)
		},
		"mul": func(a float64, b int) float64 { return a * float64(b) },
	}
	t, err := template.New("report").Funcs(funcMap).Parse(tpl)
	if err != nil {
		logger.Errorf("template parse error: %v", err)
		c.String(http.StatusInternalServerError, "template error")
		return
	}

	data := struct {
		ScanID      int64
		Status      string
		StatusColor string
		Total       int
		TypeCounts  map[string]int
		Tables      []tableSummary
	}{
		ScanID: scanID,
		Status: scanStatus,
		StatusColor: func() string {
			switch scanStatus {
			case "success":
				return "green"
			case "failed":
				return "red"
			case "running":
				return "orange"
			default:
				return "gray"
			}
		}(),
		Total:      totalCols,
		TypeCounts: typeCounts,
		Tables:     tables,
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := t.Execute(c.Writer, data); err != nil {
		logger.Errorf("template execute error: %v", err)
	}
}

// ExecuteScanV2 runs column-based + data sampling classification using LLM
func (ctrl *ScanController) ExecuteScanV2(c *gin.Context) {
	idParam := c.Param("id")
	dbID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid database ID"})
		return
	}

	// obtain database connection details from internal DB
	row := ctrl.DB.QueryRow("SELECT host, port, username, password FROM `external_databases` WHERE id = ?", dbID)
	var host, username, password string
	var port int
	if err := row.Scan(&host, &port, &username, &password); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Database not found"})
		return
	}

	// Always connect to information_schema so the service can query any schema on the server.
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/information_schema", username, password, host, port)
	externalDB, err := sql.Open("mysql", dsn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer externalDB.Close()

	logger.Infof("Starting scan v2 for database id=%d host=%s port=%d", dbID, host, port)

	scanID, err := ctrl.Service.ExecuteScanV2(dbID, externalDB)
	if err != nil {
		_ = ctrl.Service.UpdateScanStatus(scanID, "failed")
		logger.Errorf("Scan v2 failed for database id=%d: %v", dbID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Infof("Scan v2 completed for database id=%d scan_id=%d", dbID, scanID)
	c.JSON(http.StatusCreated, gin.H{"scan_id": scanID})
}
