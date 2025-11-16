// Package errors provides common error types used across the application.
package errors

import (
	"url-shorterner/internal/i18n"

	"github.com/gin-gonic/gin"
)

// GetLanguageFromContext is re-exported from i18n package for convenience.
func GetLanguageFromContext(c *gin.Context) Language {
	return i18n.GetLanguageFromContext(c)
}

// ContextKeyLanguage is re-exported from i18n package.
const ContextKeyLanguage = i18n.ContextKeyLanguage
