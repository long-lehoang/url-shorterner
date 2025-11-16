// Package errors provides common error types used across the application.
package errors

import (
	"url-shorterner/internal/i18n"
)

// Language is an alias for i18n.Language for convenience.
type Language = i18n.Language

// Language constants are re-exported from i18n package.
const (
	LanguageEN      = i18n.LanguageEN
	LanguageVI      = i18n.LanguageVI
	DefaultLanguage = i18n.DefaultLanguage
)

// GetMessage returns the error message for a given error code and language using i18n.
// If the language is not supported, it falls back to the default language.
func GetMessage(code ErrorCode, lang Language, args ...interface{}) string {
	templateData := make(map[string]interface{})

	// Handle different argument types
	if len(args) > 0 {
		// If first arg is a map, use it as template data
		if data, ok := args[0].(map[string]interface{}); ok {
			templateData = data
		} else if len(args) == 1 {
			// Single string argument (e.g., resource name)
			templateData["Resource"] = args[0]
		} else {
			// Multiple arguments - try to map common patterns
			for i, arg := range args {
				if i == 0 {
					templateData["Resource"] = arg
				} else if i == 1 {
					templateData["Message"] = arg
				}
			}
		}
	}

	return i18n.T(string(lang), string(code), templateData)
}
