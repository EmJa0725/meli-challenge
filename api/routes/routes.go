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

	// Controllers
	controllerDB := controllers.NewDatabaseController(serviceDB)
	controllerScan := controllers.NewScanController(serviceScan, db)

	v1 := r.Group("/api/v1")
	{
		v1.GET("/ping", controllers.Ping)
		v1.POST("/database", controllerDB.CreateDatabase)
		v1.POST("/database/scan/:id", controllerScan.ExecuteScan)
	}
}
