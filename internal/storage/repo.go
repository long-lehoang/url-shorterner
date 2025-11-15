// Package storage provides repository interfaces for write operations.
package storage

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	CreateURL(ctx context.Context, url *URL) error
	CreateAnalytics(ctx context.Context, record *AnalyticsRecord) error
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) CreateURL(ctx context.Context, url *URL) error {
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

func (r *repository) CreateAnalytics(ctx context.Context, record *AnalyticsRecord) error {
	query := `
		INSERT INTO analytics (id, short_code, ip_address, user_agent, referer, clicked_at)
		VALUES (@id, @short_code, @ip_address, @user_agent, @referer, @clicked_at)
	`
	args := pgx.NamedArgs{
		"id":         record.ID,
		"short_code": record.ShortCode,
		"ip_address": record.IPAddress,
		"user_agent": record.UserAgent,
		"referer":    record.Referer,
		"clicked_at": record.ClickedAt,
	}
	_, err := r.db.Exec(ctx, query, args)
	return err
}
