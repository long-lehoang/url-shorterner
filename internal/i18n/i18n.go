// Package i18n provides internationalization support for the application.
package i18n

import (
	"embed"
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed *.toml
var translationsFS embed.FS

var (
	// bundle holds all translations
	bundle *i18n.Bundle
)

func init() {
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	// Load translation files
	loadTranslations()
}

// loadTranslations loads all translation files from the embedded filesystem.
func loadTranslations() {
	entries, err := translationsFS.ReadDir(".")
	if err != nil {
		panic(fmt.Sprintf("failed to read translation directory: %v", err))
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		data, err := translationsFS.ReadFile(entry.Name())
		if err != nil {
			panic(fmt.Sprintf("failed to read translation file %s: %v", entry.Name(), err))
		}

		_, err = bundle.ParseMessageFileBytes(data, entry.Name())
		if err != nil {
			panic(fmt.Sprintf("failed to parse translation file %s: %v", entry.Name(), err))
		}
	}
}

// getLocalizer returns a localizer for the given language tags.
// If no tags are provided, returns a localizer for the default language.
func getLocalizer(langTags ...string) *i18n.Localizer {
	if len(langTags) == 0 {
		return i18n.NewLocalizer(bundle, language.English.String())
	}
	return i18n.NewLocalizer(bundle, langTags...)
}

// T translates a message ID with optional template data.
func T(lang string, messageID string, templateData map[string]interface{}) string {
	loc := getLocalizer(lang)
	msg, err := loc.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: templateData,
	})
	if err != nil {
		// Fallback to English if translation fails
		enLoc := getLocalizer(language.English.String())
		msg, _ = enLoc.Localize(&i18n.LocalizeConfig{
			MessageID:    messageID,
			TemplateData: templateData,
		})
		if msg == "" {
			return messageID
		}
	}
	return msg
}
