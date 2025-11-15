// Package transport provides HTTP handler implementations for the analytics API.
package transport

import (
	"net/http"
	"strconv"

	"url-shorterner/internal/prometheus"
	"url-shorterner/svc/analytics/app"

	"github.com/gin-gonic/gin"
)

type handlers struct {
	service app.Service
}

func NewHandlers(service app.Service) AnalyticsAPI {
	return &handlers{service: service}
}

func (h *handlers) GetAnalytics(c *gin.Context) {
	shortCode := c.Param("code")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "short code is required"})
		prometheus.HTTPRequestsTotal.WithLabelValues("GET", "/analytics/:code", "400").Inc()
		return
	}

	limit := 100
	if limitParam := c.Query("limit"); limitParam != "" {
		if parsedLimit := parseInt(limitParam); parsedLimit > 0 && parsedLimit <= 1000 {
			limit = parsedLimit
		}
	}

	stats, err := h.service.GetStats(c.Request.Context(), shortCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		prometheus.HTTPRequestsTotal.WithLabelValues("GET", "/analytics/:code", "500").Inc()
		return
	}

	records, err := h.service.GetAnalytics(c.Request.Context(), shortCode, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		prometheus.HTTPRequestsTotal.WithLabelValues("GET", "/analytics/:code", "500").Inc()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"short_code":   shortCode,
		"total_clicks": stats.TotalClicks,
		"unique_ips":   stats.UniqueIPs,
		"last_click":   stats.LastClick,
		"records":      records,
	})
	prometheus.HTTPRequestsTotal.WithLabelValues("GET", "/analytics/:code", "200").Inc()
}

func parseInt(s string) int {
	result, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return result
}
