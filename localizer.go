package localizer

import (
	"fmt"
	"sync"
)

// Localizer provides internationalization services with thread-safe operation
// and complete compatibility with existing spotsbook microservice contracts.
type Localizer struct {
	mu              sync.RWMutex
	Locale          string
	FallbackLocale  string
	LocalizationSrc []string
	Localizations   map[string]string
	store           *Store
}

// New creates a new initialized Localizer instance.
func New(sourceDirs ...string) *Localizer {
	l := &Localizer{
		Locale:          "en",
		FallbackLocale:  "en",
		LocalizationSrc: sourceDirs,
		Localizations:   make(map[string]string),
		store:           NewStore(),
	}
	return l
}

// SetSource configures the source directories for translation assets.
func (l *Localizer) SetSource(src ...string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.LocalizationSrc = src
}

// SetDefaultLocale sets the active and fallback locales.
func (l *Localizer) SetDefaultLocale(locale, fallback string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if locale != "" {
		l.Locale = locale
	}
	if fallback != "" {
		l.FallbackLocale = fallback
	}
}

// SetLocales sets the active and fallback locales, returning the receiver for chaining.
func (l *Localizer) SetLocales(locale, fallback string) *Localizer {
	l.SetDefaultLocale(locale, fallback)
	return l
}

// SetLocale sets the active locale.
func (l *Localizer) SetLocale(locale string) *Localizer {
	l.mu.Lock()
	defer l.mu.Unlock()
	if locale != "" {
		l.Locale = locale
	}
	return l
}

// SetFallbackLocale sets the fallback locale.
func (l *Localizer) SetFallbackLocale(fallback string) *Localizer {
	l.mu.Lock()
	defer l.mu.Unlock()
	if fallback != "" {
		l.FallbackLocale = fallback
	}
	return l
}

// SetupLocalization loads all translation files from the configured directories.
func (l *Localizer) SetupLocalization() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.store == nil {
		l.store = NewStore()
	}

	for _, dir := range l.LocalizationSrc {
		_ = l.store.LoadDirectory(dir)
	}

	l.Localizations = l.store.Snapshot()
}

// Translate resolves a localization key, applying variable substitutions if provided.
func (l *Localizer) Translate(key string, replacements ...map[string]string) string {
	l.mu.RLock()
	activeLocale := l.Locale
	fallbackLocale := l.FallbackLocale
	if l.store == nil {
		l.mu.RUnlock()
		return key
	}
	l.mu.RUnlock()

	lookupKey := fmt.Sprintf("%s.%s", activeLocale, key)
	pattern, found := l.store.Get(lookupKey)
	if !found && fallbackLocale != "" && fallbackLocale != activeLocale {
		fallbackKey := fmt.Sprintf("%s.%s", fallbackLocale, key)
		pattern, found = l.store.Get(fallbackKey)
	}

	if !found {
		return key
	}

	if len(replacements) == 0 {
		return pattern
	}

	merged := make(map[string]string)
	for _, m := range replacements {
		for k, v := range m {
			merged[k] = v
		}
	}

	return Interpolate(pattern, merged)
}

// T translates a key directly for a specific locale without mutating the receiver's state.
func (l *Localizer) T(locale, key string, replacements ...map[string]string) string {
	l.mu.RLock()
	fallbackLocale := l.FallbackLocale
	if l.store == nil {
		l.mu.RUnlock()
		return key
	}
	l.mu.RUnlock()

	lookupKey := fmt.Sprintf("%s.%s", locale, key)
	pattern, found := l.store.Get(lookupKey)
	if !found && fallbackLocale != "" && fallbackLocale != locale {
		pattern, found = l.store.Get(fmt.Sprintf("%s.%s", fallbackLocale, key))
	}

	if !found {
		return key
	}

	if len(replacements) == 0 {
		return pattern
	}

	merged := make(map[string]string)
	for _, m := range replacements {
		for k, v := range m {
			merged[k] = v
		}
	}

	return Interpolate(pattern, merged)
}
