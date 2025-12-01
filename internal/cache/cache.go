// Package cache provides Redis-based caching functionality for URLs and rate limiting.
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache defines the interface for cache operations.
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}

type cache struct {
	client *redis.Client
}

// NewCache creates a new Redis cache instance.
func NewCache(addr, password string) (Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &cache{client: client}, nil
}

func (c *cache) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	return val, nil
}

func (c *cache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *cache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func (c *cache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// URLCache provides URL-specific caching operations.
type URLCache struct {
	cache Cache
}

// NewURLCache creates a new URL cache instance.
func NewURLCache(c Cache) *URLCache {
	return &URLCache{cache: c}
}

// GetURL retrieves the original URL for a given short code.
func (uc *URLCache) GetURL(ctx context.Context, shortCode string) (string, error) {
	key := fmt.Sprintf("url:%s", shortCode)
	return uc.cache.Get(ctx, key)
}

// SetURL stores the original URL for a given short code with TTL.
func (uc *URLCache) SetURL(ctx context.Context, shortCode, originalURL string, ttl time.Duration) error {
	key := fmt.Sprintf("url:%s", shortCode)
	return uc.cache.Set(ctx, key, originalURL, ttl)
}

// DeleteURL removes a URL from cache.
func (uc *URLCache) DeleteURL(ctx context.Context, shortCode string) error {
	key := fmt.Sprintf("url:%s", shortCode)
	return uc.cache.Delete(ctx, key)
}

// RateLimitCache provides rate limiting window caching operations.
type RateLimitCache struct {
	cache Cache
}

// NewRateLimitCache creates a new rate limit cache instance.
func NewRateLimitCache(c Cache) *RateLimitCache {
	return &RateLimitCache{cache: c}
}

// GetWindow retrieves the timestamps for a rate limit window.
func (rlc *RateLimitCache) GetWindow(ctx context.Context, key string) ([]string, error) {
	val, err := rlc.cache.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var timestamps []string
	if err := json.Unmarshal([]byte(val), &timestamps); err != nil {
		return nil, err
	}
	return timestamps, nil
}

// SetWindow stores the timestamps for a rate limit window with TTL.
func (rlc *RateLimitCache) SetWindow(ctx context.Context, key string, timestamps []string, ttl time.Duration) error {
	data, err := json.Marshal(timestamps)
	if err != nil {
		return err
	}
	return rlc.cache.Set(ctx, key, string(data), ttl)
}

// AddToWindow adds a timestamp to the rate limit window and filters old entries.
func (rlc *RateLimitCache) AddToWindow(ctx context.Context, key string, timestamp string, windowSize time.Duration) error {
	timestamps, _ := rlc.GetWindow(ctx, key)
	if timestamps == nil {
		timestamps = make([]string, 0, 100)
	}

	timestamps = append(timestamps, timestamp)
	now := time.Now()

	filtered := make([]string, 0, len(timestamps))
	cutoff := now.Add(-windowSize)
	for _, ts := range timestamps {
		t, err := time.Parse(time.RFC3339, ts)
		if err == nil && t.After(cutoff) {
			filtered = append(filtered, ts)
		}
	}

	return rlc.SetWindow(ctx, key, filtered, windowSize+time.Second*10)
}

var (
	// ErrNotFound is returned when a cache key is not found.
	ErrNotFound = fmt.Errorf("not found")
)

