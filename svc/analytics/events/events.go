// Package events defines event types for the analytics domain.
package events

import "time"

type ClickEvent struct {
	ShortCode string
	IPAddress string
	UserAgent string
	Referer   string
	Timestamp time.Time
}
