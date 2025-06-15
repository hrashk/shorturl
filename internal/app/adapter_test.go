package app

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
)

type AdapterSuite struct {
	suite.Suite
	srv *httptest.Server
	c   Client
}

func TestControllerSuite(t *testing.T) {
	suite.Run(t, &AdapterSuite{})
}

func (as *AdapterSuite) SetupTest() {
	cfg, err := newConfig(WithLogger(as), WithMemoryStorage())
	as.Require().NoError(err)

	a, err := newAdapter(cfg)
	as.Require().NoError(err)

	as.srv = httptest.NewServer(a.handler())
	as.c = NewClient(&as.Suite)
	as.c.BaseURL = as.srv.URL
}

func (as *AdapterSuite) TearDownTest() {
	if as.srv != nil {
		as.srv.Close()
	}
}

func (as *AdapterSuite) TestCreatingShortURL() {
	const url = "https://pkg.go.dev/cmp"

	key := as.c.Shorten(url, DefaultBaseURL)
	lookupURL := as.c.LookUp(key)

	as.Equal(url, lookupURL, "Expected the original URL to match the lookup URL")
}

func (as *AdapterSuite) TestInvalidRequest() {
	resp := as.c.PutJSON("/somekey", "")
	defer resp.Body.Close()

	as.Equal(http.StatusBadRequest, resp.StatusCode, "Response status code")
}

func (as *AdapterSuite) TestDifferentKeys() {
	const url = "https://pkg.go.dev/cmp"
	const url2 = "https://pkg.go.dev/cmp/v2"

	key := as.c.Shorten(url, DefaultBaseURL)
	key2 := as.c.Shorten(url2, DefaultBaseURL)
	as.NotEqual(key, key2, "Expected different keys for different URLs")
}

func (as *AdapterSuite) TestShortenApi() {
	resp := as.c.PostJSON("/api/shorten", `{"url": "https://pkg.go.dev/cmp"}`)
	defer resp.Body.Close()

	body := as.c.readBody(resp.Body)
	as.Contains(body, DefaultBaseURL, "body")
}

func (as *AdapterSuite) Info(msg string, v ...any) {
	as.T().Logf(msg, v...)
}

func (as *AdapterSuite) Error(err error, msg string, v ...any) {
	as.T().Logf(msg+": error "+err.Error(), v...)
}

func (as *AdapterSuite) TestReceivingGzip() {
	resp := as.c.PostAcceptingGzip("/api/shorten", `{"url": "https://pkg.go.dev/cmp"}`)
	defer resp.Body.Close()

	as.Equal(http.StatusCreated, resp.StatusCode, "Response status code")
	as.Equal("gzip", resp.Header.Get("Content-Encoding"), "Content encoding")

	body := as.c.readBody(resp.Body)
	as.Contains(body, DefaultBaseURL, "body")
}

func (as *AdapterSuite) TestSendingGzip() {
	resp := as.c.PostGzippedJSON("/api/shorten", `{"url": "https://pkg.go.dev/cmp"}`)
	defer resp.Body.Close()

	as.Equal(http.StatusCreated, resp.StatusCode, "Response status code")

	bytes, err := io.ReadAll(resp.Body)
	as.Require().NoError(err, "Failed to read response body")

	body := string(bytes)
	as.Contains(body, DefaultBaseURL, "body")
}
