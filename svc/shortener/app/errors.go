// Package app defines domain-specific error types for the shortener service.
package app

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidURL          = errors.New("invalid URL")
	ErrInvalidURLFormat    = fmt.Errorf("%w: invalid URL format", ErrInvalidURL)
	ErrInvalidURLScheme    = fmt.Errorf("%w: URL must use http or https scheme", ErrInvalidURL)
	ErrAliasExists         = errors.New("alias already exists")
	ErrURLNotFound         = errors.New("url not found")
	ErrURLExpired          = errors.New("url expired")
	ErrShortCodeGeneration = errors.New("failed to generate unique short code")
)
