package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"evap-backend/internal/auth"
	"evap-backend/internal/config"
	"evap-backend/internal/store"
)

func TestRouter_HealthIsPublic(t *testing.T) {
	r := New(Deps{
		Config: &config.Config{AllowedCORSOrigins: []string{"http://localhost"}},
		Users:  store.NewMemoryUserRepository(),
		Issuer: auth.NewTokenIssuer("k", time.Hour),
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestRouter_PrivateRequiresAuth(t *testing.T) {
	r := New(Deps{
		Config: &config.Config{AllowedCORSOrigins: []string{"http://localhost"}},
		Users:  store.NewMemoryUserRepository(),
		Issuer: auth.NewTokenIssuer("k", time.Hour),
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/private", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRouter_PrivateWithValidTokenSucceeds(t *testing.T) {
	issuer := auth.NewTokenIssuer("k", time.Hour)
	r := New(Deps{
		Config: &config.Config{AllowedCORSOrigins: []string{"http://localhost"}},
		Users:  store.NewMemoryUserRepository(),
		Issuer: issuer,
	})

	token, err := issuer.Issue("google#1", auth.Claims{Provider: "google", Email: "a@example.com"})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/private", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}
