package main

import (
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

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
