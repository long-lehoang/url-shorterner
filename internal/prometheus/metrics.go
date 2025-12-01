// Package prometheus provides Prometheus metrics definitions for monitoring.
package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTPRequestsTotal counts the total number of HTTP requests.
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	// RedirectLatency measures the latency of redirect operations.
	RedirectLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "redirect_latency_seconds",
			Help:    "Redirect latency in seconds",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 10),
		},
		[]string{"cache_hit"},
	)

	// CacheHitRatio tracks the cache hit ratio for different cache types.
	CacheHitRatio = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cache_hit_ratio",
			Help: "Cache hit ratio",
		},
		[]string{"type"},
	)

	// RateLimitBlockedTotal counts the total number of requests blocked by rate limiting.
	RateLimitBlockedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_blocked_total",
			Help: "Total number of requests blocked by rate limiter",
		},
		[]string{"identifier"},
	)
)

