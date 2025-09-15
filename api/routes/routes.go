package routes

import (
	"meli-challenge/api/controllers"
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

	v1 := r.Group("/api/v1")
	{
		v1.GET("/ping", controllers.Ping)
		v1.POST("/database", controllerDB.CreateDatabase)
		v1.POST("/database/scan/:id", controllerScan.ExecuteScan)
		v1.GET("/database/scan/:id", controllerScan.GetScanResults)
		v1.POST("/classification/rule", controllerRule.CreateRule)
		v1.GET("/classification/rules", controllerRule.GetAllRules)
	}
}
