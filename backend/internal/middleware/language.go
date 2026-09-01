package middleware

import (
	"context"
	"net/http"

	"golang.org/x/text/language"
)

const languageContextKey contextKey = "language-tag"

var supportedLanguages = []language.Tag{
	language.English,
	language.Spanish,
	language.French,
}

var languageMatcher = language.NewMatcher(supportedLanguages)

// DetectLanguage negotiates the request's Accept-Language header against the
// languages supported by the server. English is the default for absent,
// malformed, and unsupported headers.
func DetectLanguage(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tag := language.English
		if requested, _, err := language.ParseAcceptLanguage(r.Header.Get("Accept-Language")); err == nil && len(requested) > 0 {
			matched, _, _ := languageMatcher.Match(requested...)
			tag = matched
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), languageContextKey, tag)))
	})
}

// LanguageFromContext returns the request language, defaulting to English when
// a caller is outside the HTTP middleware chain.
func LanguageFromContext(ctx context.Context) language.Tag {
	if tag, ok := ctx.Value(languageContextKey).(language.Tag); ok {
		return tag
	}
	return language.English
}
