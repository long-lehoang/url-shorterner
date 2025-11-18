// Package middleware provides HTTP middleware functions for rate limiting, metrics, logging, and error handling.
package middleware

import (
	"net/http"

	appErrors "url-shorterner/internal/errors"

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
				// Convert domain-specific errors to generic errors
				err = appErrors.ConvertError(err)

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
						// For all other errors, use the code for translation
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

// mapErrorToStatus maps errors to HTTP status codes using internal/errors.StatusCode.
// Domain-specific errors should be converted to generic errors in the transport layer.
func mapErrorToStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}

	// Use StatusCode from internal/errors which handles all error types
	status := appErrors.StatusCode(err)
	if status != 500 {
		return status
	}

	// Default to internal server error for unknown errors
	return http.StatusInternalServerError
}
