// Package entity defines domain entities for the shortener service.
package entity

import "time"

// URL represents a shortened URL entity.
type URL struct {
	ID          string
	ShortCode   string
	OriginalURL string
	ExpiresAt   *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

