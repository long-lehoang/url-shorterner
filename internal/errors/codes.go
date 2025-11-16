// Package errors provides common error types used across the application.
package errors

// ErrorCode represents a unique error code identifier.
type ErrorCode string

const (
	// Validation errors
	ErrCodeValidation ErrorCode = "ERR_VALIDATION"
	ErrCodeBadRequest ErrorCode = "ERR_BAD_REQUEST"

	// Not found errors
	ErrCodeNotFound ErrorCode = "ERR_NOT_FOUND"

	// Conflict errors
	ErrCodeConflict ErrorCode = "ERR_CONFLICT"

	// Authentication/Authorization errors
	ErrCodeUnauthorized ErrorCode = "ERR_UNAUTHORIZED"
	ErrCodeForbidden    ErrorCode = "ERR_FORBIDDEN"

	// Internal errors
	ErrCodeInternal ErrorCode = "ERR_INTERNAL"
)
