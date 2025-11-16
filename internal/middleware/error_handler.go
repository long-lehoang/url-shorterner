// Package middleware provides HTTP middleware functions for rate limiting, metrics, logging, and error handling.
package middleware

import (
	"errors"
	"net/http"

	appErrors "url-shorterner/internal/errors"
	"url-shorterner/svc/shortener/app"

	"github.com/gin-gonic/gin"
)

// ErrorHandler returns a Gin middleware that handles errors from c.Errors.
// It processes errors added via c.Error() and returns appropriate HTTP responses.
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Process errors if any were added to the context
		// Note: Errors are already logged by the Logger middleware
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			if err != nil {
				status := mapErrorToStatus(err)
				lang := appErrors.GetLanguageFromContext(c)

				// Get error message and code
				var errorMsg string
				var errorCode appErrors.ErrorCode

				// Extract error code and translate message based on request language
				if code, ok := appErrors.GetErrorCode(err); ok {
					errorCode = code
					// Handle special case for NotFoundError with resource
					if notFoundErr, ok := err.(*appErrors.NotFoundError); ok {
						errorMsg = appErrors.GetMessage(code, lang, map[string]interface{}{"Resource": notFoundErr.Resource})
					} else {
						errorMsg = appErrors.GetMessage(code, lang)
					}
				} else {
					// Fallback to error message if no code found
					errorMsg = err.Error()
				}

				// For internal server errors, don't expose internal error details to clients
				if status == http.StatusInternalServerError {
					errorCode = appErrors.ErrCodeInternal
					errorMsg = appErrors.GetMessage(appErrors.ErrCodeInternal, lang)
				}

				response := gin.H{"error": errorMsg}
				if errorCode != "" {
					response["code"] = string(errorCode)
				}

				c.JSON(status, response)
				c.Abort()
			}
		}
	}
}

// mapErrorToStatus maps domain errors to HTTP status codes.
func mapErrorToStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}

	// Check for common HTTP error types from internal/errors
	if status := appErrors.StatusCode(err); status != 500 {
		return status
	}

	// Map shortener domain errors
	if errors.Is(err, app.ErrAliasExists) {
		return http.StatusConflict
	}
	if errors.Is(err, app.ErrInvalidURL) || errors.Is(err, app.ErrInvalidURLFormat) || errors.Is(err, app.ErrInvalidURLScheme) {
		return http.StatusBadRequest
	}
	if errors.Is(err, app.ErrURLNotFound) {
		return http.StatusNotFound
	}
	if errors.Is(err, app.ErrURLExpired) {
		return http.StatusGone
	}

	// Default to internal server error
	return http.StatusInternalServerError
}
