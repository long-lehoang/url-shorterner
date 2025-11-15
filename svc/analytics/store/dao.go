// Package store provides DAO adapters for the analytics domain.
package store

import (
	"context"
	"url-shorterner/internal/storage"
	"url-shorterner/svc/analytics/entity"
)

type DAO interface {
	GetAnalyticsByShortCode(ctx context.Context, shortCode string, limit int) ([]*entity.Record, error)
	GetAnalyticsStats(ctx context.Context, shortCode string) (*entity.Stats, error)
}

type dao struct {
	storageDAO storage.DAO
}

func NewDAO(storageDAO storage.DAO) DAO {
	return &dao{storageDAO: storageDAO}
}

func (d *dao) GetAnalyticsByShortCode(ctx context.Context, shortCode string, limit int) ([]*entity.Record, error) {
	storageRecords, err := d.storageDAO.GetAnalyticsByShortCode(ctx, shortCode, limit)
	if err != nil {
		return nil, err
	}
	records := make([]*entity.Record, 0, len(storageRecords))
	for _, sr := range storageRecords {
		records = append(records, &entity.Record{
			ID:        sr.ID,
			ShortCode: sr.ShortCode,
			IPAddress: sr.IPAddress,
			UserAgent: sr.UserAgent,
			Referer:   sr.Referer,
			ClickedAt: sr.ClickedAt,
		})
	}
	return records, nil
}

func (d *dao) GetAnalyticsStats(ctx context.Context, shortCode string) (*entity.Stats, error) {
	storageStats, err := d.storageDAO.GetAnalyticsStats(ctx, shortCode)
	if err != nil {
		return nil, err
	}
	return &entity.Stats{
		TotalClicks: storageStats.TotalClicks,
		UniqueIPs:   storageStats.UniqueIPs,
		LastClick:   storageStats.LastClick,
	}, nil
}
