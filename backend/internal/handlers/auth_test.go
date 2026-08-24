package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"evap-backend/internal/auth"
	"evap-backend/internal/store"
)

func requestWithProvider(method, target, provider string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("provider", provider)
	req := httptest.NewRequest(method, target, nil)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func TestAuthHandler_Login_RedirectsAndSetsStateCookie(t *testing.T) {
	providers := auth.NewProviders("http://localhost", "gid", "gsecret", "hid", "hsecret")
	h := &AuthHandler{Providers: providers, Issuer: auth.NewTokenIssuer("k", time.Hour), Users: store.NewMemoryUserRepository()}

	req := requestWithProvider(http.MethodGet, "/auth/google/login", "google")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	require.Equal(t, http.StatusFound, rec.Code)
	require.NotEmpty(t, rec.Header().Get("Location"))
	require.Len(t, rec.Result().Cookies(), 1)
	require.Equal(t, oauthStateCookie, rec.Result().Cookies()[0].Name)
}

func TestAuthHandler_Login_UnknownProvider(t *testing.T) {
	h := &AuthHandler{Providers: map[auth.Provider]*auth.ProviderConfig{}}

	req := requestWithProvider(http.MethodGet, "/auth/unknown/login", "unknown")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAuthHandler_Callback_IssuesSessionAndUpsertsUser(t *testing.T) {
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"fake-token","token_type":"bearer"}`))
	}))
	defer tokenServer.Close()

	users := store.NewMemoryUserRepository()
	h := &AuthHandler{
		Providers: map[auth.Provider]*auth.ProviderConfig{
			auth.ProviderGoogle: {
				OAuth2: &oauth2.Config{
					ClientID:     "gid",
					ClientSecret: "gsecret",
					Endpoint:     oauth2.Endpoint{TokenURL: tokenServer.URL},
				},
				FetchProfile: func(ctx context.Context, client *http.Client) (auth.Profile, error) {
					return auth.Profile{ID: "1", Email: "a@example.com", Name: "Ada"}, nil
				},
			},
		},
		Issuer: auth.NewTokenIssuer("k", time.Hour),
		Users:  users,
	}

	req := requestWithProvider(http.MethodGet, "/auth/google/callback?state=abc&code=xyz", "google")
	req.AddCookie(&http.Cookie{Name: oauthStateCookie, Value: "abc"})
	rec := httptest.NewRecorder()

	h.Callback(rec, req)

	require.Equal(t, http.StatusFound, rec.Code)

	var sessionCookie *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == "session" {
			sessionCookie = c
		}
	}
	require.NotNil(t, sessionCookie)

	claims, err := h.Issuer.Verify(sessionCookie.Value)
	require.NoError(t, err)
	require.Equal(t, "a@example.com", claims.Email)

	stored, err := users.Get(req.Context(), "google", "1")
	require.NoError(t, err)
	require.Equal(t, "Ada", stored.Name)
}

func TestAuthHandler_Callback_RejectsInvalidState(t *testing.T) {
	h := &AuthHandler{
		Providers: map[auth.Provider]*auth.ProviderConfig{
			auth.ProviderGoogle: {OAuth2: &oauth2.Config{}},
		},
		Issuer: auth.NewTokenIssuer("k", time.Hour),
		Users:  store.NewMemoryUserRepository(),
	}

	req := requestWithProvider(http.MethodGet, "/auth/google/callback?state=wrong&code=xyz", "google")
	req.AddCookie(&http.Cookie{Name: oauthStateCookie, Value: "expected"})
	rec := httptest.NewRecorder()

	h.Callback(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}
