package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"url-shorterner/internal/config"
	"url-shorterner/internal/storage"
	analyticsApp "url-shorterner/svc/analytics/app"
	analyticsStore "url-shorterner/svc/analytics/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx := context.Background()

	writerPool, err := storage.NewDBPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to writer database: %v", err)
	}
	defer writerPool.Close()

	readerPool, err := storage.NewDBPool(ctx, cfg.DatabaseReaderURL)
	if err != nil {
		log.Fatalf("Failed to connect to reader database: %v", err)
	}
	defer readerPool.Close()

	storageRepo := storage.NewRepository(writerPool)
	storageDAO := storage.NewDAO(readerPool)

	analyticsRepo := analyticsStore.NewRepository(storageRepo)
	analyticsDAO := analyticsStore.NewDAO(storageDAO)
	_ = analyticsApp.NewService(analyticsRepo, analyticsDAO)

	log.Println("Analytics service starting (event-driven mode)...")

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			log.Println("Analytics service running (event-driven mode)")
		case <-quit:
			log.Println("Shutting down analytics service...")
			return
		}
	}
}
