package app

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreatingShortURL(t *testing.T) {
	c := NewInMemoryController()

	const url = "https://pkg.go.dev/cmp"

	key := invokeShortener(t, url, c)
	lookupUrl := invokeLookup(t, key, c)

	if lookupUrl != url {
		t.Errorf("Expected %q, got %q", url, lookupUrl)
	}
}

func invokeShortener(t *testing.T, url string, c ShortURLController) string {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(url))
	w := httptest.NewRecorder()

	c.RouteRequest(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	body := string(bytes)

	if !strings.HasPrefix(body, "http://localhost:8080/") { // Assuming the key is "1" for this test
		t.Errorf("Expected body %q to start with http://localhost:8080/", body)
	}

	idx := strings.LastIndex(body, "/")
	key := body[idx+1:]
	if len(key) < 6 {
		t.Errorf("Expected key length to be at least 6, got %d", len(key))
	}

	return key
}

func invokeLookup(t *testing.T, key string, c ShortURLController) string {
	r := httptest.NewRequest(http.MethodGet, "/"+key, nil)
	w := httptest.NewRecorder()

	c.RouteRequest(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusTemporaryRedirect {
		t.Errorf("Expected status code %d, got %d", http.StatusTemporaryRedirect, resp.StatusCode)
	}

	loc := resp.Header.Get("Location")

	if loc == "" {
		t.Error("Expected Location header to be set")
	}

	return loc
}

func TestInvalidRequest(t *testing.T) {
	c := NewInMemoryController()

	r := httptest.NewRequest(http.MethodPut, "/somekey", nil)
	w := httptest.NewRecorder()

	c.RouteRequest(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}
