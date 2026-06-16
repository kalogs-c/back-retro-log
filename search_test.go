package main

import (
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"testing"
)

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
