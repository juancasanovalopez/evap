package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"
)

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   language.Tag
	}{
		{name: "missing header defaults to English", want: language.English},
		{name: "invalid header defaults to English", header: "not a language tag!", want: language.English},
		{name: "unsupported header defaults to English", header: "de-DE", want: language.English},
		{name: "Spanish regional variant", header: "es-MX", want: language.Spanish},
		{name: "French regional variant", header: "fr-CA,es;q=0.8,en;q=0.5", want: language.French},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := DetectLanguage(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				matched := LanguageFromContext(r.Context())
				matchedBase, _ := matched.Base()
				wantedBase, _ := test.want.Base()
				require.Equal(t, wantedBase, matchedBase)
				w.WriteHeader(http.StatusNoContent)
			}))
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Accept-Language", test.header)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)
			require.Equal(t, http.StatusNoContent, rec.Code)
		})
	}
}
