// Package http provides common HTTP utilities and router setup functions.
package http

import (
	"url-shorterner/internal/middleware"
	"url-shorterner/internal/rate"

	"github.com/gin-gonic/gin"
)

// Router creates a router group with common middleware applied.
func Router(router *gin.Engine, path string, limiter rate.Limiter) *gin.RouterGroup {
	group := router.Group(path)
	group.Use(middleware.Logger())
	group.Use(middleware.RateLimit(limiter))
	group.Use(middleware.Prometheus())
	group.Use(middleware.ErrorHandler())
	return group
}
