// Package i18n provides internationalization support for the application.
package i18n

import (
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/text/language"
)

// Language represents a supported language.
type Language string

const (
	// LanguageEN is English language.
	LanguageEN Language = "en"
	// LanguageVI is Vietnamese language.
	LanguageVI Language = "vi"
	// Add more languages as needed
)

// DefaultLanguage is the default language used when no language is specified.
const DefaultLanguage = LanguageEN

// SupportedLanguages is a list of supported language codes.
var SupportedLanguages = []string{"en", "vi"}

const (
	// ContextKeyLanguage is the key used to store language in Gin context.
	ContextKeyLanguage = "language"
)

// GetLanguageFromContext extracts the language from Gin context.
// It checks for Accept-Language header or a custom language parameter.
// Returns DefaultLanguage if not found.
func GetLanguageFromContext(c *gin.Context) Language {
	// Check if language is already set in context (e.g., by middleware)
	if lang, exists := c.Get(ContextKeyLanguage); exists {
		if l, ok := lang.(Language); ok {
			return l
		}
		if l, ok := lang.(string); ok {
			return Language(l)
		}
	}

	// Check Accept-Language header
	acceptLang := c.GetHeader("Accept-Language")
	if acceptLang != "" {
		// Parse Accept-Language header using golang.org/x/text/language
		tags, _, _ := language.ParseAcceptLanguage(acceptLang)
		for _, tag := range tags {
			base, _ := tag.Base()
			langCode := base.String()
			// Check if it's a supported language
			for _, supported := range SupportedLanguages {
				if langCode == supported {
					return Language(langCode)
				}
			}
		}
	}

	// Check query parameter
	if langParam := c.Query("lang"); langParam != "" {
		langCode := strings.ToLower(strings.TrimSpace(langParam))
		// Check if it's a supported language
		for _, supported := range SupportedLanguages {
			if langCode == supported {
				return Language(langCode)
			}
		}
	}

	return DefaultLanguage
}
