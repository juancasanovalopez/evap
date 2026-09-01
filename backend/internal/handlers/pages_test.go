package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"evap-backend/internal/auth"
	appmw "evap-backend/internal/middleware"
)

func TestIndex_AnonymousRedirectsToLogin(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	Index(rec, req)

	require.Equal(t, http.StatusFound, rec.Code)
	require.Equal(t, "/login", rec.Header().Get("Location"))
}

func TestIndex_AuthenticatedRendersDashboard(t *testing.T) {
	issuer := auth.NewTokenIssuer("key", time.Hour)
	token, err := issuer.Issue("google#1", auth.Claims{Provider: "google", Name: "Ada"})
	require.NoError(t, err)

	handler := appmw.JWTAuth(issuer)(http.HandlerFunc(Index))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "Ada")
}

func TestLogin_RendersProviderLinks(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	rec := httptest.NewRecorder()

	Login(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	require.True(t, strings.Contains(body, "/auth/google/login"))
	require.True(t, strings.Contains(body, "/auth/github/login"))
}

func TestLogin_RendersNegotiatedFrench(t *testing.T) {
	handler := appmw.DetectLanguage(http.HandlerFunc(Login))
	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	req.Header.Set("Accept-Language", "fr-CA")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `<html lang="fr-u-rg-cazzzz">`)
	require.Contains(t, rec.Body.String(), "Se connecter avec Google")
}
