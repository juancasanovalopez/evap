package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"evap-backend/internal/auth"
	"evap-backend/internal/i18n"
)

type contextKey string

const claimsContextKey contextKey = "auth-claims"

// JWTAuth validates a bearer token (Authorization header or "session" cookie)
// and injects the parsed claims into the request context. Requests without a
// valid token receive 401 Unauthorized.
func JWTAuth(issuer *auth.TokenIssuer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := bearerToken(r)
			if token == "" {
				writeLocalizedUnauthorized(w, r, i18n.AuthMissingCredentials)
				return
			}
			claims, err := issuer.Verify(token)
			if err != nil {
				writeLocalizedUnauthorized(w, r, i18n.AuthInvalidToken)
				return
			}
			ctx := context.WithValue(r.Context(), claimsContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func writeLocalizedUnauthorized(w http.ResponseWriter, r *http.Request, key string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": i18n.Text(LanguageFromContext(r.Context()), key),
	})
}

// ClaimsFromContext retrieves the authenticated user's claims, if present.
func ClaimsFromContext(ctx context.Context) (*auth.Claims, bool) {
	claims, ok := ctx.Value(claimsContextKey).(*auth.Claims)
	return claims, ok
}

// OptionalJWTAuth populates claims into the request context when a valid
// token is present, but never rejects the request when it is missing or
// invalid. Useful for pages that render differently for logged-in users.
func OptionalJWTAuth(issuer *auth.TokenIssuer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if token := bearerToken(r); token != "" {
				if claims, err := issuer.Verify(token); err == nil {
					r = r.WithContext(context.WithValue(r.Context(), claimsContextKey, claims))
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func bearerToken(r *http.Request) string {
	if header := r.Header.Get("Authorization"); header != "" {
		parts := strings.SplitN(header, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return strings.TrimSpace(parts[1])
		}
	}
	if cookie, err := r.Cookie("session"); err == nil {
		return cookie.Value
	}
	return ""
}
