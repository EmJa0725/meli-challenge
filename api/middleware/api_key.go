package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// APIKeyAuthMiddleware checks the X-API-Key header against the API_KEY env var
func APIKeyAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		expected := os.Getenv("API_KEY")
		if expected == "" {
			// no key configured -> allow (dev mode)
			c.Next()
			return
		}

		key := c.GetHeader("X-API-Key")
		if key == "" || key != expected {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}
