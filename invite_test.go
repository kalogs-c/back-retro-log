package main

import (
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
)

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

func TestInviteFlow(t *testing.T) {
	_, _, router := testApp(t)
	ts := httptest.NewServer(router)
	defer ts.Close()

	jarA, _ := cookiejar.New(nil)
	register(t, ts, jarA)

	body := url.Values{}
	resp := request(t, ts, "POST", "/invite", strings.NewReader(body.Encode()), jarA)
	resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("create invite: expected 303, got %d", resp.StatusCode)
	}

	loc := resp.Header.Get("Location")
	if loc == "" {
		t.Fatal("expected Location header after invite creation")
	}

	resp = request(t, ts, "GET", loc, nil, jarA)
	defer resp.Body.Close()
	all, _ := io.ReadAll(resp.Body)
	re := regexp.MustCompile(`/register\?token=([a-f0-9]+)`)
	m := re.FindStringSubmatch(string(all))
	if len(m) < 2 {
		t.Fatalf("could not find invite token in invite page: %s", string(all))
	}
	token := m[1]

	jarB, _ := cookiejar.New(nil)
	body = url.Values{"username": {"newuser"}, "password": {"newpass"}, "token": {token}}
	resp = request(t, ts, "POST", "/register", strings.NewReader(body.Encode()), jarB)
	resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("register with invite: expected 303, got %d", resp.StatusCode)
	}

	resp = request(t, ts, "GET", "/catalog", nil, jarB)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("catalog for invited user: expected 200, got %d", resp.StatusCode)
	}
	all, _ = io.ReadAll(resp.Body)
	if !strings.Contains(string(all), "My Catalog") {
		t.Fatal("expected catalog page for invited user")
	}
}
