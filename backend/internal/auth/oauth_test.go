package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFetchGoogleProfile(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"sub":"g-1","email":"a@example.com","name":"Ada","picture":"http://x/pic.png"}`))
	}))
	defer srv.Close()

	original := googleUserInfoURL
	googleUserInfoURL = srv.URL
	defer func() { googleUserInfoURL = original }()

	profile, err := fetchGoogleProfile(context.Background(), srv.Client())
	require.NoError(t, err)
	require.Equal(t, "g-1", profile.ID)
	require.Equal(t, "a@example.com", profile.Email)
	require.Equal(t, "Ada", profile.Name)
	require.Equal(t, "http://x/pic.png", profile.AvatarURL)
}

func TestFetchGitHubProfile_FallsBackToPrimaryVerifiedEmail(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":42,"login":"octocat","name":"","avatar_url":"http://x/a.png"}`))
	})
	mux.HandleFunc("/user/emails", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"email":"secondary@example.com","primary":false,"verified":true},{"email":"primary@example.com","primary":true,"verified":true}]`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	origUser, origEmails := githubUserURL, githubUserEmailsURL
	githubUserURL = srv.URL + "/user"
	githubUserEmailsURL = srv.URL + "/user/emails"
	defer func() { githubUserURL, githubUserEmailsURL = origUser, origEmails }()

	profile, err := fetchGitHubProfile(context.Background(), srv.Client())
	require.NoError(t, err)
	require.Equal(t, "42", profile.ID)
	require.Equal(t, "octocat", profile.Name)
	require.Equal(t, "primary@example.com", profile.Email)
}

func TestNewState_IsUniqueAndURLSafe(t *testing.T) {
	a, err := NewState()
	require.NoError(t, err)
	b, err := NewState()
	require.NoError(t, err)
	require.NotEmpty(t, a)
	require.NotEqual(t, a, b)
}
