package main

import (
	"meli-challenge/api/routes"

	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	// Register routes
	routes.RegisterRoutes(r)
	// Start the server
	r.Run(":" + os.Getenv("API_PORT")) // listen and serve on
}
