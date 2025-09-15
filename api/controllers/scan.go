package controllers

import (
	"database/sql"
	"fmt"
	"meli-challenge/api/services"
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
	row := ctrl.DB.QueryRow("SELECT host, port, username, password FROM databases WHERE id = ?", dbID)
	var host, username, password string
	var port int
	if err := row.Scan(&host, &port, &username, &password); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Database not found"})
		return
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", username, password, host, port, "information_schema")
	externalDB, err := sql.Open("mysql", dsn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer externalDB.Close()

	scanID, err := ctrl.Service.ExecuteScan(dbID, externalDB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"scan_id": scanID})
}

func (ctrl *ScanController) GetScanResults(c *gin.Context) {
	idParam := c.Param("id")
	scanID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	results, err := ctrl.Service.GetScanResults(scanID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(results) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "no results for this scan"})
		return
	}

	c.JSON(http.StatusOK, results)
}
