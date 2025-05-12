package app

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreatingShortURL(t *testing.T) {
	c := NewInMemoryController()

	const url = "https://pkg.go.dev/cmp"

	key := invokeShortener(t, url, c)
	lookupUrl := invokeLookup(t, key, c)

	assert.Equal(t, url, lookupUrl, "Expected the original URL to match the lookup URL")
}

func invokeShortener(t *testing.T, url string, c ShortURLController) string {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(url))
	w := httptest.NewRecorder()

	c.RouteRequest(w, r)

	resp := w.Result()
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Response status code")

	bytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body")

	body := string(bytes)

	assert.Regexp(t, `^http://localhost:8080/`, body, "Expected body to start with http://localhost:8080/")

	idx := strings.LastIndex(body, "/")
	key := body[idx+1:]
	assert.GreaterOrEqual(t, len(key), 6, "Expected key length to be at least 6")

	return key
}

func invokeLookup(t *testing.T, key string, c ShortURLController) string {
	r := httptest.NewRequest(http.MethodGet, "/"+key, nil)
	w := httptest.NewRecorder()

	c.RouteRequest(w, r)

	resp := w.Result()
	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode, "Response status code")

	loc := resp.Header.Get("Location")
	assert.NotEmpty(t, loc, "Expected Location header to be set")

	return loc
}

func TestInvalidRequest(t *testing.T) {
	c := NewInMemoryController()

	r := httptest.NewRequest(http.MethodPut, "/somekey", nil)
	w := httptest.NewRecorder()

	c.RouteRequest(w, r)

	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Response status code")
}
