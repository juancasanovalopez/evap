package handlers

import (
	"net/http"
	"strconv"
	"time"

	"evap-backend/internal/i18n"
	"evap-backend/internal/middleware"
	"evap-backend/internal/store"
)

const (
	defaultReadingsLimit = 20
	maxReadingsLimit     = 100
	// readingsMaxAge caps diagnostics to the last 24h, regardless of limit.
	readingsMaxAge = 24 * time.Hour
)

// ReadingsHandler returns the authenticated user's most recent sensor
// readings, scoped by owner_user_id so one user never sees another's data.
func ReadingsHandler(repo store.ReadingRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := middleware.ClaimsFromContext(r.Context())
		if !ok {
			writeLocalizedError(w, r, http.StatusUnauthorized, i18n.AuthUnauthorized)
			return
		}

		limit := int32(defaultReadingsLimit)
		if raw := r.URL.Query().Get("limit"); raw != "" {
			if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 && parsed <= maxReadingsLimit {
				limit = int32(parsed)
			}
		}

		ownerID := ownerSlug(claims.Subject)
		since := time.Now().Add(-readingsMaxAge)
		readings, err := repo.ListRecentByOwner(r.Context(), ownerID, limit, since)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "could not fetch sensor readings"})
			return
		}
		if readings == nil {
			readings = []store.Reading{}
		}
		writeJSON(w, http.StatusOK, readings)
	}
}
