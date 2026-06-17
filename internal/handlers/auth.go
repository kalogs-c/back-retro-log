package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"back-retro-log/internal/auth"
	"back-retro-log/internal/db"
	"back-retro-log/internal/i18n"
	"back-retro-log/ui"
)

type AuthHandler struct {
	Queries  *db.Queries
	Sessions *auth.SessionManager
	BaseURL  string
}

func (h *AuthHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	count, _ := h.Queries.CountUsers(r.Context())
	ui.Layout(false, ui.LoginPage("", count > 0)).Render(r.Context(), w)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	loc := i18n.FromContext(r.Context())

	if err := r.ParseForm(); err != nil {
		http.Error(w, loc.T("error_bad_request"), http.StatusBadRequest)
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")

	count, _ := h.Queries.CountUsers(r.Context())
	usersExist := count > 0

	user, err := h.Queries.GetUserByUsername(r.Context(), username)
	if err != nil {
		if err == sql.ErrNoRows {
			ui.Layout(false, ui.LoginPage(loc.T("login_invalid"), usersExist)).Render(r.Context(), w)
			return
		}
		http.Error(w, loc.T("error_internal"), http.StatusInternalServerError)
		return
	}

	ok, err := auth.VerifyPassword(password, user.PasswordHash)
	if err != nil || !ok {
		ui.Layout(false, ui.LoginPage(loc.T("login_invalid"), usersExist)).Render(r.Context(), w)
		return
	}

	if err := h.Sessions.Create(w, user.ID); err != nil {
		http.Error(w, loc.T("error_internal"), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/catalog", http.StatusSeeOther)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	_ = h.Sessions.Delete(w, r)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *AuthHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	loc := i18n.FromContext(r.Context())
	count, _ := h.Queries.CountUsers(r.Context())
	token := r.URL.Query().Get("token")

	if count > 0 && token == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	errorMsg := ""
	if token != "" {
		inv, err := h.Queries.GetInvitationByToken(r.Context(), token)
		if err != nil {
			errorMsg = loc.T("register_invalid_token")
		} else if inv.UsedBy.Valid {
			errorMsg = loc.T("register_token_used")
		} else {
			expiresAt, err := time.Parse(time.RFC3339, inv.ExpiresAt)
			if err != nil || time.Now().UTC().After(expiresAt) {
				errorMsg = loc.T("register_token_expired")
			}
		}
	}

	ui.Layout(false, ui.RegisterPage(token, errorMsg)).Render(r.Context(), w)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	loc := i18n.FromContext(r.Context())

	if err := r.ParseForm(); err != nil {
		http.Error(w, loc.T("error_bad_request"), http.StatusBadRequest)
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")
	token := r.FormValue("token")

	if username == "" || password == "" {
		ui.Layout(false, ui.RegisterPage(token, loc.T("register_required"))).Render(r.Context(), w)
		return
	}

	existingUser, err := h.Queries.GetUserByUsername(r.Context(), username)
	if err == nil && existingUser.ID > 0 {
		ui.Layout(false, ui.RegisterPage(token, loc.T("register_taken"))).Render(r.Context(), w)
		return
	}

	count, _ := h.Queries.CountUsers(r.Context())

	var invID int64
	if count > 0 {
		if token == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		inv, err := h.Queries.GetInvitationByToken(r.Context(), token)
		if err != nil || inv.UsedBy.Valid {
			ui.Layout(false, ui.RegisterPage(token, loc.T("register_invalid_invite"))).Render(r.Context(), w)
			return
		}
		expiresAt, err := time.Parse(time.RFC3339, inv.ExpiresAt)
		if err != nil || time.Now().UTC().After(expiresAt) {
			ui.Layout(false, ui.RegisterPage(token, loc.T("register_token_expired"))).Render(r.Context(), w)
			return
		}
		invID = inv.ID
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		http.Error(w, loc.T("error_internal"), http.StatusInternalServerError)
		return
	}

	user, err := h.Queries.CreateUser(r.Context(), db.CreateUserParams{
		Username:     username,
		PasswordHash: hash,
	})
	if err != nil {
		ui.Layout(false, ui.RegisterPage(token, loc.T("register_taken"))).Render(r.Context(), w)
		return
	}

	if invID != 0 {
		_ = h.Queries.UseInvitation(r.Context(), db.UseInvitationParams{
			UsedBy: sql.NullInt64{Int64: user.ID, Valid: true},
			ID:     invID,
		})
	}

	if err := h.Sessions.Create(w, user.ID); err != nil {
		http.Error(w, loc.T("error_internal"), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/catalog", http.StatusSeeOther)
}

func (h *AuthHandler) InvitePage(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(CtxUserID).(int64)
	invitations, _ := h.Queries.ListInvitationsByUser(r.Context(), userID)
	link := r.URL.Query().Get("flash")
	ui.Layout(true, ui.InvitePage(invitations, link)).Render(r.Context(), w)
}

func (h *AuthHandler) CreateInvite(w http.ResponseWriter, r *http.Request) {
	loc := i18n.FromContext(r.Context())
	userID := r.Context().Value(CtxUserID).(int64)

	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		http.Error(w, loc.T("error_internal"), http.StatusInternalServerError)
		return
	}
	token := hex.EncodeToString(b)
	expiresAt := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)

	_, err := h.Queries.CreateInvitation(r.Context(), db.CreateInvitationParams{
		Token:     token,
		CreatedBy: userID,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		http.Error(w, loc.T("error_invite_create"), http.StatusInternalServerError)
		return
	}

	link := fmt.Sprintf("%s/register?token=%s", h.BaseURL, token)
	http.Redirect(w, r, "/invite?flash="+link, http.StatusSeeOther)
}

func (h *AuthHandler) InviteCopy(w http.ResponseWriter, r *http.Request) {
	ui.InviteCopyResult(true).Render(r.Context(), w)
}
