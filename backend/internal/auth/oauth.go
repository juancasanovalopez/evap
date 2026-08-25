package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

// Provider identifies a supported OAuth2 identity provider.
type Provider string

const (
	ProviderGoogle Provider = "google"
	ProviderGitHub Provider = "github"
)

// Profile is the normalized identity data returned by a provider.
type Profile struct {
	ID        string
	Email     string
	Name      string
	AvatarURL string
}

// ProviderConfig bundles an oauth2.Config with a function to fetch the
// authenticated user's profile once an access token has been obtained.
type ProviderConfig struct {
	OAuth2       *oauth2.Config
	FetchProfile func(ctx context.Context, client *http.Client) (Profile, error)
}

// NewProviders builds the supported OAuth2 provider configurations.
func NewProviders(redirectBaseURL, googleClientID, googleClientSecret, githubClientID, githubClientSecret string) map[Provider]*ProviderConfig {
	return map[Provider]*ProviderConfig{
		ProviderGoogle: {
			OAuth2: &oauth2.Config{
				ClientID:     googleClientID,
				ClientSecret: googleClientSecret,
				Endpoint:     google.Endpoint,
				RedirectURL:  redirectBaseURL + "/auth/google/callback",
				Scopes:       []string{"openid", "email", "profile"},
			},
			FetchProfile: fetchGoogleProfile,
		},
		ProviderGitHub: {
			OAuth2: &oauth2.Config{
				ClientID:     githubClientID,
				ClientSecret: githubClientSecret,
				Endpoint:     github.Endpoint,
				RedirectURL:  redirectBaseURL + "/auth/github/callback",
				Scopes:       []string{"read:user", "user:email"},
			},
			FetchProfile: fetchGitHubProfile,
		},
	}
}

// NewState generates a random, URL-safe CSRF state value for the OAuth2 flow.
func NewState() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("auth: generating state: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// Provider endpoint URLs, overridable in tests.
var (
	googleUserInfoURL   = "https://openidconnect.googleapis.com/v1/userinfo"
	githubUserURL       = "https://api.github.com/user"
	githubUserEmailsURL = "https://api.github.com/user/emails"
)

func fetchGoogleProfile(ctx context.Context, client *http.Client) (Profile, error) {
	var body struct {
		ID      string `json:"sub"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := getJSON(ctx, client, googleUserInfoURL, &body); err != nil {
		return Profile{}, err
	}
	return Profile{ID: body.ID, Email: body.Email, Name: body.Name, AvatarURL: body.Picture}, nil
}

func fetchGitHubProfile(ctx context.Context, client *http.Client) (Profile, error) {
	var body struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := getJSON(ctx, client, githubUserURL, &body); err != nil {
		return Profile{}, err
	}
	name := body.Name
	if name == "" {
		name = body.Login
	}
	email := body.Email
	if email == "" {
		// GitHub only returns a public email if the user has one; fall back to
		// the dedicated emails endpoint for the primary verified address.
		if e, err := fetchGitHubPrimaryEmail(ctx, client); err == nil {
			email = e
		}
	}
	return Profile{ID: fmt.Sprintf("%d", body.ID), Email: email, Name: name, AvatarURL: body.AvatarURL}, nil
}

func fetchGitHubPrimaryEmail(ctx context.Context, client *http.Client) (string, error) {
	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := getJSON(ctx, client, githubUserEmailsURL, &emails); err != nil {
		return "", err
	}
	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}
	return "", fmt.Errorf("auth: no verified primary email found")
}

func getJSON(ctx context.Context, client *http.Client, url string, dest interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("auth: unexpected status %d fetching %s: %s", resp.StatusCode, url, string(body))
	}
	return json.NewDecoder(resp.Body).Decode(dest)
}
