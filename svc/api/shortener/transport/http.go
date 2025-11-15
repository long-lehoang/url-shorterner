// Package transport provides HTTP transport layer for the shortener API.
package transport

import (
	"url-shorterner/internal/events"
	"url-shorterner/internal/middleware"
	"url-shorterner/internal/rate"
	"url-shorterner/svc/shortener/app"

	"github.com/gin-gonic/gin"
)

// ShortenRequest represents the request body for shortening a URL
//
// swagger:model ShortenRequest
type ShortenRequest struct {
	// The original URL to be shortened
	// required: true
	// example: https://example.com
	URL string `json:"url" binding:"required"`

	// Expiration time in seconds from now (optional)
	// example: 3600
	ExpiresIn *int `json:"expires_in,omitempty"`

	// Custom alias for the shortened URL (optional, must be unique)
	// example: my-custom-alias
	Alias *string `json:"alias,omitempty"`
}

// BatchShortenRequest represents the request body for batch URL shortening
//
// swagger:model BatchShortenRequest
type BatchShortenRequest struct {
	// List of URLs to shorten
	// required: true
	Items []app.BatchItem `json:"items" binding:"required"`
}

// ErrorResponse represents an error response
//
// swagger:model ErrorResponse
type ErrorResponse struct {
	// Error message describing what went wrong
	// example: error message
	Error string `json:"error"`
}

type ShortenerAPI interface {
	// Shorten creates a shortened URL with optional expiration and custom alias
	//
	// swagger:operation POST /shorten shortener shortenURL
	//
	// Create a shortened URL with optional expiration and custom alias.
	//
	// The service generates a unique short code or uses the provided alias if available.
	// If an alias is provided, it must be unique and not already in use.
	// The expiration time is optional and specified in seconds from the current time.
	//
	// ---
	// summary: Shorten a URL
	// description: |
	//   Create a shortened URL with optional expiration and custom alias.
	//
	//   **Features:**
	//   - Automatic short code generation if no alias provided
	//   - Custom alias support (must be unique)
	//   - Optional expiration time
	//   - URL validation and format checking
	// tags:
	//   - shortener
	// consumes:
	//   - application/json
	// produces:
	//   - application/json
	// parameters:
	//   - name: body
	//     in: body
	//     required: true
	//     schema:
	//       $ref: "#/definitions/ShortenRequest"
	// responses:
	//   "200":
	//     description: Successfully created shortened URL
	//     schema:
	//       $ref: "#/definitions/ShortenResponse"
	//   "400":
	//     description: Invalid request - URL format or validation error
	//     schema:
	//       $ref: "#/definitions/ErrorResponse"
	//   "409":
	//     description: Conflict - alias already exists
	//     schema:
	//       $ref: "#/definitions/ErrorResponse"
	//   "500":
	//     description: Internal server error
	//     schema:
	//       $ref: "#/definitions/ErrorResponse"
	// security:
	//   - ApiKeyAuth: []
	Shorten(*gin.Context)

	// ShortenBatch creates shortened URLs for multiple URLs in a single request
	//
	// swagger:operation POST /shorten/batch shortener shortenBatchURLs
	//
	// Create shortened URLs for multiple URLs in a single request.
	//
	// Each URL is processed independently, and errors for individual URLs are included in the response.
	// This endpoint is optimized for bulk operations and maintains the same validation rules as the single URL endpoint.
	//
	// ---
	// summary: Shorten multiple URLs
	// description: |
	//   Create shortened URLs for multiple URLs in a single request.
	//
	//   **Features:**
	//   - Batch processing of multiple URLs
	//   - Independent processing (one failure doesn't affect others)
	//   - Detailed error reporting per URL
	//   - Same validation rules as single URL endpoint
	// tags:
	//   - shortener
	// consumes:
	//   - application/json
	// produces:
	//   - application/json
	// parameters:
	//   - name: body
	//     in: body
	//     required: true
	//     schema:
	//       $ref: "#/definitions/BatchShortenRequest"
	// responses:
	//   "200":
	//     description: Batch processing results with success and error details
	//     schema:
	//       type: object
	//       properties:
	//         results:
	//           type: array
	//           items:
	//             type: object
	//   "400":
	//     description: Invalid request format
	//     schema:
	//       $ref: "#/definitions/ErrorResponse"
	//   "500":
	//     description: Internal server error
	//     schema:
	//       $ref: "#/definitions/ErrorResponse"
	// security:
	//   - ApiKeyAuth: []
	ShortenBatch(*gin.Context)

	// Redirect redirects to the original URL associated with the provided short code
	//
	// swagger:operation GET /{code} shortener redirectToURL
	//
	// Redirects to the original URL associated with the provided short code.
	//
	// This endpoint performs a permanent redirect (301) to the original URL.
	// It validates the short code, checks expiration, and handles various error cases.
	//
	// ---
	// summary: Redirect to original URL
	// description: |
	//   Redirects to the original URL associated with the provided short code.
	//
	//   **Behavior:**
	//   - Returns 301 redirect if URL is valid and not expired
	//   - Returns 404 if short code is not found
	//   - Returns 410 if URL has expired
	//   - Records click analytics for valid redirects
	// tags:
	//   - shortener
	// consumes:
	//   - application/json
	// produces:
	//   - application/json
	// parameters:
	//   - name: code
	//     in: path
	//     required: true
	//     type: string
	//     description: Short code for the URL
	//     example: abc123
	// responses:
	//   "200":
	//     description: Redirect successful (alternative response)
	//   "301":
	//     description: Moved Permanently - Redirect to original URL
	//   "404":
	//     description: URL not found
	//     schema:
	//       $ref: "#/definitions/ErrorResponse"
	//   "410":
	//     description: URL expired
	//     schema:
	//       $ref: "#/definitions/ErrorResponse"
	//   "500":
	//     description: Internal server error
	//     schema:
	//       $ref: "#/definitions/ErrorResponse"
	Redirect(*gin.Context)
}

func SetupRouter(router *gin.Engine, service app.Service, publisher events.Publisher, limiter rate.Limiter) {
	apiGroup := router.Group("/")
	apiGroup.Use(middleware.RateLimit(limiter))
	apiGroup.Use(middleware.Prometheus())

	api := NewHandlers(service, publisher)
	apiGroup.POST("/shorten", api.Shorten)
	apiGroup.POST("/shorten/batch", api.ShortenBatch)
	apiGroup.GET("/:code", api.Redirect)
}
