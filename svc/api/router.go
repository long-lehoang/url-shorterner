package api

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"url-shorterner/internal/rate"
	"url-shorterner/svc/analytics"
	"url-shorterner/svc/shortener"
)

func SetupRouter(
	shortenerService shortener.Service,
	analyticsService analytics.Service,
	limiter rate.Limiter,
) *gin.Engine {
	router := gin.Default()

	handlers := NewHandlers(shortenerService, analyticsService)

	router.Use(RateLimitMiddleware(limiter))
	router.Use(PrometheusMiddleware())

	api := router.Group("/")
	{
		api.POST("/shorten", handlers.Shorten)
		api.POST("/shorten/batch", handlers.ShortenBatch)
		api.GET("/:code", handlers.Redirect)
		api.GET("/analytics/:code", handlers.GetAnalytics)
	}

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	return router
}

