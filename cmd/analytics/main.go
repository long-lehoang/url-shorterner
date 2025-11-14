package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"url-shorterner/internal/cache"
	"url-shorterner/internal/config"
	"url-shorterner/internal/storage"
	"url-shorterner/svc/analytics"
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

	dao := storage.NewDAO(pool)
	repo := storage.NewRepository(dao)
	analyticsService := analytics.NewService(repo)

	log.Println("Analytics service starting...")

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			log.Println("Processing analytics...")
			_ = analyticsService
			_ = redisCache
		case <-quit:
			log.Println("Shutting down analytics service...")
			return
		}
	}
}

