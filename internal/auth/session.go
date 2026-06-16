package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"back-retro-log/internal/db"
)

type SessionManager struct {
	queries    *db.Queries
	cookieName string
	maxAge     time.Duration
}

func NewSessionManager(queries *db.Queries) *SessionManager {
	return &SessionManager{
		queries:    queries,
		cookieName: "backlog_session",
		maxAge:     24 * time.Hour,
	}
}

func (m *SessionManager) Create(w http.ResponseWriter, userID int64) error {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}
	sessionID := hex.EncodeToString(token)
	expiresAt := time.Now().Add(m.maxAge).UTC().Format(time.RFC3339)

	_, err := m.queries.CreateSession(context.Background(), db.CreateSessionParams{
		ID:        sessionID,
		UserID:    userID,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     m.cookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(m.maxAge.Seconds()),
	})
	return nil
}

func (m *SessionManager) Get(r *http.Request) (*db.Session, error) {
	cookie, err := r.Cookie(m.cookieName)
	if err != nil {
		return nil, fmt.Errorf("no session cookie")
	}

	session, err := m.queries.GetSessionByID(context.Background(), cookie.Value)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	expiresAt, err := time.Parse(time.RFC3339, session.ExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("invalid session expiry: %w", err)
	}
	if time.Now().UTC().After(expiresAt) {
		_ = m.queries.DeleteSession(context.Background(), session.ID)
		return nil, fmt.Errorf("session expired")
	}

	return &session, nil
}

func (m *SessionManager) Delete(w http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie(m.cookieName)
	if err == nil {
		_ = m.queries.DeleteSession(context.Background(), cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     m.cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
	return nil
}

func (m *SessionManager) Cleanup() {
	_ = m.queries.DeleteExpiredSessions(context.Background())
}
