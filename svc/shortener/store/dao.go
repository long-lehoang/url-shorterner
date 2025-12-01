// Package store provides DAO implementations for the shortener domain.
package store

import (
	"context"
	"errors"
	"time"

	"url-shorterner/internal/storage"
	"url-shorterner/svc/shortener/entity"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DAO defines the data access interface for shortener read operations.
type DAO interface {
	GetURLByShortCode(ctx context.Context, shortCode string) (*entity.URL, error)
	CheckShortCodeExists(ctx context.Context, shortCode string) (bool, error)
}

type dao struct {
	db *pgxpool.Pool
}

// NewDAO creates a new shortener DAO instance.
func NewDAO(db *pgxpool.Pool) DAO {
	return &dao{db: db}
}

func (d *dao) GetURLByShortCode(ctx context.Context, shortCode string) (*entity.URL, error) {
	query := `
		SELECT id, short_code, original_url, expires_at, created_at, updated_at
		FROM urls
		WHERE short_code = @short_code
	`
	args := pgx.NamedArgs{
		"short_code": shortCode,
	}

	var url entity.URL
	var expiresAt *time.Time
	err := d.db.QueryRow(ctx, query, args).Scan(
		&url.ID,
		&url.ShortCode,
		&url.OriginalURL,
		&expiresAt,
		&url.CreatedAt,
		&url.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}

	url.ExpiresAt = expiresAt
	return &url, nil
}

func (d *dao) CheckShortCodeExists(ctx context.Context, shortCode string) (bool, error) {
	query := `
		SELECT EXISTS(SELECT 1 FROM urls WHERE short_code = @short_code)
	`
	args := pgx.NamedArgs{
		"short_code": shortCode,
	}

	var exists bool
	err := d.db.QueryRow(ctx, query, args).Scan(&exists)
	return exists, err
}
