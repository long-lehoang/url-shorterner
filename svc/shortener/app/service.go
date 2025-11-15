package app

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"time"

	"url-shorterner/internal/cache"
	"url-shorterner/internal/storage"
	"url-shorterner/internal/uuid"
	"url-shorterner/svc/shortener/entity"
	shortenerStore "url-shorterner/svc/shortener/store"

	"github.com/bits-and-blooms/bloom/v3"
)

type Service interface {
	Shorten(ctx context.Context, originalURL string, expiresIn *int, alias *string) (*ShortenResponse, error)
	ShortenBatch(ctx context.Context, items []BatchItem) ([]BatchResult, error)
	GetOriginalURL(ctx context.Context, shortCode string) (string, error)
}

type service struct {
	repo         shortenerStore.Repository
	dao          shortenerStore.DAO
	urlCache     *cache.URLCache
	bloomFilter  *bloom.BloomFilter
	shortCodeLen int
	domain       string
}

func NewService(
	repo shortenerStore.Repository,
	dao shortenerStore.DAO,
	urlCache *cache.URLCache,
	bloomN uint,
	bloomP float64,
	shortCodeLen int,
	domain string,
) Service {
	bf := bloom.NewWithEstimates(bloomN, bloomP)
	return &service{
		repo:         repo,
		dao:          dao,
		urlCache:     urlCache,
		bloomFilter:  bf,
		shortCodeLen: shortCodeLen,
		domain:       domain,
	}
}

// ShortenResponse represents the response after successfully shortening a URL
//
// swagger:model ShortenResponse
type ShortenResponse struct {
	// The generated or provided short code
	ShortCode string `json:"short_code"`

	// The complete shortened URL
	ShortURL string `json:"short_url"`

	// Expiration timestamp (null if no expiration)
	ExpiresAt *time.Time `json:"expires_at"`
}

// BatchItem represents a single URL item in a batch shorten request
//
// swagger:model BatchItem
type BatchItem struct {
	// The original URL to be shortened
	// required: true
	URL string `json:"url"`

	// Expiration time in seconds from now (optional)
	ExpiresIn *int `json:"expires_in,omitempty"`

	// Custom alias for the shortened URL (optional, must be unique)
	Alias *string `json:"alias,omitempty"`
}

// BatchResult represents the result of shortening a single URL in a batch operation
//
// swagger:model BatchResult
type BatchResult struct {
	// The original URL that was processed
	URL string `json:"url"`

	// The shortened URL (empty if error occurred)
	Short string `json:"short"`

	// Error message if shortening failed (empty if successful)
	Error string `json:"error,omitempty"`
}

func (s *service) Shorten(ctx context.Context, originalURL string, expiresIn *int, alias *string) (*ShortenResponse, error) {
	if err := validateURL(originalURL); err != nil {
		return nil, err
	}

	var shortCode string
	if alias != nil && *alias != "" {
		shortCode = *alias
		exists, err := s.dao.CheckShortCodeExists(ctx, shortCode)
		if err != nil {
			return nil, fmt.Errorf("failed to check alias: %w", err)
		}
		if exists {
			return nil, ErrAliasExists
		}
	} else {
		var err error
		shortCode, err = s.generateUniqueShortCode(ctx)
		if err != nil {
			return nil, err
		}
	}

	now := time.Now().UTC()
	var expiresAt *time.Time
	if expiresIn != nil {
		exp := now.Add(time.Duration(*expiresIn) * time.Second)
		expiresAt = &exp
	}

	urlEntity := &entity.URL{
		ID:          uuid.Generate(),
		ShortCode:   shortCode,
		OriginalURL: originalURL,
		ExpiresAt:   expiresAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.CreateURL(ctx, urlEntity); err != nil {
		return nil, fmt.Errorf("failed to create URL: %w", err)
	}

	s.bloomFilter.Add([]byte(shortCode))

	var ttl time.Duration
	if expiresAt != nil {
		ttl = time.Until(*expiresAt)
		if ttl > 0 {
			_ = s.urlCache.SetURL(ctx, shortCode, originalURL, ttl)
		}
	} else {
		_ = s.urlCache.SetURL(ctx, shortCode, originalURL, 365*24*time.Hour)
	}

	shortURL := fmt.Sprintf("%s/%s", s.domain, shortCode)

	return &ShortenResponse{
		ShortCode: shortCode,
		ShortURL:  shortURL,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *service) ShortenBatch(ctx context.Context, items []BatchItem) ([]BatchResult, error) {
	results := make([]BatchResult, 0, len(items))
	for _, item := range items {
		resp, err := s.Shorten(ctx, item.URL, item.ExpiresIn, item.Alias)
		if err != nil {
			results = append(results, BatchResult{
				URL:   item.URL,
				Short: "",
				Error: err.Error(),
			})
			continue
		}
		results = append(results, BatchResult{
			URL:   item.URL,
			Short: resp.ShortURL,
		})
	}
	return results, nil
}

func (s *service) GetOriginalURL(ctx context.Context, shortCode string) (string, error) {
	if !s.bloomFilter.Test([]byte(shortCode)) {
		return "", ErrURLNotFound
	}

	cachedURL, err := s.urlCache.GetURL(ctx, shortCode)
	if err == nil {
		return cachedURL, nil
	}

	urlEntity, err := s.dao.GetURLByShortCode(ctx, shortCode)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return "", ErrURLNotFound
		}
		return "", fmt.Errorf("failed to get URL: %w", err)
	}

	if urlEntity.ExpiresAt != nil && time.Now().UTC().After(*urlEntity.ExpiresAt) {
		return "", ErrURLExpired
	}

	var ttl time.Duration
	if urlEntity.ExpiresAt != nil {
		ttl = time.Until(*urlEntity.ExpiresAt)
		if ttl > 0 {
			_ = s.urlCache.SetURL(ctx, shortCode, urlEntity.OriginalURL, ttl)
		}
	} else {
		_ = s.urlCache.SetURL(ctx, shortCode, urlEntity.OriginalURL, 365*24*time.Hour)
	}

	return urlEntity.OriginalURL, nil
}

func (s *service) generateUniqueShortCode(ctx context.Context) (string, error) {
	maxAttempts := 10
	for i := 0; i < maxAttempts; i++ {
		code := generateShortCode(s.shortCodeLen)
		exists, err := s.dao.CheckShortCodeExists(ctx, code)
		if err != nil {
			return "", fmt.Errorf("failed to check short code: %w", err)
		}
		if !exists {
			return code, nil
		}
	}
	return "", fmt.Errorf("%w: after %d attempts", ErrShortCodeGeneration, maxAttempts)
}

func validateURL(u string) error {
	parsed, err := url.Parse(u)
	if err != nil {
		return ErrInvalidURLFormat
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ErrInvalidURLScheme
	}
	return nil
}

func generateShortCode(length int) string {
	bytes := make([]byte, length*3/4+1)
	_, _ = rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)[:length]
}
