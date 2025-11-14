package storage

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Repository interface {
	CreateURL(ctx context.Context, shortCode, originalURL string, expiresAt *time.Time) (*URL, error)
	GetURLByShortCode(ctx context.Context, shortCode string) (*URL, error)
	CheckShortCodeExists(ctx context.Context, shortCode string) (bool, error)
	RecordClick(ctx context.Context, shortCode, ipAddress, userAgent, referer string) error
	GetAnalytics(ctx context.Context, shortCode string, limit int) ([]*AnalyticsRecord, error)
	GetAnalyticsStats(ctx context.Context, shortCode string) (*AnalyticsStats, error)
}

type repository struct {
	dao DAO
}

func NewRepository(dao DAO) Repository {
	return &repository{dao: dao}
}

func (r *repository) CreateURL(ctx context.Context, shortCode, originalURL string, expiresAt *time.Time) (*URL, error) {
	now := time.Now().UTC()
	url := &URL{
		ID:          uuid.New().String(),
		ShortCode:   shortCode,
		OriginalURL: originalURL,
		ExpiresAt:   expiresAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := r.dao.CreateURL(ctx, url); err != nil {
		return nil, err
	}

	return url, nil
}

func (r *repository) GetURLByShortCode(ctx context.Context, shortCode string) (*URL, error) {
	url, err := r.dao.GetURLByShortCode(ctx, shortCode)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if url.ExpiresAt != nil && url.ExpiresAt.Before(time.Now().UTC()) {
		return nil, ErrExpired
	}

	return url, nil
}

func (r *repository) CheckShortCodeExists(ctx context.Context, shortCode string) (bool, error) {
	return r.dao.CheckShortCodeExists(ctx, shortCode)
}

func (r *repository) RecordClick(ctx context.Context, shortCode, ipAddress, userAgent, referer string) error {
	record := &AnalyticsRecord{
		ID:        uuid.New().String(),
		ShortCode: shortCode,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Referer:   referer,
		ClickedAt: time.Now().UTC(),
	}

	return r.dao.CreateAnalytics(ctx, record)
}

func (r *repository) GetAnalytics(ctx context.Context, shortCode string, limit int) ([]*AnalyticsRecord, error) {
	return r.dao.GetAnalyticsByShortCode(ctx, shortCode, limit)
}

func (r *repository) GetAnalyticsStats(ctx context.Context, shortCode string) (*AnalyticsStats, error) {
	return r.dao.GetAnalyticsStats(ctx, shortCode)
}

