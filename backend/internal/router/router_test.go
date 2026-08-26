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

func TestRouter_SimulateRequiresAuth(t *testing.T) {
	r := New(Deps{
		Config: &config.Config{AllowedCORSOrigins: []string{"http://localhost"}},
		Users:  store.NewMemoryUserRepository(),
		Issuer: auth.NewTokenIssuer("k", time.Hour),
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/simulate?area=32&profundidad=1.2&lat=40.4167&lon=-3.7037&fecha_inicio=2025-07-15&fecha_fin=2025-07-17", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRouter_IndexRedirectsAnonymousToLogin(t *testing.T) {
	r := New(Deps{
		Config: &config.Config{AllowedCORSOrigins: []string{"http://localhost"}},
		Users:  store.NewMemoryUserRepository(),
		Issuer: auth.NewTokenIssuer("k", time.Hour),
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusFound, rec.Code)
	require.Equal(t, "/login", rec.Header().Get("Location"))
}

func TestRouter_ServesVendoredStaticAssets(t *testing.T) {
	r := New(Deps{
		Config: &config.Config{AllowedCORSOrigins: []string{"http://localhost"}},
		Users:  store.NewMemoryUserRepository(),
		Issuer: auth.NewTokenIssuer("k", time.Hour),
	})

	for _, path := range []string{
		"/static/vendor/leaflet/leaflet.js",
		"/static/vendor/leaflet/leaflet.css",
		"/static/vendor/chartjs/chart.umd.min.js",
		"/static/js/dashboard.js",
	} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equalf(t, http.StatusOK, rec.Code, "expected %s to be served", path)
	}
}
