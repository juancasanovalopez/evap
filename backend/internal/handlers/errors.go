package handlers

import (
	"net/http"

	"evap-backend/internal/i18n"
	"evap-backend/internal/middleware"
)

type errorResponse struct {
	Error string `json:"error"`
}

func writeLocalizedError(w http.ResponseWriter, r *http.Request, status int, key string, args ...any) {
	writeJSON(w, status, errorResponse{
		Error: i18n.Text(middleware.LanguageFromContext(r.Context()), key, args...),
	})
}
