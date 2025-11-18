// Package errors provides common error types used across the application.
package errors

// ErrorCode represents a unique error code identifier.
type ErrorCode string

const (
	// Validation errors
	ErrCodeValidation       ErrorCode = "ERR_VALIDATION"
	ErrCodeBadRequest       ErrorCode = "ERR_BAD_REQUEST"
	ErrCodeInvalidURL       ErrorCode = "ERR_INVALID_URL"
	ErrCodeInvalidURLFormat ErrorCode = "ERR_INVALID_URL_FORMAT"
	ErrCodeInvalidURLScheme ErrorCode = "ERR_INVALID_URL_SCHEME"

	// Not found errors
	ErrCodeNotFound ErrorCode = "ERR_NOT_FOUND"

	// Conflict errors
	ErrCodeConflict    ErrorCode = "ERR_CONFLICT"
	ErrCodeAliasExists ErrorCode = "ERR_ALIAS_EXISTS"

	// Expired/Gone errors
	ErrCodeExpired ErrorCode = "ERR_EXPIRED"

	// Authentication/Authorization errors
	ErrCodeUnauthorized ErrorCode = "ERR_UNAUTHORIZED"
	ErrCodeForbidden    ErrorCode = "ERR_FORBIDDEN"

	// Internal errors
	ErrCodeInternal            ErrorCode = "ERR_INTERNAL"
	ErrCodeShortCodeGeneration ErrorCode = "ERR_SHORT_CODE_GENERATION"
)

// Resource names used in error messages
const (
	ResourceURL       = "URL"
	ResourceShortCode = "ShortCode"
	ResourceAlias     = "Alias"
)
