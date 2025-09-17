package services_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	testifyMock "github.com/stretchr/testify/mock"

	"meli-challenge/api/models"
	"meli-challenge/api/services"
)

// Mock repositories for ScanService
type MockScanRepo struct{ testifyMock.Mock }
type MockRuleRepo struct{ testifyMock.Mock }

// --- ScanRepo methods ---
func (m *MockScanRepo) CreateHistory(databaseID int64) (int64, error) {
	args := m.Called(databaseID)
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockScanRepo) SaveResult(scanID int64, result models.ScanResult) error {
	args := m.Called(scanID, result)
	return args.Error(0)
}
func (m *MockScanRepo) GetResultsByScanID(scanID int64) ([]models.ScanResult, error) {
	args := m.Called(scanID)
	return args.Get(0).([]models.ScanResult), args.Error(1)
}
func (m *MockScanRepo) UpdateHistoryStatus(scanID int64, status string) error {
	args := m.Called(scanID, status)
	return args.Error(0)
}

// --- RuleRepo methods ---
func (m *MockRuleRepo) CreateRule(rule models.ClassificationRule) (int64, error) {
	args := m.Called(rule)
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockRuleRepo) GetAllRules() ([]models.ClassificationRule, error) {
	args := m.Called()
	return args.Get(0).([]models.ClassificationRule), args.Error(1)
}

func TestExecuteScan(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	// Mock query for listing tables in schema
	mock.ExpectQuery("SELECT TABLE_SCHEMA, TABLE_NAME FROM information_schema.tables").
		WillReturnRows(sqlmock.NewRows([]string{"TABLE_SCHEMA", "TABLE_NAME"}).
			AddRow("target_sample_db", "users"))

	// Mock query for listing columns in "users"
	mock.ExpectQuery("SELECT COLUMN_NAME FROM information_schema.columns").
		WithArgs("target_sample_db", "users").
		WillReturnRows(sqlmock.NewRows([]string{"COLUMN_NAME"}).
			AddRow("username"))

	scanRepo := new(MockScanRepo)
	ruleRepo := new(MockRuleRepo)

	// Setup mock expectations
	ruleRepo.On("GetAllRules").Return([]models.ClassificationRule{
		{ID: 1, TypeName: "USERNAME", Regex: "(?i)^user(name)?$"},
	}, nil)

	// Return scanID = 1 for history
	scanRepo.On("CreateHistory", int64(1)).Return(int64(1), nil)
	// Accept any column scan results
	scanRepo.On("SaveResult", int64(1), testifyMock.Anything).Return(nil)
	// Accept either "success" or "failed" status
	scanRepo.On("UpdateHistoryStatus", int64(1), testifyMock.Anything).Return(nil)

	svc := services.NewScanService(scanRepo, ruleRepo)

	// Run ExecuteScan
	scanID, err := svc.ExecuteScan(1, db)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, int64(1), scanID)

	// Verify that SaveResult and UpdateHistoryStatus were called
	scanRepo.AssertCalled(t, "SaveResult", int64(1), testifyMock.Anything)
	scanRepo.AssertCalled(t, "UpdateHistoryStatus", int64(1), testifyMock.Anything)
}
