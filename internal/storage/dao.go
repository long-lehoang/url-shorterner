// Package storage provides database access objects for read operations.
package storage

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type URL struct {
	ID          string
	ShortCode   string
	OriginalURL string
	ExpiresAt   *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type AnalyticsRecord struct {
	ID        string
	ShortCode string
	IPAddress string
	UserAgent string
	Referer   string
	ClickedAt time.Time
}

type AnalyticsStats struct {
	TotalClicks int
	UniqueIPs   int
	LastClick   *time.Time
}

type DAO interface {
	GetURLByShortCode(ctx context.Context, shortCode string) (*URL, error)
	CheckShortCodeExists(ctx context.Context, shortCode string) (bool, error)
	GetAnalyticsByShortCode(ctx context.Context, shortCode string, limit int) ([]*AnalyticsRecord, error)
	GetAnalyticsStats(ctx context.Context, shortCode string) (*AnalyticsStats, error)
}

type dao struct {
	db *pgxpool.Pool
}

func NewDAO(db *pgxpool.Pool) DAO {
	return &dao{db: db}
}

func (d *dao) GetURLByShortCode(ctx context.Context, shortCode string) (*URL, error) {
	query := `
		SELECT id, short_code, original_url, expires_at, created_at, updated_at
		FROM urls
		WHERE short_code = @short_code
	`
	args := pgx.NamedArgs{
		"short_code": shortCode,
	}

	var url URL
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

func (d *dao) GetAnalyticsByShortCode(ctx context.Context, shortCode string, limit int) ([]*AnalyticsRecord, error) {
	query := `
		SELECT id, short_code, ip_address, user_agent, referer, clicked_at
		FROM analytics
		WHERE short_code = @short_code
		ORDER BY clicked_at DESC
		LIMIT @limit
	`
	args := pgx.NamedArgs{
		"short_code": shortCode,
		"limit":      limit,
	}

	rows, err := d.db.Query(ctx, query, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]*AnalyticsRecord, 0, limit)
	for rows.Next() {
		var record AnalyticsRecord
		err := rows.Scan(
			&record.ID,
			&record.ShortCode,
			&record.IPAddress,
			&record.UserAgent,
			&record.Referer,
			&record.ClickedAt,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, &record)
	}

	return records, rows.Err()
}

func (d *dao) GetAnalyticsStats(ctx context.Context, shortCode string) (*AnalyticsStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_clicks,
			COUNT(DISTINCT ip_address) as unique_ips,
			MAX(clicked_at) as last_click
		FROM analytics
		WHERE short_code = @short_code
	`
	args := pgx.NamedArgs{
		"short_code": shortCode,
	}

	var stats AnalyticsStats
	var lastClick *time.Time
	err := d.db.QueryRow(ctx, query, args).Scan(
		&stats.TotalClicks,
		&stats.UniqueIPs,
		&lastClick,
	)
	if err != nil {
		return nil, err
	}

	stats.LastClick = lastClick
	return &stats, nil
}
