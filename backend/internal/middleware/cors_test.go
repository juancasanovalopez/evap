package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCORS_AllowsWhitelistedOrigin(t *testing.T) {
	handler := CORS([]string{"https://evap.example.com"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("Origin", "https://evap.example.com")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, "https://evap.example.com", rec.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_RejectsNonWhitelistedOrigin(t *testing.T) {
	handler := CORS([]string{"https://evap.example.com"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_AllowsAcceptLanguageHeader(t *testing.T) {
	handler := CORS([]string{"https://evap.example.com"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/simulate", nil)
	req.Header.Set("Origin", "https://evap.example.com")
	req.Header.Set("Access-Control-Request-Method", http.MethodGet)
	req.Header.Set("Access-Control-Request-Headers", "Accept-Language")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Contains(t, rec.Header().Get("Access-Control-Allow-Headers"), "Accept-Language")
}
