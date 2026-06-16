package main

import (
	"database/sql"
	"io"
	"net/http"
	"net/http/cookiejar"
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

func TestRegistrationAndLogin(t *testing.T) {
	_, _, router := testApp(t)
	ts := httptest.NewServer(router)
	defer ts.Close()

	jar, _ := cookiejar.New(nil)

	body := url.Values{"username": {"alice"}, "password": {"secret123"}, "token": {""}}
	resp := request(t, ts, "POST", "/register", strings.NewReader(body.Encode()), jar)
	resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected 303 after register, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/catalog" {
		t.Fatalf("expected redirect to /catalog, got %s", loc)
	}

	resp = request(t, ts, "POST", "/register", strings.NewReader(body.Encode()), jar)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 with error for duplicate, got %d", resp.StatusCode)
	}
}

func TestAuthMiddleware(t *testing.T) {
	_, _, router := testApp(t)
	ts := httptest.NewServer(router)
	defer ts.Close()

	resp := request(t, ts, "GET", "/catalog", nil, nil)
	resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected 303 redirect to login, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/login" {
		t.Fatalf("expected redirect to /login, got %s", loc)
	}
}

func TestLoginLogout(t *testing.T) {
	_, _, router := testApp(t)
	ts := httptest.NewServer(router)
	defer ts.Close()

	jar, _ := cookiejar.New(nil)
	register(t, ts, jar)

	resp := request(t, ts, "POST", "/logout", nil, jar)
	resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("logout: expected 303, got %d", resp.StatusCode)
	}

	resp = request(t, ts, "GET", "/catalog", nil, jar)
	resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected 303 after logout, got %d", resp.StatusCode)
	}

	body := url.Values{"username": {"testuser"}, "password": {"testpass"}}
	resp = request(t, ts, "POST", "/login", strings.NewReader(body.Encode()), jar)
	resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("login: expected 303, got %d", resp.StatusCode)
	}

	resp = request(t, ts, "GET", "/catalog", nil, jar)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on catalog after login, got %d", resp.StatusCode)
	}
}

func TestCatalogAddAndStatus(t *testing.T) {
	_, _, router := testApp(t)
	ts := httptest.NewServer(router)
	defer ts.Close()

	jar, _ := cookiejar.New(nil)
	register(t, ts, jar)

	body := url.Values{
		"rawg_id":      {"123"},
		"title":        {"Test Game"},
		"cover_url":    {""},
		"release_date": {"2024"},
	}
	resp := request(t, ts, "POST", "/catalog/add", strings.NewReader(body.Encode()), jar)
	resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("add game: expected 303, got %d", resp.StatusCode)
	}

	resp = request(t, ts, "POST", "/catalog/add", strings.NewReader(body.Encode()), jar)
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("duplicate add: expected 409, got %d", resp.StatusCode)
	}
	dupBody, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if !strings.Contains(string(dupBody), "already in your catalog") {
		t.Fatalf("expected per-user error message, got: %s", string(dupBody))
	}

	resp = request(t, ts, "GET", "/catalog", nil, jar)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("catalog: expected 200, got %d", resp.StatusCode)
	}
	all, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(all), "Test Game") {
		t.Fatal("expected game title in catalog page")
	}
}

func TestSearchResults(t *testing.T) {
	_, _, router := testApp(t)
	ts := httptest.NewServer(router)
	defer ts.Close()

	jar, _ := cookiejar.New(nil)
	register(t, ts, jar)

	resp := request(t, ts, "GET", "/search/results?q=zelda", nil, jar)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("search: expected 200, got %d", resp.StatusCode)
	}
	all, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(all), "Elden Ring") {
		t.Fatal("expected dummy provider results in search page")
	}
}

func TestBadLogin(t *testing.T) {
	_, _, router := testApp(t)
	ts := httptest.NewServer(router)
	defer ts.Close()

	body := url.Values{"username": {"nonexistent"}, "password": {"wrong"}}
	resp := request(t, ts, "POST", "/login", strings.NewReader(body.Encode()), nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("bad login: expected 200 with error page, got %d", resp.StatusCode)
	}
	all, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(all), "Invalid") {
		t.Fatal("expected error message on bad login")
	}
}

func TestInviteNoAuth(t *testing.T) {
	_, _, router := testApp(t)
	ts := httptest.NewServer(router)
	defer ts.Close()

	resp := request(t, ts, "GET", "/invite", nil, nil)
	resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("invite without auth: expected 303, got %d", resp.StatusCode)
	}
}
