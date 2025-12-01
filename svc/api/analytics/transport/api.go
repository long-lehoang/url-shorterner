// Package transport provides HTTP handler implementations for the analytics API.
package transport

import (
	"errors"
	"net/http"
	"strconv"

	"url-shorterner/svc/analytics/app"

	"github.com/gin-gonic/gin"
)

type api struct {
	service app.Service
}

// NewAnalyticsAPI creates a new analytics API handler instance.
func NewAnalyticsAPI(service app.Service) AnalyticsAPI {
	return &api{service: service}
}

func (a *api) GetAnalytics(c *gin.Context) {
	shortCode := c.Param("code")
	if shortCode == "" {
		c.Error(errors.New("short code is required")) //nolint:errcheck // Error is handled by ErrorHandler middleware
		return
	}

	limit := 100
	if limitParam := c.Query("limit"); limitParam != "" {
		if parsedLimit := parseInt(limitParam); parsedLimit > 0 && parsedLimit <= 1000 {
			limit = parsedLimit
		}
	}

	stats, err := a.service.GetStats(c.Request.Context(), shortCode)
	if err != nil {
		c.Error(err) //nolint:errcheck // Error is handled by ErrorHandler middleware
		return
	}

	records, err := a.service.GetAnalytics(c.Request.Context(), shortCode, limit)
	if err != nil {
		c.Error(err) //nolint:errcheck // Error is handled by ErrorHandler middleware
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"short_code":   shortCode,
		"total_clicks": stats.TotalClicks,
		"unique_ips":   stats.UniqueIPs,
		"last_click":   stats.LastClick,
		"records":      records,
	})
}

func parseInt(s string) int {
	result, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return result
}
