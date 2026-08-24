package handlers

import (
	"net/http"

	"evap-backend/internal/middleware"
)

// PrivateUserResponse is returned by the protected /api/v1/private endpoint.
type PrivateUserResponse struct {
	Provider  string `json:"provider"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

// Private returns the authenticated user's information. It must be mounted
// behind middleware.JWTAuth so claims are always present in the context.
func Private(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	writeJSON(w, http.StatusOK, PrivateUserResponse{
		Provider:  claims.Provider,
		Email:     claims.Email,
		Name:      claims.Name,
		AvatarURL: claims.AvatarURL,
	})
}
