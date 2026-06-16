package main

import (
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealth(t *testing.T) {
	_, _, router := testApp(t)
	ts := httptest.NewServer(router)
	defer ts.Close()

	resp := request(t, ts, "GET", "/health", nil, nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("health: expected 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "OK" {
		t.Fatalf("health: expected 'OK', got %s", string(body))
	}
}

func TestRootRedirect(t *testing.T) {
	_, _, router := testApp(t)
	ts := httptest.NewServer(router)
	defer ts.Close()

	jar, _ := cookiejar.New(nil)
	register(t, ts, jar)

	resp := request(t, ts, "GET", "/", nil, jar)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("root: expected 303, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/catalog" {
		t.Fatalf("root: expected redirect to /catalog, got %s", loc)
	}
}

func TestStaticFiles(t *testing.T) {
	_, _, router := testApp(t)
	ts := httptest.NewServer(router)
	defer ts.Close()

	resp := request(t, ts, "GET", "/static/style.css", nil, nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("static css: expected 200, got %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "css") && !strings.Contains(ct, "text") {
		t.Fatalf("static css: expected text/css content-type, got %s", ct)
	}
}
