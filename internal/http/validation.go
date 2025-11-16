// Package http provides common HTTP utilities and router setup functions.
package http

import (
	"url-shorterner/internal/errors"

	"github.com/gin-gonic/gin"
)

// BindAndValidate binds the request body to a struct and validates it.
// If validation fails, it adds an error to c.Errors and returns false.
// The ErrorHandler middleware will process the error and send the response.
// Gin's validator already provides readable error messages, so we use them directly.
func BindAndValidate(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		c.Error(errors.NewValidationError(err.Error()))
		c.Abort()
		return false
	}
	return true
}

// BindQuery binds query parameters to a struct and validates it.
func BindQuery(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBindQuery(obj); err != nil {
		c.Error(errors.NewValidationError(err.Error()))
		c.Abort()
		return false
	}
	return true
}

// BindURI binds URI parameters to a struct and validates it.
func BindURI(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBindUri(obj); err != nil {
		c.Error(errors.NewValidationError(err.Error()))
		c.Abort()
		return false
	}
	return true
}
