// Package store provides repository implementations for the analytics domain.
package store

import (
	"context"

	"url-shorterner/svc/analytics/entity"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	CreateAnalytics(ctx context.Context, record *entity.Record) error
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) CreateAnalytics(ctx context.Context, record *entity.Record) error {
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
