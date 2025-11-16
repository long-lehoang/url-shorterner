// Package middleware provides HTTP middleware functions for rate limiting, metrics, logging, and error handling.
package middleware

import (
	"time"

	"url-shorterner/internal/log"

	"github.com/gin-gonic/gin"
)

// Logger returns a Gin middleware that logs HTTP requests and responses.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		if raw != "" {
			path = path + "?" + raw
		}

		if errorMessage != "" {
			log.Error("[%s] %s %s %d %v | %s | %s",
				clientIP,
				method,
				path,
				statusCode,
				latency,
				errorMessage,
				c.Request.UserAgent(),
			)
		} else {
			log.Info("[%s] %s %s %d %v | %s",
				clientIP,
				method,
				path,
				statusCode,
				latency,
				c.Request.UserAgent(),
			)
		}
	}
}
