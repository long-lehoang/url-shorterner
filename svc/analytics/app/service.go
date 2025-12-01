// Package app provides the core business logic for analytics operations.
package app

import (
	"context"
	"time"

	"url-shorterner/internal/uuid"
	"url-shorterner/svc/analytics/entity"
	analyticsStore "url-shorterner/svc/analytics/store"
)

// Service defines the interface for analytics operations.
type Service interface {
	RecordClick(ctx context.Context, shortCode, ipAddress, userAgent, referer string) error
	GetAnalytics(ctx context.Context, shortCode string, limit int) ([]*entity.Record, error)
	GetStats(ctx context.Context, shortCode string) (*entity.Stats, error)
}

type service struct {
	repo analyticsStore.Repository
	dao  analyticsStore.DAO
}

// NewService creates a new analytics service instance.
func NewService(repo analyticsStore.Repository, dao analyticsStore.DAO) Service {
	return &service{
		repo: repo,
		dao:  dao,
	}
}

func (s *service) RecordClick(ctx context.Context, shortCode, ipAddress, userAgent, referer string) error {
	record := &entity.Record{
		ID:        uuid.Generate(),
		ShortCode: shortCode,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Referer:   referer,
		ClickedAt: time.Now().UTC(),
	}
	return s.repo.CreateAnalytics(ctx, record)
}

func (s *service) GetAnalytics(ctx context.Context, shortCode string, limit int) ([]*entity.Record, error) {
	return s.dao.GetAnalyticsByShortCode(ctx, shortCode, limit)
}

func (s *service) GetStats(ctx context.Context, shortCode string) (*entity.Stats, error) {
	return s.dao.GetAnalyticsStats(ctx, shortCode)
}
