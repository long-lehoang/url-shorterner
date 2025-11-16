// Package middleware provides HTTP middleware functions for rate limiting, metrics, logging, and error handling.
package middleware

import (
	"net/http"

	"url-shorterner/internal/log"

	"github.com/gin-gonic/gin"
)

// Recovery returns a Gin middleware that recovers from panics and logs the error.
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		log.Error("Panic recovered: %v | %s %s | %s",
			recovered,
			c.Request.Method,
			c.Request.URL.Path,
			c.ClientIP(),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		c.Abort()
	})
}
