package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"url-shorterner/internal/prometheus"
	"url-shorterner/internal/rate"
)

func RateLimitMiddleware(limiter rate.Limiter) gin.HandlerFunc {
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

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

