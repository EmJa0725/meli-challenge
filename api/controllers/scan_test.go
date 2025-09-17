package controllers_test

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"meli-challenge/api/controllers"
	"meli-challenge/api/models"
)

// DummyScanService implements ScanService for testing
type DummyScanService struct{}

func (d *DummyScanService) ExecuteScan(databaseID int64, externalDB *sql.DB) (int64, error) {
	return 123, nil
}

func (d *DummyScanService) GetScanResults(scanID int64) (models.DatabaseResult, error) {
	return models.DatabaseResult{
		Database: []models.SchemaView{
			{
				SchemaName: "target_sample_db",
				SchemaTables: []models.TableView{
					{
						TableName: "users",
						Columns: []models.ColumnView{
							{ColumnName: "username", InfoType: "USERNAME"},
						},
					},
				},
			},
		},
	}, nil
}

func (d *DummyScanService) UpdateScanStatus(scanID int64, status string) error {
	return nil
}

func TestGetScanResults(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// db is not needed here, we can pass nil
	ctrl := controllers.NewScanController(&DummyScanService{}, nil)
	r.GET("/api/v1/database/scan/:id", ctrl.GetScanResults)

	req, _ := http.NewRequest("GET", "/api/v1/database/scan/123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "USERNAME")
}
