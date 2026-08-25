// Package router builds the chi HTTP router used both by the Lambda
// entrypoint and by local/integration tests via httptest.
package router

import (
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"evap-backend/internal/auth"
	"evap-backend/internal/config"
	"evap-backend/internal/handlers"
	appmw "evap-backend/internal/middleware"
	"evap-backend/internal/store"
)

// Deps holds the dependencies required to build the router.
type Deps struct {
	Config *config.Config
	Users  store.UserRepository
	Issuer *auth.TokenIssuer
	// SecureCookies should be true in production (HTTPS); false for local dev over HTTP.
	SecureCookies bool
}

// New builds the fully-wired chi router for the application.
func New(deps Deps) *chi.Mux {
	providers := auth.NewProviders(
		deps.Config.OAuthRedirectBaseURL,
		deps.Config.GoogleClientID, deps.Config.GoogleClientSecret,
		deps.Config.GitHubClientID, deps.Config.GitHubClientSecret,
	)

	authHandler := &handlers.AuthHandler{
		Providers:     providers,
		Issuer:        deps.Issuer,
		Users:         deps.Users,
		SecureCookies: deps.SecureCookies,
	}

	r := chi.NewRouter()
	r.Use(chimw.RequestID, chimw.RealIP, chimw.Recoverer, chimw.Logger)
	r.Use(appmw.CORS(deps.Config.AllowedCORSOrigins))
	r.Use(appmw.RateLimit(5, 10)) // 5 req/s sustained, burst of 10, per client IP.

	r.Get("/health", handlers.Health)

	r.Group(func(r chi.Router) {
		r.Use(appmw.OptionalJWTAuth(deps.Issuer))
		r.Get("/", handlers.Index)
	})
	r.Get("/login", handlers.Login)

	r.Route("/auth/{provider}", func(r chi.Router) {
		r.Get("/login", authHandler.Login)
		r.Get("/callback", authHandler.Callback)
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(appmw.JWTAuth(deps.Issuer))
		r.Get("/private", handlers.Private)
		r.Get("/simulate", handlers.Simulate)
	})

	return r
}

// DefaultJWTTTL is the lifetime of application-issued session tokens.
const DefaultJWTTTL = 24 * time.Hour
