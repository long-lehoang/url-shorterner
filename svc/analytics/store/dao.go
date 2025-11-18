// Package store provides DAO implementations for the analytics domain.
package store

import (
	"context"
	"time"

	"url-shorterner/svc/analytics/entity"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DAO interface {
	GetAnalyticsByShortCode(ctx context.Context, shortCode string, limit int) ([]*entity.Record, error)
	GetAnalyticsStats(ctx context.Context, shortCode string) (*entity.Stats, error)
}

type dao struct {
	db *pgxpool.Pool
}

func NewDAO(db *pgxpool.Pool) DAO {
	return &dao{db: db}
}

func (d *dao) GetAnalyticsByShortCode(ctx context.Context, shortCode string, limit int) ([]*entity.Record, error) {
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

	records := make([]*entity.Record, 0, limit)
	for rows.Next() {
		var record entity.Record
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

func (d *dao) GetAnalyticsStats(ctx context.Context, shortCode string) (*entity.Stats, error) {
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

	var stats entity.Stats
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
