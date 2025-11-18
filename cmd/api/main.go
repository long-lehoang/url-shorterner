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

	redisCache, err := cache.NewCache(cfg.RedisAddr, cfg.RedisPassword)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	urlCache := cache.NewURLCache(redisCache)
	rateLimitCache := cache.NewRateLimitCache(redisCache)

	shortenerRepo := shortenerStore.NewRepository(writerPool)
	shortenerDAO := shortenerStore.NewDAO(readerPool)
	var eventPublisher events.Publisher
	// TODO: Initialize event publisher implementation when available

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

	shortenerTransport.SetupRouter(router, shortenerService, limiter)
	analyticsTransport.SetupRouter(router, analyticsService, limiter)

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(os.Getenv("GIN_MODE"))
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
