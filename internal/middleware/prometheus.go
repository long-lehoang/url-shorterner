// Package middleware provides HTTP middleware functions for rate limiting, metrics, logging, and error handling.
package middleware

import (
	"fmt"
	"time"

	"url-shorterner/internal/prometheus"

	"github.com/gin-gonic/gin"
)

// Prometheus returns a Gin middleware that records Prometheus metrics.
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
