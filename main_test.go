package main

import (
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"back-retro-log/internal/app"
	"back-retro-log/internal/auth"
	"back-retro-log/internal/db"
	"back-retro-log/internal/providers"

	_ "modernc.org/sqlite"
)

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	if err := runMigrations(sqlDB); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}
	return sqlDB
}

func testApp(t *testing.T) (queries *db.Queries, sessions *auth.SessionManager, router http.Handler) {
	t.Helper()
	sqlDB := testDB(t)
	queries = db.New(sqlDB)
	sessions = auth.NewSessionManager(queries)
	provider := providers.NewDummy()
	router = app.NewRouter(queries, sessions, provider, "http://test.local", staticFS)
	return
}

func noRedirectClient(jar http.CookieJar) *http.Client {
	return &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func request(t *testing.T, ts *httptest.Server, method, path string, body io.Reader, jar http.CookieJar) *http.Response {
	t.Helper()
	client := noRedirectClient(jar)
	req, err := http.NewRequest(method, ts.URL+path, body)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to do request: %v", err)
	}
	return resp
}

func register(t *testing.T, ts *httptest.Server, jar http.CookieJar) {
	t.Helper()
	body := url.Values{"username": {"testuser"}, "password": {"testpass"}, "token": {""}}
	resp := request(t, ts, "POST", "/register", strings.NewReader(body.Encode()), jar)
	resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("register: expected 303, got %d", resp.StatusCode)
	}
}
