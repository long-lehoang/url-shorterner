// Package middleware provides HTTP middleware functions for rate limiting and metrics.
package middleware

import (
	"fmt"
	"net/http"
	"time"

	"url-shorterner/internal/prometheus"
	"url-shorterner/internal/rate"

	"github.com/gin-gonic/gin"
)

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

func Prometheus() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()
		prometheus.HTTPRequestsTotal.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			fmt.Sprintf("%d", c.Writer.Status()),
		).Inc()
		_ = duration // Can be used for latency metrics if needed
	}
}
