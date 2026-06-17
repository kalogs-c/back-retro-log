package app

import (
	"io/fs"
	"net/http"
	"strings"

	"back-retro-log/internal/auth"
	"back-retro-log/internal/db"
	"back-retro-log/internal/handlers"
	"back-retro-log/internal/i18n"
	"back-retro-log/internal/providers"
)

func NewRouter(
	queries *db.Queries,
	sessions *auth.SessionManager,
	provider providers.GameProvider,
	baseURL string,
	staticFS fs.FS,
	localeStore *i18n.Store,
) http.Handler {
	authH := &handlers.AuthHandler{
		Queries:  queries,
		Sessions: sessions,
		BaseURL:  baseURL,
	}
	catalogH := &handlers.CatalogHandler{
		Queries: queries,
	}
	searchH := &handlers.SearchHandler{
		Provider: provider,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("GET /login", authH.LoginPage)
	mux.HandleFunc("POST /login", authH.Login)
	mux.HandleFunc("GET /register", authH.RegisterPage)
	mux.HandleFunc("POST /register", authH.Register)

	mux.HandleFunc("GET /lang", func(w http.ResponseWriter, r *http.Request) {
		lang := r.URL.Query().Get("l")
		if lang == "" {
			lang = "pt-BR"
		}
		http.SetCookie(w, &http.Cookie{
			Name:   "lang",
			Value:  lang,
			Path:   "/",
			MaxAge: 86400 * 365,
		})
		redirect := r.URL.Query().Get("redirect")
		if redirect == "" {
			redirect = "/"
		}
		http.Redirect(w, r, redirect, http.StatusSeeOther)
	})

	sub, _ := fs.Sub(staticFS, "static")
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(sub))))

	protected := http.NewServeMux()
	protected.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/catalog", http.StatusSeeOther)
	})
	protected.HandleFunc("GET /logout", authH.Logout)
	protected.HandleFunc("POST /logout", authH.Logout)
	protected.HandleFunc("GET /catalog", catalogH.List)
	protected.HandleFunc("GET /catalog/", catalogH.List)
	protected.HandleFunc("GET /catalog/search", catalogH.Search)
	protected.HandleFunc("POST /catalog/add", catalogH.Add)
	protected.HandleFunc("PUT /catalog/{id}/status", catalogH.UpdateStatus)
	protected.HandleFunc("DELETE /catalog/{id}", catalogH.Delete)
	protected.HandleFunc("GET /search", searchH.Page)
	protected.HandleFunc("GET /search/results", searchH.Results)
	protected.HandleFunc("GET /invite", authH.InvitePage)
	protected.HandleFunc("POST /invite", authH.CreateInvite)
	protected.HandleFunc("POST /invite/copy", authH.InviteCopy)

	mux.Handle("/", AuthMiddleware(queries, sessions)(protected))

	return LocaleMiddleware(localeStore)(mux)
}

func LocaleMiddleware(store *i18n.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			lang := r.URL.Query().Get("lang")
			if lang == "" {
				if c, err := r.Cookie("lang"); err == nil {
					lang = c.Value
				}
			}
			if lang == "" {
				lang = negotiateLang(r.Header.Get("Accept-Language"))
			}
			loc := store.Get(lang)
			ctx := i18n.ToContext(r.Context(), loc)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func negotiateLang(header string) string {
	if strings.Contains(header, "pt") {
		return "pt-BR"
	}
	return "en"
}
