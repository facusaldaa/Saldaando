package i18n

import (
	"fmt"
	"sync"
)

// Language represents a language code
type Language string

const (
	LanguageEnglish Language = "en"
	LanguageSpanishAR Language = "es_AR"
)

// Translator handles translations
type Translator struct {
	language Language
	mu       sync.RWMutex
}

// NewTranslator creates a new translator with default language
func NewTranslator(lang Language) *Translator {
	if lang == "" {
		lang = LanguageEnglish
	}
	return &Translator{language: lang}
}

// SetLanguage sets the current language
func (t *Translator) SetLanguage(lang Language) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.language = lang
}

// GetLanguage returns the current language
func (t *Translator) GetLanguage() Language {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.language
}

// T translates a key with optional arguments
func (t *Translator) T(key string, args ...interface{}) string {
	t.mu.RLock()
	lang := t.language
	t.mu.RUnlock()

	translations := getTranslations(lang)
	text, ok := translations[key]
	if !ok {
		// Fallback to English
		enTranslations := getTranslations(LanguageEnglish)
		text, ok = enTranslations[key]
		if !ok {
			return key // Return key if translation not found
		}
	}

	if len(args) > 0 {
		return fmt.Sprintf(text, args...)
	}
	return text
}

// getTranslations returns translations for a language
func getTranslations(lang Language) map[string]string {
	switch lang {
	case LanguageSpanishAR:
		return spanishARTranslations
	default:
		return englishTranslations
	}
}

