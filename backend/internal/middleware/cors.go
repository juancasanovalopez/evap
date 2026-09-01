package middleware

import (
	"net/http"

	"github.com/go-chi/cors"
)

// CORS builds a CORS middleware restricted to an explicit origin whitelist.
// Credentials (cookies/Authorization) are allowed, so the wildcard "*" origin
// must never be used here.
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "Accept-Language"},
		AllowCredentials: true,
		MaxAge:           300,
	})
}
