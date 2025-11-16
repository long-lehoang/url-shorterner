// Package uuid provides utilities for generating unique identifiers.
package uuid

import (
	"github.com/google/uuid"
)

// Generate generates a UUID v4 using cryptographically secure random bytes.
func Generate() string {
	return uuid.New().String()
}
