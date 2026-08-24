package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"evap-backend/internal/auth"
	appmw "evap-backend/internal/middleware"
)

func TestPrivate_ReturnsAuthenticatedUser(t *testing.T) {
	issuer := auth.NewTokenIssuer("key", time.Hour)
	token, err := issuer.Issue("google#1", auth.Claims{Provider: "google", Email: "a@example.com", Name: "Ada"})
	require.NoError(t, err)

	handler := appmw.JWTAuth(issuer)(http.HandlerFunc(Private))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/private", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, `{"provider":"google","email":"a@example.com","name":"Ada","avatar_url":""}`, rec.Body.String())
}

func TestPrivate_RejectsMissingToken(t *testing.T) {
	issuer := auth.NewTokenIssuer("key", time.Hour)
	handler := appmw.JWTAuth(issuer)(http.HandlerFunc(Private))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/private", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}
