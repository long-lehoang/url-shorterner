// Package entity defines domain entities for the analytics service.
package entity

import "time"

// Record represents a single click record for analytics
//
// swagger:model Record
type Record struct {
	// Unique identifier for the click record
	ID string

	// The short code that was clicked
	ShortCode string

	// IP address of the user who clicked
	IPAddress string

	// User agent string from the HTTP request
	UserAgent string

	// Referer header from the HTTP request
	Referer string

	// Timestamp when the click occurred
	ClickedAt time.Time
}

// Stats represents aggregated analytics statistics
//
// swagger:model Stats
type Stats struct {
	// Total number of clicks
	TotalClicks int

	// Number of unique IP addresses
	UniqueIPs int

	// Timestamp of the last click (null if no clicks)
	LastClick *time.Time
}

