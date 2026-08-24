package handlers

import (
	"html/template"
	"net/http"

	"evap-backend/internal/middleware"
	"evap-backend/web"
)

var pageTemplates = template.Must(template.ParseFS(web.Templates, "templates/*.html"))

type indexPageData struct {
	Authenticated bool
	Name          string
}

// Index renders the SSR home page, greeting the user if they have a valid
// session cookie. Mount behind middleware.OptionalJWTAuth so claims are
// populated when present, without requiring authentication.
func Index(w http.ResponseWriter, r *http.Request) {
	data := indexPageData{}
	if claims, ok := middleware.ClaimsFromContext(r.Context()); ok {
		data.Authenticated = true
		data.Name = claims.Name
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := pageTemplates.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, "failed to render page", http.StatusInternalServerError)
	}
}

// Login renders the SSR login page with buttons for each OAuth2 provider.
func Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := pageTemplates.ExecuteTemplate(w, "login.html", nil); err != nil {
		http.Error(w, "failed to render page", http.StatusInternalServerError)
	}
}
