// Package config provides configuration loading from environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds application configuration loaded from environment variables.
type Config struct {
	Port              int
	DatabaseURL       string
	DatabaseReaderURL string
	RedisAddr         string
	RedisPassword     string
	ShortCodeLength   int
	RateLimitMax      int
	RateLimitWindow   time.Duration
	BloomN            uint
	BloomP            float64
	Domain            string
}

// Load reads configuration from environment variables and returns a Config instance.
func Load() (*Config, error) {
	databaseURL := getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/shortener?sslmode=disable")

	databaseReaderURL := getEnv("DATABASE_READER_URL", "")
	if databaseReaderURL == "" {
		databaseReaderURL = databaseURL
	}

	cfg := &Config{
		Port:              getEnvInt("PORT", 8080),
		DatabaseURL:       databaseURL,
		DatabaseReaderURL: databaseReaderURL,
		RedisAddr:         getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:     getEnv("REDIS_PASSWORD", ""),
		ShortCodeLength:   getEnvInt("SHORT_CODE_LENGTH", 8),
		RateLimitMax:      getEnvInt("RATE_LIMIT_MAX", 100),
		RateLimitWindow:   time.Duration(getEnvInt("RATE_LIMIT_WINDOW_SECONDS", 60)) * time.Second,
		BloomN:            uint(getEnvInt("BLOOM_N", 1000000)), //nolint:gosec // G115: Bloom filter size is configurable and validated
		BloomP:            getEnvFloat("BLOOM_P", 0.001),
		Domain:            getEnv("DOMAIN", "http://localhost:8080"),
	}

	if cfg.ShortCodeLength < 4 || cfg.ShortCodeLength > 20 {
		return nil, fmt.Errorf("SHORT_CODE_LENGTH must be between 4 and 20")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}
