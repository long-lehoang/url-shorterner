// Package store provides DAO adapters for the shortener domain.
package store

import (
	"context"
	"url-shorterner/internal/storage"
	"url-shorterner/svc/shortener/entity"
)

type DAO interface {
	GetURLByShortCode(ctx context.Context, shortCode string) (*entity.URL, error)
	CheckShortCodeExists(ctx context.Context, shortCode string) (bool, error)
}

type dao struct {
	storageDAO storage.DAO
}

func NewDAO(storageDAO storage.DAO) DAO {
	return &dao{storageDAO: storageDAO}
}

func (d *dao) GetURLByShortCode(ctx context.Context, shortCode string) (*entity.URL, error) {
	storageURL, err := d.storageDAO.GetURLByShortCode(ctx, shortCode)
	if err != nil {
		return nil, err
	}
	return &entity.URL{
		ID:          storageURL.ID,
		ShortCode:   storageURL.ShortCode,
		OriginalURL: storageURL.OriginalURL,
		ExpiresAt:   storageURL.ExpiresAt,
		CreatedAt:   storageURL.CreatedAt,
		UpdatedAt:   storageURL.UpdatedAt,
	}, nil
}

func (d *dao) CheckShortCodeExists(ctx context.Context, shortCode string) (bool, error) {
	return d.storageDAO.CheckShortCodeExists(ctx, shortCode)
}
