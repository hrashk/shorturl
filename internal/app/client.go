package app

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const ContentTypeJSON = "application/json"

type Client struct {
	BaseURL string
	t       testing.TB
	hcl     *http.Client
}

func NewClient(tb testing.TB) Client {
	c := Client{
		t:   tb,
		hcl: &http.Client{},
	}
	c.hcl.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return c
}

func (c Client) Shorten(url, baseURL string) string {
	body := c.callShortener(url)

	return c.extractKey(baseURL, body)
}

func (c Client) extractKey(baseURL string, body string) string {
	assert.Regexp(c.t, "^"+baseURL, body, "Redirect URL")

	idx := strings.LastIndex(body, "/")
	key := body[idx+1:]
	assert.GreaterOrEqual(c.t, len(key), 6, "Expected key length to be at least 6")

	return key
}

func (c Client) callShortener(url string) string {
	resp := c.POST("", "text/plain", url)
	defer resp.Body.Close()

	assert.Equal(c.t, http.StatusCreated, resp.StatusCode, "Response status code")

	return c.readBody(resp.Body)
}

func (c Client) readBody(body io.Reader) string {
	bytes, err := io.ReadAll(body)
	require.NoError(c.t, err, "Failed to read response body")

	return string(bytes)
}

func (c Client) POST(query string, contentType string, body string) *http.Response {
	resp, err := c.hcl.Post(c.BaseURL+query, contentType, strings.NewReader(body))
	require.NoError(c.t, err, "Failed to POST")

	return resp
}

func (c Client) PostJSON(query string, body string) *http.Response {
	return c.POST(query, ContentTypeJSON, body)
}

func (c Client) PUT(query string, contentType string, body string) *http.Response {
	req, err := http.NewRequest(http.MethodPut, c.BaseURL+query, strings.NewReader(body))
	require.NoError(c.t, err, "Failed to creae a PUT request")
	req.Header.Set("Content-Type", contentType)

	resp, err := c.hcl.Do(req)
	require.NoError(c.t, err, "Failed to PUT")

	return resp
}

func (c Client) PutJSON(query string, body string) *http.Response {
	return c.PUT(query, ContentTypeJSON, body)
}

func (c Client) PostAcceptingGzip(query string, body string) *http.Response {
	req, err := http.NewRequest(http.MethodPost, c.BaseURL+query, strings.NewReader(body))
	require.NoError(c.t, err, "Failed to creae a POST request")
	req.Header.Set("Content-Type", ContentTypeJSON)
	req.Header.Add("Accept-Encoding", "gzip")

	resp, err := c.hcl.Do(req)
	require.NoError(c.t, err, "Failed to POST")

	return resp
}

func (c Client) PostGzippedJSON(query string, body string) *http.Response {
	b, err := compress([]byte(body))
	require.NoError(c.t, err, "Failed to compress")

	req, err := http.NewRequest(http.MethodPost, c.BaseURL+query, bytes.NewReader(b))
	require.NoError(c.t, err, "Failed to creae a POST request")
	req.Header.Set("Content-Type", ContentTypeJSON)
	req.Header.Add("Content-Encoding", "gzip")

	resp, err := c.hcl.Do(req)
	require.NoError(c.t, err, "Failed to POST")

	return resp
}

func (c Client) LookUp(shortURL string) string {
	resp := c.GET("/" + shortURL)
	defer resp.Body.Close()

	assert.Equal(c.t, http.StatusTemporaryRedirect, resp.StatusCode, "Response status code")

	loc := resp.Header.Get("Location")
	assert.NotEmpty(c.t, loc, "Expected Location header to be set")

	return loc
}

func (c Client) LookUpNotFound(shortURL string) {
	resp := c.GET("/" + shortURL)
	defer resp.Body.Close()

	assert.Equal(c.t, http.StatusNotFound, resp.StatusCode, "response status code")
}

func (c Client) GET(query string) *http.Response {
	resp, err := c.hcl.Get(c.BaseURL + query)
	require.NoError(c.t, err, "Failed to make request")

	return resp
}

func (c Client) Ping() {
	resp := c.GET("/ping")
	defer resp.Body.Close()

	require.Equal(c.t, http.StatusOK, resp.StatusCode, "response status code")
}

func (c Client) PingFailed() {
	resp := c.GET("/ping")
	defer resp.Body.Close()

	assert.Equal(c.t, http.StatusInternalServerError, resp.StatusCode, "response status code")
}

func compress(data []byte) ([]byte, error) {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)

	_, err := w.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed write data : %w", err)
	}

	err = w.Close()
	if err != nil {
		return nil, fmt.Errorf("failed compress data: %w", err)
	}
	return b.Bytes(), nil
}
