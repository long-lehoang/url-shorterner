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
			switch len(args) {
			case 1:
				templateData["Resource"] = args[0]
			case 2:
				templateData["Resource"] = args[0]
				templateData["Message"] = args[1]
			default:
				// For more than 2 args, only use first two
				if len(args) > 0 {
					templateData["Resource"] = args[0]
				}
				if len(args) > 1 {
					templateData["Message"] = args[1]
				}
			}
		}
	}

	return i18n.T(string(lang), string(code), templateData)
}
