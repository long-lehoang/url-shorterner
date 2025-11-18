// Package integration provides integration tests for the API.
package integration

import (
	"context"
	"fmt"
	"os"
	"time"

	"url-shorterner/internal/cache"
	"url-shorterner/internal/config"
	"url-shorterner/internal/events"
	"url-shorterner/internal/middleware"
	"url-shorterner/internal/rate"
	"url-shorterner/internal/storage"
	analyticsApp "url-shorterner/svc/analytics/app"
	analyticsStore "url-shorterner/svc/analytics/store"
	analyticsTransport "url-shorterner/svc/api/analytics/transport"
	shortenerTransport "url-shorterner/svc/api/shortener/transport"
	shortenerApp "url-shorterner/svc/shortener/app"
	shortenerStore "url-shorterner/svc/shortener/store"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// SetupTestConfig creates a test configuration with default values.
// Environment variables can override defaults:
//   - TEST_DATABASE_URL: PostgreSQL connection string
//   - TEST_REDIS_ADDR: Redis address
func SetupTestConfig() (*config.Config, error) {
	// Use test database and Redis
	// Default to regular database if TEST_DATABASE_URL is not set
	databaseURL := getEnv("TEST_DATABASE_URL", "postgres://postgres:password@localhost:5432/shortener?sslmode=disable")
	redisAddr := getEnv("TEST_REDIS_ADDR", "localhost:6379")

	cfg := &config.Config{
		Port:              8080,
		DatabaseURL:       databaseURL,
		DatabaseReaderURL: databaseURL,
		RedisAddr:         redisAddr,
		RedisPassword:     "",
		ShortCodeLength:   8,
		RateLimitMax:      1000, // High limit for tests
		RateLimitWindow:   60 * time.Second,
		BloomN:            1000000,
		BloomP:            0.001,
		Domain:            "http://localhost:8080",
	}

	return cfg, nil
}

// SetupTestRouter creates a test router with all dependencies initialized.
// It sets up database connections, Redis cache, services, and registers all routes.
func SetupTestRouter(cfg *config.Config) *gin.Engine {
	ctx := context.Background()

	writerPool, err := storage.NewDBPool(ctx, cfg.DatabaseURL)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to writer database: %v", err))
	}

	readerPool, err := storage.NewDBPool(ctx, cfg.DatabaseReaderURL)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to reader database: %v", err))
	}

	redisCache, err := cache.NewCache(cfg.RedisAddr, cfg.RedisPassword)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to Redis: %v", err))
	}

	urlCache := cache.NewURLCache(redisCache)
	rateLimitCache := cache.NewRateLimitCache(redisCache)

	shortenerRepo := shortenerStore.NewRepository(writerPool)
	shortenerDAO := shortenerStore.NewDAO(readerPool)
	var eventPublisher events.Publisher

	shortenerService := shortenerApp.NewService(
		shortenerRepo,
		shortenerDAO,
		urlCache,
		cfg.BloomN,
		cfg.BloomP,
		cfg.ShortCodeLength,
		cfg.Domain,
		eventPublisher,
	)

	analyticsRepo := analyticsStore.NewRepository(writerPool)
	analyticsDAO := analyticsStore.NewDAO(readerPool)
	analyticsService := analyticsApp.NewService(analyticsRepo, analyticsDAO)

	limiter := rate.NewLimiter(rateLimitCache, cfg.RateLimitMax, cfg.RateLimitWindow)

	router := gin.New()
	router.Use(middleware.Recovery())
	router.Use(middleware.Logger())

	shortenerTransport.SetupRouter(router, shortenerService, limiter)
	analyticsTransport.SetupRouter(router, analyticsService, limiter)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	return router
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
