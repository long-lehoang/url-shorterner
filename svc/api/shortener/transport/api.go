// Package transport provides HTTP handler implementations for the shortener API.
package transport

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	eventsPublisher "url-shorterner/internal/events"
	"url-shorterner/internal/prometheus"
	analyticsEvents "url-shorterner/svc/analytics/events"
	"url-shorterner/svc/shortener/app"

	"github.com/gin-gonic/gin"
)

type handlers struct {
	service   app.Service
	publisher eventsPublisher.Publisher
}

func NewHandlers(service app.Service, publisher eventsPublisher.Publisher) ShortenerAPI {
	return &handlers{
		service:   service,
		publisher: publisher,
	}
}

// Shorten implements ShortenerAPI.Shorten
// See ShortenerAPI interface in http.go for API documentation
func (h *handlers) Shorten(c *gin.Context) {
	var req ShortenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		prometheus.HTTPRequestsTotal.WithLabelValues("POST", "/shorten", "400").Inc()
		return
	}

	resp, err := h.service.Shorten(c.Request.Context(), req.URL, req.ExpiresIn, req.Alias)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, app.ErrAliasExists) {
			status = http.StatusConflict
		} else if errors.Is(err, app.ErrInvalidURL) || errors.Is(err, app.ErrInvalidURLFormat) || errors.Is(err, app.ErrInvalidURLScheme) {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"error": err.Error()})
		prometheus.HTTPRequestsTotal.WithLabelValues("POST", "/shorten", fmt.Sprintf("%d", status)).Inc()
		return
	}

	c.JSON(http.StatusOK, resp)
	prometheus.HTTPRequestsTotal.WithLabelValues("POST", "/shorten", "200").Inc()
}

// ShortenBatch implements ShortenerAPI.ShortenBatch
// See ShortenerAPI interface in http.go for API documentation
func (h *handlers) ShortenBatch(c *gin.Context) {
	var req BatchShortenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		prometheus.HTTPRequestsTotal.WithLabelValues("POST", "/shorten/batch", "400").Inc()
		return
	}

	results, err := h.service.ShortenBatch(c.Request.Context(), req.Items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		prometheus.HTTPRequestsTotal.WithLabelValues("POST", "/shorten/batch", "500").Inc()
		return
	}

	c.JSON(http.StatusOK, gin.H{"results": results})
	prometheus.HTTPRequestsTotal.WithLabelValues("POST", "/shorten/batch", "200").Inc()
}

// Redirect implements ShortenerAPI.Redirect
// See ShortenerAPI interface in http.go for API documentation
func (h *handlers) Redirect(c *gin.Context) {
	shortCode := c.Param("code")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "short code is required"})
		prometheus.HTTPRequestsTotal.WithLabelValues("GET", "/:code", "400").Inc()
		return
	}

	start := time.Now()
	originalURL, err := h.service.GetOriginalURL(c.Request.Context(), shortCode)
	latency := time.Since(start).Seconds()

	cacheHit := "true"
	if err != nil {
		cacheHit = "false"
		if errors.Is(err, app.ErrURLNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
			prometheus.HTTPRequestsTotal.WithLabelValues("GET", "/:code", "404").Inc()
			prometheus.RedirectLatency.WithLabelValues(cacheHit).Observe(latency)
			return
		} else if errors.Is(err, app.ErrURLExpired) {
			c.JSON(http.StatusGone, gin.H{"error": "URL expired"})
			prometheus.HTTPRequestsTotal.WithLabelValues("GET", "/:code", "410").Inc()
			prometheus.RedirectLatency.WithLabelValues(cacheHit).Observe(latency)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		prometheus.HTTPRequestsTotal.WithLabelValues("GET", "/:code", "500").Inc()
		prometheus.RedirectLatency.WithLabelValues(cacheHit).Observe(latency)
		return
	}

	go func() {
		ctx := context.Background()
		clickEvent := analyticsEvents.ClickEvent{
			ShortCode: shortCode,
			IPAddress: c.ClientIP(),
			UserAgent: c.GetHeader("User-Agent"),
			Referer:   c.GetHeader("Referer"),
			Timestamp: time.Now().UTC(),
		}
		if err := h.publisher.PublishClickEvent(ctx, clickEvent); err != nil {
			// Log error but don't fail the redirect
			// In production, consider using a proper logger
			_ = err
		}
	}()

	c.Redirect(http.StatusMovedPermanently, originalURL)
	prometheus.HTTPRequestsTotal.WithLabelValues("GET", "/:code", "301").Inc()
	prometheus.RedirectLatency.WithLabelValues(cacheHit).Observe(latency)
}
