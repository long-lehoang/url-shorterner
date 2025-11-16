// Package middleware provides HTTP middleware functions for rate limiting, metrics, logging, and error handling.
package middleware

import (
	"net/http"

	"url-shorterner/internal/prometheus"
	"url-shorterner/internal/rate"

	"github.com/gin-gonic/gin"
)

// RateLimit returns a Gin middleware that enforces rate limiting.
func RateLimit(limiter rate.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		identifier := c.ClientIP()
		allowed, err := limiter.Allow(c.Request.Context(), identifier)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "rate limit check failed"})
			c.Abort()
			return
		}

		if !allowed {
			prometheus.RateLimitBlockedTotal.WithLabelValues(identifier).Inc()
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			c.Abort()
			return
		}

		c.Next()
	}
}
