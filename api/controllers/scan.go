package controllers

import (
	"database/sql"
	"fmt"
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
