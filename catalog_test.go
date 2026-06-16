package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func entryID(t *testing.T, ts *httptest.Server, jar http.CookieJar) string {
	t.Helper()
	resp := request(t, ts, "GET", "/catalog", nil, jar)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	re := regexp.MustCompile(`card-(\d+)`)
	m := re.FindStringSubmatch(string(body))
	if len(m) < 2 {
		t.Fatal("could not find entry ID in catalog page")
	}
	return m[1]
}

func addGame(t *testing.T, ts *httptest.Server, jar http.CookieJar, rawgID int, title string) {
	t.Helper()
	body := url.Values{
		"rawg_id":      {strconv.Itoa(rawgID)},
		"title":        {title},
		"cover_url":    {""},
		"release_date": {"2024"},
	}
	resp := request(t, ts, "POST", "/catalog/add", strings.NewReader(body.Encode()), jar)
	resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("add game %q: expected 303, got %d", title, resp.StatusCode)
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

func TestCatalogStatusUpdate(t *testing.T) {
	_, _, router := testApp(t)
	ts := httptest.NewServer(router)
	defer ts.Close()

	jar, _ := cookiejar.New(nil)
	register(t, ts, jar)
	addGame(t, ts, jar, 1, "Zelda")

	id := entryID(t, ts, jar)
	body := url.Values{"status": {"finished"}}
	resp := request(t, ts, "PUT", fmt.Sprintf("/catalog/%s/status", id), strings.NewReader(body.Encode()), jar)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("update status: expected 200, got %d", resp.StatusCode)
	}

	if h := resp.Header.Get("HX-Trigger"); h == "" {
		t.Fatal("expected HX-Trigger header on status update")
	}

	updated, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(updated), "finished") {
		t.Fatal("expected rendered card to show new status")
	}

	resp = request(t, ts, "GET", "/catalog", nil, jar)
	defer resp.Body.Close()
	all, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(all), "finished") {
		t.Fatal("expected catalog page to reflect updated status")
	}
}

func TestCatalogDelete(t *testing.T) {
	_, _, router := testApp(t)
	ts := httptest.NewServer(router)
	defer ts.Close()

	jar, _ := cookiejar.New(nil)
	register(t, ts, jar)
	addGame(t, ts, jar, 1, "Zelda")

	id := entryID(t, ts, jar)
	resp := request(t, ts, "DELETE", fmt.Sprintf("/catalog/%s", id), nil, jar)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("delete: expected 200, got %d", resp.StatusCode)
	}

	resp = request(t, ts, "GET", "/catalog", nil, jar)
	defer resp.Body.Close()
	all, _ := io.ReadAll(resp.Body)
	if strings.Contains(string(all), "Zelda") {
		t.Fatal("expected game to be removed from catalog after delete")
	}
}

func TestCatalogSearch(t *testing.T) {
	_, _, router := testApp(t)
	ts := httptest.NewServer(router)
	defer ts.Close()

	jar, _ := cookiejar.New(nil)
	register(t, ts, jar)
	addGame(t, ts, jar, 1, "The Legend of Zelda")
	addGame(t, ts, jar, 2, "Elden Ring")

	resp := request(t, ts, "GET", "/catalog/search?q=Zelda", nil, jar)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("catalog search: expected 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Zelda") {
		t.Fatal("expected Zelda in search results")
	}
	if strings.Contains(string(body), "Elden") {
		t.Fatal("expected Elden Ring NOT in search results for 'Zelda'")
	}

	resp = request(t, ts, "GET", "/catalog/search?q=nonexistent", nil, jar)
	defer resp.Body.Close()
	body, _ = io.ReadAll(resp.Body)
	if strings.Contains(string(body), "Zelda") {
		t.Fatal("expected no games in search results for 'nonexistent'")
	}

	resp = request(t, ts, "GET", "/catalog/search", nil, jar)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("catalog search empty query: expected 200, got %d", resp.StatusCode)
	}
	body, _ = io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Zelda") || !strings.Contains(string(body), "Elden") {
		t.Fatal("expected all games with empty search query")
	}
}
