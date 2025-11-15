// Package store provides repository adapters for the analytics domain.
package store

import (
	"context"
	"url-shorterner/internal/storage"
	"url-shorterner/svc/analytics/entity"
)

type Repository interface {
	CreateAnalytics(ctx context.Context, record *entity.Record) error
}

type repository struct {
	repo storage.Repository
}

func NewRepository(repo storage.Repository) Repository {
	return &repository{repo: repo}
}

func (r *repository) CreateAnalytics(ctx context.Context, record *entity.Record) error {
	storageRecord := &storage.AnalyticsRecord{
		ID:        record.ID,
		ShortCode: record.ShortCode,
		IPAddress: record.IPAddress,
		UserAgent: record.UserAgent,
		Referer:   record.Referer,
		ClickedAt: record.ClickedAt,
	}
	return r.repo.CreateAnalytics(ctx, storageRecord)
}
