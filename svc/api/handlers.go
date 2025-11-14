package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"url-shorterner/internal/prometheus"
	"url-shorterner/svc/analytics"
	"url-shorterner/svc/shortener"
)

type Handlers struct {
	shortenerService shortener.Service
	analyticsService analytics.Service
}

func NewHandlers(shortenerService shortener.Service, analyticsService analytics.Service) *Handlers {
	return &Handlers{
		shortenerService: shortenerService,
		analyticsService: analyticsService,
	}
}

type ShortenRequest struct {
	URL       string  `json:"url" binding:"required"`
	ExpiresIn *int    `json:"expires_in,omitempty"`
	Alias     *string `json:"alias,omitempty"`
}

type BatchShortenRequest struct {
	Items []shortener.BatchItem `json:"items" binding:"required"`
}

func (h *Handlers) Shorten(c *gin.Context) {
	var req ShortenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		prometheus.HTTPRequestsTotal.WithLabelValues("POST", "/shorten", "400").Inc()
		return
	}

	resp, err := h.shortenerService.Shorten(c.Request.Context(), req.URL, req.ExpiresIn, req.Alias)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "alias already exists" {
			status = http.StatusConflict
		} else if err.Error() == "invalid URL: invalid URL format" || err.Error() == "invalid URL: URL must use http or https scheme" {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"error": err.Error()})
		prometheus.HTTPRequestsTotal.WithLabelValues("POST", "/shorten", fmt.Sprintf("%d", status)).Inc()
		return
	}

	c.JSON(http.StatusOK, resp)
	prometheus.HTTPRequestsTotal.WithLabelValues("POST", "/shorten", "200").Inc()
}

func (h *Handlers) ShortenBatch(c *gin.Context) {
	var req BatchShortenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		prometheus.HTTPRequestsTotal.WithLabelValues("POST", "/shorten/batch", "400").Inc()
		return
	}

	results, err := h.shortenerService.ShortenBatch(c.Request.Context(), req.Items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		prometheus.HTTPRequestsTotal.WithLabelValues("POST", "/shorten/batch", "500").Inc()
		return
	}

	c.JSON(http.StatusOK, gin.H{"results": results})
	prometheus.HTTPRequestsTotal.WithLabelValues("POST", "/shorten/batch", "200").Inc()
}

func (h *Handlers) Redirect(c *gin.Context) {
	shortCode := c.Param("code")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "short code is required"})
		prometheus.HTTPRequestsTotal.WithLabelValues("GET", "/:code", "400").Inc()
		return
	}

	start := time.Now()
	originalURL, err := h.shortenerService.GetOriginalURL(c.Request.Context(), shortCode)
	latency := time.Since(start).Seconds()

	cacheHit := "true"
	if err != nil {
		cacheHit = "false"
		if err.Error() == "url not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
			prometheus.HTTPRequestsTotal.WithLabelValues("GET", "/:code", "404").Inc()
			prometheus.RedirectLatency.WithLabelValues(cacheHit).Observe(latency)
			return
		} else if err.Error() == "url expired" {
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
		ipAddress := c.ClientIP()
		userAgent := c.GetHeader("User-Agent")
		referer := c.GetHeader("Referer")
		_ = h.analyticsService.RecordClick(c.Request.Context(), shortCode, ipAddress, userAgent, referer)
	}()

	c.Redirect(http.StatusMovedPermanently, originalURL)
	prometheus.HTTPRequestsTotal.WithLabelValues("GET", "/:code", "301").Inc()
	prometheus.RedirectLatency.WithLabelValues(cacheHit).Observe(latency)
}

func (h *Handlers) GetAnalytics(c *gin.Context) {
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

	stats, err := h.analyticsService.GetStats(c.Request.Context(), shortCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		prometheus.HTTPRequestsTotal.WithLabelValues("GET", "/analytics/:code", "500").Inc()
		return
	}

	records, err := h.analyticsService.GetAnalytics(c.Request.Context(), shortCode, limit)
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
	var result int
	for _, char := range s {
		if char >= '0' && char <= '9' {
			result = result*10 + int(char-'0')
		} else {
			return 0
		}
	}
	return result
}

