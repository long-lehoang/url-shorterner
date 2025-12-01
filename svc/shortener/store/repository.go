// Package store provides repository implementations for the shortener domain.
package store

import (
	"context"

	"url-shorterner/svc/shortener/entity"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the interface for shortener write operations.
type Repository interface {
	CreateURL(ctx context.Context, url *entity.URL) error
}

type repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new shortener repository instance.
func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) CreateURL(ctx context.Context, url *entity.URL) error {
	query := `
		INSERT INTO urls (id, short_code, original_url, expires_at, created_at, updated_at)
		VALUES (@id, @short_code, @original_url, @expires_at, @created_at, @updated_at)
	`
	args := pgx.NamedArgs{
		"id":           url.ID,
		"short_code":   url.ShortCode,
		"original_url": url.OriginalURL,
		"expires_at":   url.ExpiresAt,
		"created_at":   url.CreatedAt,
		"updated_at":   url.UpdatedAt,
	}
	_, err := r.db.Exec(ctx, query, args)
	return err
}
