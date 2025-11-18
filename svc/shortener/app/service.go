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
	appErrors "url-shorterner/internal/errors"
	eventsPublisher "url-shorterner/internal/events"
	"url-shorterner/internal/storage"
	"url-shorterner/internal/uuid"
	analyticsEvents "url-shorterner/svc/analytics/events"
	"url-shorterner/svc/shortener/entity"
	shortenerStore "url-shorterner/svc/shortener/store"

	"github.com/bits-and-blooms/bloom/v3"
)

type Service interface {
	Shorten(ctx context.Context, originalURL string, expiresIn *int, alias *string) (*ShortenResponse, error)
	ShortenBatch(ctx context.Context, items []BatchItem) ([]BatchResult, error)
	GetOriginalURL(ctx context.Context, shortCode string, clickInfo *ClickInfo) (string, error)
}

type ClickInfo struct {
	IPAddress string
	UserAgent string
	Referer   string
}

type service struct {
	repo         shortenerStore.Repository
	dao          shortenerStore.DAO
	urlCache     *cache.URLCache
	bloomFilter  *bloom.BloomFilter
	shortCodeLen int
	domain       string
	publisher    eventsPublisher.Publisher
}

func NewService(
	repo shortenerStore.Repository,
	dao shortenerStore.DAO,
	urlCache *cache.URLCache,
	bloomN uint,
	bloomP float64,
	shortCodeLen int,
	domain string,
	publisher eventsPublisher.Publisher,
) Service {
	bf := bloom.NewWithEstimates(bloomN, bloomP)
	return &service{
		repo:         repo,
		dao:          dao,
		urlCache:     urlCache,
		bloomFilter:  bf,
		shortCodeLen: shortCodeLen,
		domain:       domain,
		publisher:    publisher,
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
			return nil, appErrors.Invalid(appErrors.ErrCodeInternal, map[string]interface{}{"Message": "failed to check alias"})
		}
		if exists {
			return nil, appErrors.Conflict(appErrors.ErrCodeAliasExists, nil)
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
		return nil, appErrors.Invalid(appErrors.ErrCodeInternal, map[string]interface{}{"Message": "failed to create URL"})
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

func (s *service) GetOriginalURL(ctx context.Context, shortCode string, clickInfo *ClickInfo) (string, error) {
	if !s.bloomFilter.Test([]byte(shortCode)) {
		return "", appErrors.NotFound(appErrors.ResourceURL)
	}

	cachedURL, err := s.urlCache.GetURL(ctx, shortCode)
	if err == nil {
		s.publishClickEvent(ctx, shortCode, clickInfo)
		return cachedURL, nil
	}

	urlEntity, err := s.dao.GetURLByShortCode(ctx, shortCode)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return "", appErrors.NotFound(appErrors.ResourceURL)
		}
		return "", appErrors.Invalid(appErrors.ErrCodeInternal, map[string]interface{}{"Message": "failed to get URL"})
	}

	if urlEntity.ExpiresAt != nil && time.Now().UTC().After(*urlEntity.ExpiresAt) {
		return "", appErrors.Expired(appErrors.ErrCodeExpired, map[string]interface{}{"Resource": appErrors.ResourceURL})
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

	s.publishClickEvent(ctx, shortCode, clickInfo)
	return urlEntity.OriginalURL, nil
}

func (s *service) publishClickEvent(ctx context.Context, shortCode string, clickInfo *ClickInfo) {
	if s.publisher == nil || clickInfo == nil {
		return
	}

	go func() {
		clickEvent := analyticsEvents.ClickEvent{
			ShortCode: shortCode,
			IPAddress: clickInfo.IPAddress,
			UserAgent: clickInfo.UserAgent,
			Referer:   clickInfo.Referer,
			Timestamp: time.Now().UTC(),
		}
		if err := s.publisher.PublishClickEvent(ctx, clickEvent); err != nil {
			// Log error but don't fail the redirect
			// In production, consider using a proper logger
			_ = err
		}
	}()
}

func (s *service) generateUniqueShortCode(ctx context.Context) (string, error) {
	maxAttempts := 10
	for i := 0; i < maxAttempts; i++ {
		code := generateShortCode(s.shortCodeLen)
		exists, err := s.dao.CheckShortCodeExists(ctx, code)
		if err != nil {
			return "", appErrors.Invalid(appErrors.ErrCodeInternal, map[string]interface{}{"Message": "failed to check short code"})
		}
		if !exists {
			return code, nil
		}
	}
	return "", appErrors.Invalid(appErrors.ErrCodeShortCodeGeneration, map[string]interface{}{"Attempts": maxAttempts})
}

func validateURL(u string) error {
	parsed, err := url.Parse(u)
	if err != nil {
		return appErrors.Invalid(appErrors.ErrCodeInvalidURLFormat, nil)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return appErrors.Invalid(appErrors.ErrCodeInvalidURLScheme, nil)
	}
	return nil
}

func generateShortCode(length int) string {
	bytes := make([]byte, length*3/4+1)
	_, _ = rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)[:length]
}
