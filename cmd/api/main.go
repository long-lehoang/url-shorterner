package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"url-shorterner/internal/cache"
	"url-shorterner/internal/config"
	"url-shorterner/internal/rate"
	"url-shorterner/internal/storage"
	"url-shorterner/svc/analytics"
	"url-shorterner/svc/api"
	"url-shorterner/svc/shortener"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx := context.Background()

	pool, err := storage.NewDBPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	redisCache, err := cache.NewCache(cfg.RedisAddr, cfg.RedisPassword)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	urlCache := cache.NewURLCache(redisCache)
	rateLimitCache := cache.NewRateLimitCache(redisCache)

	dao := storage.NewDAO(pool)
	repo := storage.NewRepository(dao)

	shortenerService := shortener.NewService(
		repo,
		urlCache,
		cfg.BloomN,
		cfg.BloomP,
		cfg.ShortCodeLength,
		cfg.Domain,
	)

	analyticsService := analytics.NewService(repo)

	limiter := rate.NewLimiter(rateLimitCache, cfg.RateLimitMax, cfg.RateLimitWindow)

	router := api.SetupRouter(shortenerService, analyticsService, limiter)

	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("API server starting on port %d", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down API server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("API server exited")
}

