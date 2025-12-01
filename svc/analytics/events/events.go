// Package events defines event types for the analytics domain.
package events

import "time"

// ClickEvent represents a click event for analytics tracking.
type ClickEvent struct {
	ShortCode string
	IPAddress string
	UserAgent string
	Referer   string
	Timestamp time.Time
}
