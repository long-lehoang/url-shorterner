package rate

import (
	"context"
	"fmt"
	"time"

	"url-shorterner/internal/cache"
)

type Limiter interface {
	Allow(ctx context.Context, identifier string) (bool, error)
}

type limiter struct {
	rateLimitCache *cache.RateLimitCache
	maxRequests    int
	windowSize     time.Duration
}

func NewLimiter(rateLimitCache *cache.RateLimitCache, maxRequests int, windowSize time.Duration) Limiter {
	return &limiter{
		rateLimitCache: rateLimitCache,
		maxRequests:    maxRequests,
		windowSize:     windowSize,
	}
}

func (l *limiter) Allow(ctx context.Context, identifier string) (bool, error) {
	key := fmt.Sprintf("ratelimit:%s", identifier)
	now := time.Now()
	timestamp := now.Format(time.RFC3339)

	timestamps, err := l.rateLimitCache.GetWindow(ctx, key)
	if err != nil && err != cache.ErrNotFound {
		return false, err
	}

	if timestamps == nil {
		timestamps = make([]string, 0, l.maxRequests)
	}

	cutoff := now.Add(-l.windowSize)
	validCount := 0
	for _, ts := range timestamps {
		t, err := time.Parse(time.RFC3339, ts)
		if err == nil && t.After(cutoff) {
			validCount++
		}
	}

	if validCount >= l.maxRequests {
		return false, nil
	}

	if err := l.rateLimitCache.AddToWindow(ctx, key, timestamp, l.windowSize); err != nil {
		return false, err
	}

	return true, nil
}

