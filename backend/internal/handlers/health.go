// Package handlers implements HTTP handlers for the evap API and SSR pages.
package handlers

import (
	"encoding/json"
	"net/http"
)

// Health reports liveness for smoke tests / uptime checks.
func Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
