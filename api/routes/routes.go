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
	repo := repositories.NewDatabaseRepository(db)
	service := services.NewDatabaseService(repo)
	controller := controllers.NewDatabaseController(service)

	v1 := r.Group("/api/v1")
	{
		v1.GET("/ping", controllers.Ping)
		v1.POST("/databases", controller.CreateDatabase)
	}
}
