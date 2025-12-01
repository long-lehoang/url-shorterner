// Package errors provides common error types used across the application.
package errors

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

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

// NewNotFoundError creates a new not found error.
func NewNotFoundError(resource string) *NotFoundError {
	return &NotFoundError{
		errorCode: ErrCodeNotFound,
		Resource:  resource,
		Message:   GetMessage(ErrCodeNotFound, DefaultLanguage, map[string]interface{}{"Resource": resource}),
	}
}

func (e *NotFoundError) Error() string {
	return e.Message
}

// Code returns the error code.
func (e *NotFoundError) Code() ErrorCode {
	return e.errorCode
}

// ConflictError represents a 409 Conflict error.
type ConflictError struct {
	code    ErrorCode
	message string
}

// Ensure ConflictError implements CodedError
var _ CodedError = (*ConflictError)(nil)

// NewConflictError creates a new conflict error with a message.
func NewConflictError(message string) *ConflictError {
	return &ConflictError{
		code:    ErrCodeConflict,
		message: message,
	}
}

func (e *ConflictError) Error() string {
	return e.message
}

// Code returns the error code.
func (e *ConflictError) Code() ErrorCode {
	return e.code
}

// GoneError represents a 410 Gone error (resource expired).
type GoneError struct {
	code    ErrorCode
	message string
}

// Ensure GoneError implements CodedError
var _ CodedError = (*GoneError)(nil)

// NewGoneError creates a new gone error with a message.
func NewGoneError(message string) *GoneError {
	return &GoneError{
		code:    ErrCodeNotFound, // Use NotFound code, but StatusCode will map to 410
		message: message,
	}
}

func (e *GoneError) Error() string {
	return e.message
}

// Code returns the error code.
func (e *GoneError) Code() ErrorCode {
	return e.code
}

// InvalidError represents a validation/invalid input error.
type InvalidError struct {
	Code    ErrorCode
	Message string
	Data    map[string]interface{}
}

func (e *InvalidError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	// Fallback to code if no message
	return string(e.Code)
}

// GetCode returns the error code for i18n translation.
func (e *InvalidError) GetCode() ErrorCode {
	if e.Code != "" {
		return e.Code
	}
	return ErrCodeValidation
}

// Invalid creates a new InvalidError with an error code and optional context data.
// The message will be translated in the error handler based on request language.
func Invalid(code ErrorCode, data map[string]interface{}) *InvalidError {
	return &InvalidError{
		Code: code,
		Data: data,
	}
}

// InvalidWithMessage creates a new InvalidError with a message (for backward compatibility).
func InvalidWithMessage(message string) *InvalidError {
	return &InvalidError{
		Code:    ErrCodeValidation,
		Message: message,
	}
}

// DomainNotFoundError represents a domain-specific resource not found error.
type DomainNotFoundError struct {
	Code     ErrorCode
	Resource string
	Message  string
}

func (e *DomainNotFoundError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("%s not found", e.Resource)
}

// GetCode returns the error code for i18n translation.
func (e *DomainNotFoundError) GetCode() ErrorCode {
	if e.Code != "" {
		return e.Code
	}
	return ErrCodeNotFound
}

// NotFound creates a new DomainNotFoundError.
// The message will be translated in the error handler based on request language.
// resource should be one of the Resource constants (e.g., ResourceURL, ResourceShortCode).
func NotFound(resource string) *DomainNotFoundError {
	return &DomainNotFoundError{
		Code:     ErrCodeNotFound,
		Resource: resource,
	}
}

// DomainConflictError represents a domain-specific conflict error (e.g., duplicate resource).
type DomainConflictError struct {
	Code    ErrorCode
	Message string
	Data    map[string]interface{}
}

func (e *DomainConflictError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return string(e.Code)
}

// GetCode returns the error code for i18n translation.
func (e *DomainConflictError) GetCode() ErrorCode {
	if e.Code != "" {
		return e.Code
	}
	return ErrCodeConflict
}

// Conflict creates a new DomainConflictError with an error code and optional context data.
// The message will be translated in the error handler based on request language.
func Conflict(code ErrorCode, data map[string]interface{}) *DomainConflictError {
	return &DomainConflictError{
		Code: code,
		Data: data,
	}
}

// ConflictWithMessage creates a new DomainConflictError with a message (for backward compatibility).
func ConflictWithMessage(message string) *DomainConflictError {
	return &DomainConflictError{
		Code:    ErrCodeConflict,
		Message: message,
	}
}

// DomainExpiredError represents a domain-specific expired resource error.
type DomainExpiredError struct {
	Code    ErrorCode
	Message string
	Data    map[string]interface{}
}

func (e *DomainExpiredError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return string(e.Code)
}

