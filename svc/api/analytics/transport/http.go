// Package transport provides HTTP transport layer for the analytics API.
package transport

import (
	"time"

	"url-shorterner/internal/http"
	"url-shorterner/internal/rate"
	"url-shorterner/svc/analytics/app"
	"url-shorterner/svc/analytics/entity"

	"github.com/gin-gonic/gin"
)

// AnalyticsResponse represents the analytics data for a short code
//
// swagger:model AnalyticsResponse
type AnalyticsResponse struct {
	// The short code for which analytics are retrieved
	// example: abc123
	ShortCode string `json:"short_code"`

	// Total number of clicks for this short code
	// example: 42
	TotalClicks int `json:"total_clicks"`

	// Number of unique IP addresses that clicked the URL
	// example: 10
	UniqueIPs int `json:"unique_ips"`

	// Timestamp of the last click (null if no clicks)
	LastClick *time.Time `json:"last_click"`

	// List of detailed click records (paginated)
	Records []*entity.Record `json:"records"`
}

// ErrorResponse represents an error response
//
// swagger:model ErrorResponse
type ErrorResponse struct {
	// Error message describing what went wrong
	// example: error message
	Error string `json:"error"`
}

// AnalyticsAPI defines the HTTP interface for analytics endpoints.
type AnalyticsAPI interface {
	// GetAnalytics retrieves comprehensive click analytics for a short code
	//
	// swagger:operation GET /analytics/{code} analytics getAnalytics
	//
	// Retrieve comprehensive click analytics for a short code.
	//
	// This endpoint provides detailed analytics data including total clicks, unique IPs,
	// last click timestamp, and paginated click records.
	//
	// ---
	// summary: Get analytics for a short code
	// description: |
	//   Retrieve comprehensive click analytics for a short code.
	//
	//   This endpoint provides detailed analytics data including:
	//   - Total number of clicks
	//   - Number of unique IP addresses
	//   - Last click timestamp
	//   - Paginated list of click records with IP, user agent, referer, and timestamp
	//
	//   **Features:**
	//   - Pagination support via limit parameter
	//   - Detailed click records with metadata
	//   - Unique IP tracking
	//   - Timestamp tracking for all clicks
	// tags:
	//   - analytics
	// consumes:
	//   - application/json
	// produces:
	//   - application/json
	// parameters:
	//   - name: code
	//     in: path
	//     required: true
	//     type: string
	//     description: Short code for which to retrieve analytics
	//     example: abc123
	//   - name: limit
	//     in: query
	//     type: integer
	//     required: false
	//     default: 100
	//     minimum: 1
	//     maximum: 1000
	//     description: Limit number of records to return (max 1000)
	// responses:
	//   "200":
	//     description: Analytics data retrieved successfully
	//     schema:
	//       $ref: "#/definitions/AnalyticsResponse"
	//   "400":
	//     description: Invalid request - short code required
	//     schema:
	//       $ref: "#/definitions/ErrorResponse"
	//   "500":
	//     description: Internal server error
	//     schema:
	//       $ref: "#/definitions/ErrorResponse"
	// security:
	//   - ApiKeyAuth: []
	GetAnalytics(*gin.Context)
}

// SetupRouter registers analytics API routes on the provided router.
func SetupRouter(router *gin.Engine, service app.Service, limiter rate.Limiter) {
	apiGroup := http.Router(router, "/", limiter)

	api := NewAnalyticsAPI(service)
	apiGroup.GET("/analytics/:code", api.GetAnalytics)
}
