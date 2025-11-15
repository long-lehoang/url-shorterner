// Package uuid provides utilities for generating unique identifiers.
package uuid

import (
	"crypto/rand"
	"encoding/base64"
)

// Generate generates a unique identifier using cryptographically secure random bytes.
func Generate() string {
	bytes := make([]byte, 16)
	_, _ = rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

