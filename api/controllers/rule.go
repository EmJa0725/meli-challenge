package controllers

import (
	"meli-challenge/api/models"
	"meli-challenge/api/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RuleController struct {
	Service services.RuleService
}

func NewRuleController(s services.RuleService) *RuleController {
	return &RuleController{Service: s}
}

func (ctrl *RuleController) GetAllRules(c *gin.Context) {
	rules, err := ctrl.Service.GetAllRules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rules)
}

func (ctrl *RuleController) CreateRule(c *gin.Context) {
	var req models.ClassificationRule
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := ctrl.Service.CreateRule(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}
