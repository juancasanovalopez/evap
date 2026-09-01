package handlers

import (
	"html/template"
	"net/http"

	"evap-backend/internal/i18n"
	"evap-backend/internal/middleware"
	"evap-backend/web"
)

type indexPageData struct {
	Name     string
	Language string
}

type loginPageData struct {
	Language string
}

func pageTemplateFor(r *http.Request) *template.Template {
	tag := middleware.LanguageFromContext(r.Context())
	return template.Must(template.New("pages").Funcs(template.FuncMap{
		"msg": func(key string, args ...any) string {
			return i18n.Text(tag, key, args...)
		},
	}).ParseFS(web.Templates, "templates/*.html"))
}

// Index renders the SSR dashboard, requiring an authenticated session.
// Unauthenticated visitors are redirected to /login.
func Index(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	data := indexPageData{Name: claims.Name, Language: middleware.LanguageFromContext(r.Context()).String()}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := pageTemplateFor(r).ExecuteTemplate(w, "index.html", data); err != nil {
		writeLocalizedError(w, r, http.StatusInternalServerError, i18n.PageRenderFailed)
	}
}

// Login renders the SSR login page with buttons for each OAuth2 provider.
func Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := loginPageData{Language: middleware.LanguageFromContext(r.Context()).String()}
	if err := pageTemplateFor(r).ExecuteTemplate(w, "login.html", data); err != nil {
		writeLocalizedError(w, r, http.StatusInternalServerError, i18n.PageRenderFailed)
	}
}
