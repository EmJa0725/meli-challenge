package controllers

import (
	"meli-challenge/api/models"
	"meli-challenge/api/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}

type DatabaseController struct {
	service services.DatabaseService
}

func NewDatabaseController(service services.DatabaseService) *DatabaseController {
	return &DatabaseController{service: service}
}

func (ctrl *DatabaseController) CreateDatabase(c *gin.Context) {
	var req models.Database
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := ctrl.service.RegisterDatabase(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}
