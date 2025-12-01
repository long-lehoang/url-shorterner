// Package main provides the entry point for database migration tool.
package main

import (
	"context"
	"log"
	"os"

	"url-shorterner/internal/config"
	"url-shorterner/internal/storage"
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

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err) //nolint:gocritic // exitAfterDefer: intentional exit on fatal error
	}

	migrationsPath := wd + "/migrations"
	if envPath := os.Getenv("MIGRATIONS_PATH"); envPath != "" {
		migrationsPath = envPath
	}

	log.Printf("Running migrations from: %s", migrationsPath)
	if err := storage.RunMigrations(ctx, pool, migrationsPath); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Migrations completed successfully")
}

