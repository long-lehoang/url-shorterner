// Package errors provides common error types used across the application.
package errors

import "errors"

// ValidationError represents a validation error.
type ValidationError struct {
	code    ErrorCode
	message string
}

// Ensure ValidationError implements CodedError
var _ CodedError = (*ValidationError)(nil)

// NewValidationError creates a new validation error with a message.
func NewValidationError(message string) *ValidationError {
	return &ValidationError{
		code:    ErrCodeValidation,
		message: message,
	}
}

func (e *ValidationError) Error() string {
	return e.message
}

// Code returns the error code.
func (e *ValidationError) Code() ErrorCode {
	return e.code
}

// NotFoundError represents a 404 Not Found error.
type NotFoundError struct {
	errorCode ErrorCode
	Resource  string
	Message   string
}

// Ensure NotFoundError implements CodedError
var _ CodedError = (*NotFoundError)(nil)

func (e *NotFoundError) Error() string {
	return e.Message
}

// Code returns the error code.
func (e *NotFoundError) Code() ErrorCode {
	return e.errorCode
}

// StatusCode returns the HTTP status code for an error.
// It checks if the error implements CodedError interface or is a known error type.
func StatusCode(err error) int {
	if err == nil {
		return 200
	}

	// Check for ValidationError
	var validationErr *ValidationError
	if errors.As(err, &validationErr) {
		return 400
	}

	// Check for NotFoundError
	var notFoundErr *NotFoundError
	if errors.As(err, &notFoundErr) {
		return 404
	}

	// Check if error has a code and map based on error code
	if code, ok := GetErrorCode(err); ok {
		switch code {
		case ErrCodeBadRequest, ErrCodeValidation:
			return 400
		case ErrCodeNotFound:
			return 404
		case ErrCodeConflict:
			return 409
		case ErrCodeUnauthorized:
			return 401
		case ErrCodeForbidden:
			return 403
		case ErrCodeInternal:
			return 500
		}
	}

	// Default to 500 for unknown errors
	return 500
}
