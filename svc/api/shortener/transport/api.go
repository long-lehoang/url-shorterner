// Package transport provides HTTP handler implementations for the shortener API.
package transport

import (
	"errors"
	"net/http"

	"url-shorterner/svc/shortener/app"

	"github.com/gin-gonic/gin"
)

type api struct {
	service app.Service
}

// NewShortenerAPI creates a new shortener API handler instance.
func NewShortenerAPI(service app.Service) ShortenerAPI {
	return &api{
		service: service,
	}
}

// Shorten implements ShortenerAPI.Shorten
// See ShortenerAPI interface in http.go for API documentation
func (a *api) Shorten(c *gin.Context) {
	var req ShortenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err) //nolint:errcheck // Error is handled by ErrorHandler middleware
		return
	}

	resp, err := a.service.Shorten(c.Request.Context(), req.URL, req.ExpiresIn, req.Alias)
	if err != nil {
		c.Error(err) //nolint:errcheck // Error is handled by ErrorHandler middleware
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ShortenBatch implements ShortenerAPI.ShortenBatch
// See ShortenerAPI interface in http.go for API documentation
func (a *api) ShortenBatch(c *gin.Context) {
	var req BatchShortenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err) //nolint:errcheck // Error is handled by ErrorHandler middleware
		return
	}

	results, err := a.service.ShortenBatch(c.Request.Context(), req.Items)
	if err != nil {
		c.Error(err) //nolint:errcheck // Error is handled by ErrorHandler middleware
		return
	}

	c.JSON(http.StatusOK, gin.H{"results": results})
}

// Redirect implements ShortenerAPI.Redirect
// See ShortenerAPI interface in http.go for API documentation
func (a *api) Redirect(c *gin.Context) {
	shortCode := c.Param("code")
	if shortCode == "" {
		c.Error(errors.New("short code is required")) //nolint:errcheck // Error is handled by ErrorHandler middleware
		return
	}

	clickInfo := &app.ClickInfo{
		IPAddress: c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
		Referer:   c.GetHeader("Referer"),
	}

	originalURL, err := a.service.GetOriginalURL(c.Request.Context(), shortCode, clickInfo)
	if err != nil {
		c.Error(err) //nolint:errcheck // Error is handled by ErrorHandler middleware
		return
	}

	c.Redirect(http.StatusMovedPermanently, originalURL)
}
