package analytics

import (
	"context"
	"url-shorterner/internal/storage"
)

type Service interface {
	RecordClick(ctx context.Context, shortCode, ipAddress, userAgent, referer string) error
	GetAnalytics(ctx context.Context, shortCode string, limit int) ([]*storage.AnalyticsRecord, error)
	GetStats(ctx context.Context, shortCode string) (*storage.AnalyticsStats, error)
}

type service struct {
	repo storage.Repository
}

func NewService(repo storage.Repository) Service {
	return &service{repo: repo}
}

func (s *service) RecordClick(ctx context.Context, shortCode, ipAddress, userAgent, referer string) error {
	return s.repo.RecordClick(ctx, shortCode, ipAddress, userAgent, referer)
}

func (s *service) GetAnalytics(ctx context.Context, shortCode string, limit int) ([]*storage.AnalyticsRecord, error) {
	return s.repo.GetAnalytics(ctx, shortCode, limit)
}

func (s *service) GetStats(ctx context.Context, shortCode string) (*storage.AnalyticsStats, error) {
	return s.repo.GetAnalyticsStats(ctx, shortCode)
}

