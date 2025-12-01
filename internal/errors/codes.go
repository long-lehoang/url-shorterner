// Package errors provides common error types used across the application.
package errors

// ErrorCode represents a unique error code identifier.
type ErrorCode string

const (
	// ErrCodeValidation indicates a validation error.
	ErrCodeValidation ErrorCode = "ERR_VALIDATION"
	// ErrCodeBadRequest indicates a bad request error.
	ErrCodeBadRequest ErrorCode = "ERR_BAD_REQUEST"
	// ErrCodeInvalidURL indicates an invalid URL error.
	ErrCodeInvalidURL ErrorCode = "ERR_INVALID_URL"
	// ErrCodeInvalidURLFormat indicates an invalid URL format error.
	ErrCodeInvalidURLFormat ErrorCode = "ERR_INVALID_URL_FORMAT"
	// ErrCodeInvalidURLScheme indicates an invalid URL scheme error.
	ErrCodeInvalidURLScheme ErrorCode = "ERR_INVALID_URL_SCHEME"

	// ErrCodeNotFound indicates a resource not found error.
	ErrCodeNotFound ErrorCode = "ERR_NOT_FOUND"

	// ErrCodeConflict indicates a conflict error.
	ErrCodeConflict ErrorCode = "ERR_CONFLICT"
	// ErrCodeAliasExists indicates that an alias already exists.
	ErrCodeAliasExists ErrorCode = "ERR_ALIAS_EXISTS"

	// ErrCodeExpired indicates that a resource has expired.
	ErrCodeExpired ErrorCode = "ERR_EXPIRED"

	// ErrCodeUnauthorized indicates an unauthorized access error.
	ErrCodeUnauthorized ErrorCode = "ERR_UNAUTHORIZED"
	// ErrCodeForbidden indicates a forbidden access error.
	ErrCodeForbidden ErrorCode = "ERR_FORBIDDEN"

	// ErrCodeInternal indicates an internal server error.
	ErrCodeInternal ErrorCode = "ERR_INTERNAL"
	// ErrCodeShortCodeGeneration indicates a failure to generate a unique short code.
	ErrCodeShortCodeGeneration ErrorCode = "ERR_SHORT_CODE_GENERATION"
)

// Resource names used in error messages
const (
	ResourceURL       = "URL"
	ResourceShortCode = "ShortCode"
	ResourceAlias     = "Alias"
)
