// Package store provides repository adapters for the shortener domain.
package store

import (
	"context"
	"url-shorterner/internal/storage"
	"url-shorterner/svc/shortener/entity"
)

type Repository interface {
	CreateURL(ctx context.Context, url *entity.URL) error
}

type repository struct {
	repo storage.Repository
}

func NewRepository(repo storage.Repository) Repository {
	return &repository{repo: repo}
}

func (r *repository) CreateURL(ctx context.Context, url *entity.URL) error {
	storageURL := &storage.URL{
		ID:          url.ID,
		ShortCode:   url.ShortCode,
		OriginalURL: url.OriginalURL,
		ExpiresAt:   url.ExpiresAt,
		CreatedAt:   url.CreatedAt,
		UpdatedAt:   url.UpdatedAt,
	}
	return r.repo.CreateURL(ctx, storageURL)
}
