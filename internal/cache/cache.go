// Package cache provides Redis-based caching functionality for URLs and rate limiting.
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}

type cache struct {
	client *redis.Client
}

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

type URLCache struct {
	cache Cache
}

func NewURLCache(c Cache) *URLCache {
	return &URLCache{cache: c}
}

func (uc *URLCache) GetURL(ctx context.Context, shortCode string) (string, error) {
	key := fmt.Sprintf("url:%s", shortCode)
	return uc.cache.Get(ctx, key)
}

func (uc *URLCache) SetURL(ctx context.Context, shortCode, originalURL string, ttl time.Duration) error {
	key := fmt.Sprintf("url:%s", shortCode)
	return uc.cache.Set(ctx, key, originalURL, ttl)
}

func (uc *URLCache) DeleteURL(ctx context.Context, shortCode string) error {
	key := fmt.Sprintf("url:%s", shortCode)
	return uc.cache.Delete(ctx, key)
}

type RateLimitCache struct {
	cache Cache
}

func NewRateLimitCache(c Cache) *RateLimitCache {
	return &RateLimitCache{cache: c}
}

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

func (rlc *RateLimitCache) SetWindow(ctx context.Context, key string, timestamps []string, ttl time.Duration) error {
	data, err := json.Marshal(timestamps)
	if err != nil {
		return err
	}
	return rlc.cache.Set(ctx, key, string(data), ttl)
}

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