// GetCode returns the error code for i18n translation.
func (e *DomainExpiredError) GetCode() ErrorCode {
	if e.Code != "" {
		return e.Code
	}
	return ErrCodeNotFound // Will be mapped to 410 in StatusCode
}

// Expired creates a new DomainExpiredError with an error code and optional context data.
// The message will be translated in the error handler based on request language.
func Expired(code ErrorCode, data map[string]interface{}) *DomainExpiredError {
	return &DomainExpiredError{
		Code: code,
		Data: data,
	}
}

// ExpiredWithMessage creates a new DomainExpiredError with a message (for backward compatibility).
func ExpiredWithMessage(message string) *DomainExpiredError {
	return &DomainExpiredError{
		Code:    ErrCodeNotFound,
		Message: message,
	}
}

// StatusCode returns the HTTP status code for an error.
// It checks if the error implements CodedError interface or is a known error type.
// It also checks for typed domain errors (like app.InvalidError) and maps them appropriately.
func StatusCode(err error) int {
	if err == nil {
		return 200
	}

	// Check for typed domain errors first (before conversion)
	errType := getErrorTypeName(err)
	switch errType {
	case "*errors.InvalidError":
		return 400 // BadRequest
	case "*errors.DomainNotFoundError":
		return 404
	case "*errors.DomainConflictError":
		return 409
	case "*errors.DomainExpiredError":
		return 410 // Gone
	}

	// Check for GoneError (410)
	var goneErr *GoneError
	if errors.As(err, &goneErr) {
		return 410
	}

	// Check for ValidationError
	var validationErr *ValidationError
	if errors.As(err, &validationErr) {
		return 400
	}

	// Check for ConflictError
	var conflictErr *ConflictError
	if errors.As(err, &conflictErr) {
		return 409
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

// ConvertError converts domain-specific errors to generic errors.
// It checks for typed domain errors using reflection and converts them to generic errors.
// Domain-specific errors are converted to generic errors to maintain proper dependency flow.
func ConvertError(err error) error {
	if err == nil {
		return nil
	}

	// Check if error is already a generic error type
	if _, ok := GetErrorCode(err); ok {
		return err
	}

	// Check for typed domain errors using reflection
	errType := getErrorTypeName(err)

	switch errType {
	case "*errors.InvalidError":
		// Extract error code from InvalidError
		invalidErr := err.(*InvalidError)
		code := invalidErr.GetCode()
		// Preserve the code for translation in handler, message will be ignored
		return &ValidationError{
			code:    code,
			message: "", // Empty message - handler will translate based on code
		}
	case "*errors.DomainNotFoundError":
		// Extract resource name from DomainNotFoundError struct using reflection
		resource := "Resource"
		rv := reflect.ValueOf(err)
		if rv.Kind() == reflect.Ptr && !rv.IsNil() {
			elem := rv.Elem()
			if elem.Kind() == reflect.Struct {
				resourceField := elem.FieldByName("Resource")
				if resourceField.IsValid() && resourceField.Kind() == reflect.String {
					resource = resourceField.String()
				}
			}
		}
		return NewNotFoundError(resource)
	case "*errors.DomainConflictError":
		// Extract error code and data from DomainConflictError
		conflictErr := err.(*DomainConflictError)
		code := conflictErr.GetCode()
		// Preserve the code for translation in handler, message will be ignored
		return &ConflictError{
			code:    code,
			message: "", // Empty message - handler will translate based on code
		}
	case "*errors.DomainExpiredError":
		// Use ERR_EXPIRED code for translation
		return &GoneError{
			code:    ErrCodeExpired,
			message: "", // Empty message - handler will translate based on code
		}
	}

	// Fallback to message-based pattern matching for legacy errors
	errMsg := strings.ToLower(err.Error())

	// Handle common validation patterns
	if strings.Contains(errMsg, "invalid") || strings.Contains(errMsg, "validation") || strings.Contains(errMsg, "required") {
		return NewValidationError(err.Error())
	}

	// Handle not found patterns
	if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "does not exist") {
		return NewNotFoundError("Resource")
	}

	// Handle conflict patterns
	if strings.Contains(errMsg, "already exists") || strings.Contains(errMsg, "duplicate") || strings.Contains(errMsg, "conflict") {
		return NewConflictError(err.Error())
	}

	// Handle expired/gone patterns
	if strings.Contains(errMsg, "expired") || strings.Contains(errMsg, "gone") {
		return NewGoneError(err.Error())
	}

	// Return error as-is if no pattern matches
	return err
}

// getErrorTypeName returns the type name of an error for type checking.
func getErrorTypeName(err error) string {
	if err == nil {
		return ""
	}
	// Use reflection to get the type name
	return fmt.Sprintf("%T", err)
}
