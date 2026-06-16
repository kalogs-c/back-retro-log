package app

import (
	"io/fs"
	"net/http"

	"back-retro-log/internal/auth"
	"back-retro-log/internal/db"
	"back-retro-log/internal/handlers"
	"back-retro-log/internal/providers"
)

func NewRouter(
	queries *db.Queries,
	sessions *auth.SessionManager,
	provider providers.GameProvider,
	baseURL string,
	staticFS fs.FS,
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

	mux.Handle("/", AuthMiddleware(queries, sessions)(protected))

	return mux
}
