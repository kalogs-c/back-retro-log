package app

import (
	"context"
	"database/sql"
	"net/http"

	"back-retro-log/internal/auth"
	"back-retro-log/internal/db"
	"back-retro-log/internal/handlers"
)

func AuthMiddleware(queries *db.Queries, sessions *auth.SessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := sessions.Get(r)
			if err != nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			user, err := queries.GetUserByID(r.Context(), session.UserID)
			if err != nil {
				if err == sql.ErrNoRows {
					_ = sessions.Delete(w, r)
					http.Redirect(w, r, "/login", http.StatusSeeOther)
					return
				}
				http.Error(w, "Internal error", http.StatusInternalServerError)
				return
			}

			ctx := context.WithValue(r.Context(), handlers.CtxUserID, user.ID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
