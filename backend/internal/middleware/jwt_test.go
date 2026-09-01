package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"evap-backend/internal/auth"
)

func TestJWTAuth_RejectsMissingToken(t *testing.T) {
	issuer := auth.NewTokenIssuer("key", time.Hour)
	handler := JWTAuth(issuer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/private", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestJWTAuth_LocalizesMissingTokenError(t *testing.T) {
	issuer := auth.NewTokenIssuer("key", time.Hour)
	handler := DetectLanguage(JWTAuth(issuer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/private", nil)
	req.Header.Set("Accept-Language", "fr-CA")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var response map[string]string
	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	require.Equal(t, "autorisation manquante ou incorrecte", response["error"])
}

func TestJWTAuth_AcceptsValidBearerToken(t *testing.T) {
	issuer := auth.NewTokenIssuer("key", time.Hour)
	token, err := issuer.Issue("google#1", auth.Claims{Provider: "google", Email: "a@example.com"})
	require.NoError(t, err)

	var gotEmail string
	handler := JWTAuth(issuer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := ClaimsFromContext(r.Context())
		require.True(t, ok)
		gotEmail = claims.Email
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/private", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "a@example.com", gotEmail)
}

func TestJWTAuth_AcceptsSessionCookie(t *testing.T) {
	issuer := auth.NewTokenIssuer("key", time.Hour)
	token, err := issuer.Issue("github#1", auth.Claims{Provider: "github"})
	require.NoError(t, err)

	handler := JWTAuth(issuer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/private", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestOptionalJWTAuth_DoesNotRejectMissingToken(t *testing.T) {
	issuer := auth.NewTokenIssuer("key", time.Hour)
	handler := OptionalJWTAuth(issuer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, ok := ClaimsFromContext(r.Context())
		require.False(t, ok)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}
