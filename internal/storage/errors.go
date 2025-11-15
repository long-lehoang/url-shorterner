// Package storage defines storage layer error types.
package storage

import "errors"

var (
	// ErrNotFound is returned when a requested resource is not found.
	ErrNotFound = errors.New("url not found")
	// ErrExpired is returned when a URL has expired.
	ErrExpired = errors.New("url expired")
)

