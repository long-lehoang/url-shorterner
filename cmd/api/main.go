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
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "url-shorterner/docs"
)

// Package main provides the URL Shortener API server.
//
// A high-performance URL shortener service with analytics and rate limiting capabilities.
//
// This API provides endpoints for creating shortened URLs, retrieving analytics,
// and managing URL redirections with advanced features like expiration, custom aliases,
// and comprehensive click tracking.
//
//     Schemes: http, https
//     Host: localhost:8080
//     BasePath: /
//     Version: 1.0
//     Title: URL Shortener API
//     Description: A high-performance URL shortener service with analytics and rate limiting.
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
//     SecurityDefinitions:
//       ApiKeyAuth:
//         type: apiKey
//         in: header
//         name: Authorization
//         description: API key authentication (optional for public endpoints)
//
//     Contact:
//       name: API Support
//       url: http://www.example.com/support
//       email: support@example.com
//
//     License:
//       name: Apache 2.0
//       url: http://www.apache.org/licenses/LICENSE-2.0.html
//
//     TermsOfService: http://swagger.io/terms/
//
// swagger:meta

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

	storageRepo := storage.NewRepository(writerPool)
	storageDAO := storage.NewDAO(readerPool)

	shortenerRepo := shortenerStore.NewRepository(storageRepo)
	shortenerDAO := shortenerStore.NewDAO(storageDAO)
	shortenerService := shortenerApp.NewService(
		shortenerRepo,
		shortenerDAO,
		urlCache,
		cfg.BloomN,
		cfg.BloomP,
		cfg.ShortCodeLength,
		cfg.Domain,
	)

	analyticsRepo := analyticsStore.NewRepository(storageRepo)
	analyticsDAO := analyticsStore.NewDAO(storageDAO)
	analyticsService := analyticsApp.NewService(analyticsRepo, analyticsDAO)

	limiter := rate.NewLimiter(rateLimitCache, cfg.RateLimitMax, cfg.RateLimitWindow)

	var eventPublisher events.Publisher
	router := gin.Default()

	shortenerTransport.SetupRouter(router, shortenerService, eventPublisher, limiter)
	analyticsTransport.SetupRouter(router, analyticsService, limiter)

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("http://localhost:8080/swagger/doc.json")))

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
