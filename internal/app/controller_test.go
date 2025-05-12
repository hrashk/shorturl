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

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://pkg.go.dev/cmp"))
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
}
