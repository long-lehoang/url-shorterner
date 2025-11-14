package shortener

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net/url"
	"strings"
	"time"

	"github.com/willf/bloom"
	"url-shorterner/internal/cache"
	"url-shorterner/internal/storage"
)

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

type Service interface {
	Shorten(ctx context.Context, originalURL string, expiresIn *int, alias *string) (*ShortenResponse, error)
	ShortenBatch(ctx context.Context, items []BatchItem) ([]BatchResult, error)
	GetOriginalURL(ctx context.Context, shortCode string) (string, error)
}

type ShortenResponse struct {
	ShortCode string    `json:"short_code"`
	ShortURL  string    `json:"short_url"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type BatchItem struct {
	URL       string `json:"url"`
	ExpiresIn *int   `json:"expires_in,omitempty"`
	Alias     *string `json:"alias,omitempty"`
}

type BatchResult struct {
	URL      string `json:"url"`
	Short    string `json:"short"`
	Error    string `json:"error,omitempty"`
}

type service struct {
	repo        storage.Repository
	urlCache    *cache.URLCache
	bloomFilter *bloom.BloomFilter
	codeLength  int
	domain      string
}

func NewService(repo storage.Repository, urlCache *cache.URLCache, bloomN uint, bloomP float64, codeLength int, domain string) Service {
	bf := bloom.NewWithEstimates(bloomN, bloomP)
	return &service{
		repo:        repo,
		urlCache:    urlCache,
		bloomFilter: bf,
		codeLength:  codeLength,
		domain:      domain,
	}
}

func (s *service) Shorten(ctx context.Context, originalURL string, expiresIn *int, alias *string) (*ShortenResponse, error) {
	if err := s.validateURL(originalURL); err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	var shortCode string
	if alias != nil && *alias != "" {
		shortCode = *alias
		exists, err := s.repo.CheckShortCodeExists(ctx, shortCode)
		if err != nil {
			return nil, fmt.Errorf("failed to check alias: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("alias already exists")
		}
	} else {
		var err error
		shortCode, err = s.generateUniqueCode(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to generate code: %w", err)
		}
	}

	var expiresAt *time.Time
	if expiresIn != nil {
		exp := time.Now().UTC().Add(time.Duration(*expiresIn) * time.Second)
		expiresAt = &exp
	}

	_, err := s.repo.CreateURL(ctx, shortCode, originalURL, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create URL: %w", err)
	}

	var ttl time.Duration
	if expiresAt != nil {
		ttl = time.Until(*expiresAt)
		if ttl <= 0 {
			return nil, fmt.Errorf("expires_in is in the past")
		}
	} else {
		ttl = 365 * 24 * time.Hour
	}

	if err := s.urlCache.SetURL(ctx, shortCode, originalURL, ttl); err != nil {
	}

	s.bloomFilter.Add([]byte(shortCode))

	return &ShortenResponse{
		ShortCode: shortCode,
		ShortURL:  fmt.Sprintf("%s/%s", strings.TrimSuffix(s.domain, "/"), shortCode),
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
		return "", storage.ErrNotFound
	}

	originalURL, err := s.urlCache.GetURL(ctx, shortCode)
	if err == nil {
		return originalURL, nil
	}

	urlRecord, err := s.repo.GetURLByShortCode(ctx, shortCode)
	if err != nil {
		return "", err
	}

	var ttl time.Duration
	if urlRecord.ExpiresAt != nil {
		ttl = time.Until(*urlRecord.ExpiresAt)
		if ttl <= 0 {
			return "", storage.ErrExpired
		}
	} else {
		ttl = 365 * 24 * time.Hour
	}

	if err := s.urlCache.SetURL(ctx, shortCode, urlRecord.OriginalURL, ttl); err != nil {
	}

	return urlRecord.OriginalURL, nil
}

func (s *service) generateUniqueCode(ctx context.Context) (string, error) {
	maxAttempts := 10
	for i := 0; i < maxAttempts; i++ {
		code := s.generateRandomCode()
		exists, err := s.repo.CheckShortCodeExists(ctx, code)
		if err != nil {
			return "", err
		}
		if !exists {
			return code, nil
		}
	}
	return "", fmt.Errorf("failed to generate unique code after %d attempts", maxAttempts)
}

func (s *service) generateRandomCode() string {
	var result strings.Builder
	result.Grow(s.codeLength)

	for i := 0; i < s.codeLength; i++ {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(base62Chars))))
		result.WriteByte(base62Chars[idx.Int64()])
	}

	return result.String()
}

func (s *service) validateURL(rawURL string) error {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL format")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL must use http or https scheme")
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("URL must have a host")
	}

	if strings.HasPrefix(parsedURL.Host, "localhost") || strings.HasPrefix(parsedURL.Host, "127.0.0.1") {
		return fmt.Errorf("localhost URLs are not allowed")
	}

	return nil
}

