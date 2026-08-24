package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"evap-backend/internal/auth"
	"evap-backend/internal/store"
)

const oauthStateCookie = "oauth_state"

// AuthHandler wires OAuth2 login/callback routes to the token issuer and
// user repository.
type AuthHandler struct {
	Providers     map[auth.Provider]*auth.ProviderConfig
	Issuer        *auth.TokenIssuer
	Users         store.UserRepository
	SecureCookies bool
}

// Login redirects the user to the provider's OAuth2 consent screen.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	provider := auth.Provider(chi.URLParam(r, "provider"))
	cfg, ok := h.Providers[provider]
	if !ok {
		http.Error(w, "unknown provider", http.StatusNotFound)
		return
	}

	state, err := auth.NewState()
	if err != nil {
		http.Error(w, "failed to start login", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookie,
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.SecureCookies,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((5 * time.Minute).Seconds()),
	})

	http.Redirect(w, r, cfg.OAuth2.AuthCodeURL(state), http.StatusFound)
}

// Callback exchanges the authorization code for a token, fetches the user's
// profile, upserts it, and issues an application JWT as an HttpOnly cookie.
func (h *AuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	provider := auth.Provider(chi.URLParam(r, "provider"))
	cfg, ok := h.Providers[provider]
	if !ok {
		http.Error(w, "unknown provider", http.StatusNotFound)
		return
	}

	stateCookie, err := r.Cookie(oauthStateCookie)
	if err != nil || r.URL.Query().Get("state") != stateCookie.Value {
		http.Error(w, "invalid oauth state", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing authorization code", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	token, err := cfg.OAuth2.Exchange(ctx, code)
	if err != nil {
		http.Error(w, "failed to exchange authorization code", http.StatusBadGateway)
		return
	}

	client := cfg.OAuth2.Client(ctx, token)
	profile, err := cfg.FetchProfile(ctx, client)
	if err != nil {
		http.Error(w, "failed to fetch user profile", http.StatusBadGateway)
		return
	}

	user, err := h.Users.Upsert(ctx, store.User{
		Provider:   string(provider),
		ProviderID: profile.ID,
		Email:      profile.Email,
		Name:       profile.Name,
		AvatarURL:  profile.AvatarURL,
	})
	if err != nil {
		http.Error(w, "failed to persist user profile", http.StatusInternalServerError)
		return
	}

	jwtToken, err := h.Issuer.Issue(string(provider)+"#"+user.ProviderID, auth.Claims{
		Provider:  user.Provider,
		Email:     user.Email,
		Name:      user.Name,
		AvatarURL: user.AvatarURL,
	})
	if err != nil {
		http.Error(w, "failed to issue session token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    jwtToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.SecureCookies,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((24 * time.Hour).Seconds()),
	})
	http.SetCookie(w, &http.Cookie{
		Name:   oauthStateCookie,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	http.Redirect(w, r, "/", http.StatusFound)
}
