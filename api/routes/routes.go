package routes

import (
	"meli-challenge/api/controllers"
	"meli-challenge/api/middleware"
	"meli-challenge/api/repositories"
	"meli-challenge/api/services"
	"meli-challenge/config"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	db := config.InitDB()

	// Repositories
	repoDB := repositories.NewDatabaseRepository(db)
	repoScan := repositories.NewScanRepository(db)
	repoRule := repositories.NewRuleRepository(db)

	// Services
	serviceDB := services.NewDatabaseService(repoDB)
	serviceScan := services.NewScanService(repoScan, repoRule)
	serviceRule := services.NewRuleService(repoRule)

	// Controllers
	controllerDB := controllers.NewDatabaseController(serviceDB)
	controllerScan := controllers.NewScanController(serviceScan, db)
	controllerRule := controllers.NewRuleController(serviceRule)

	// Apply API key middleware to all v1 routes
	v1 := r.Group("/api/v1", middleware.APIKeyAuthMiddleware())
	{
		v1.GET("/ping", controllers.Ping)
		v1.POST("/database", controllerDB.CreateDatabase)
		v1.POST("/database/scan/:id", controllerScan.ExecuteScan)
		v1.GET("/database/scan/:id", controllerScan.GetScanResults)
		v1.POST("/classification/rule", controllerRule.CreateRule)
		v1.GET("/classification/rules", controllerRule.GetAllRules)
	}

	// Apply API key middleware also to v2 routes
	v2 := r.Group("/api/v2", middleware.APIKeyAuthMiddleware())
	{
		v2.POST("/database/scan/:id", controllerScan.ExecuteScanV2)
	}
}
