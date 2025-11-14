package storage

import "errors"

var (
	ErrNotFound = errors.New("url not found")
	ErrExpired  = errors.New("url expired")
)

