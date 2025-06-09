package app

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type AdapterSuite struct {
	suite.Suite
	srv *httptest.Server
}

func TestControllerSuite(t *testing.T) {
	suite.Run(t, &AdapterSuite{})
}

func (as *AdapterSuite) SetupTest() {
	cfg, err := newConfig(Logger(as), StoragePath(""))
	as.Require().NoError(err)

	h, err := newHandler(cfg)
	as.Require().NoError(err)

	as.srv = httptest.NewServer(h)
	as.srv.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
}

func (as *AdapterSuite) TearDownTest() {
	if as.srv != nil {
		as.srv.Close()
	}
}

func (as *AdapterSuite) TestCreatingShortURL() {
	const url = "https://pkg.go.dev/cmp"

	key := as.invokeShortener(url)
	lookupURL := as.invokeLookup(key)

	as.Equal(url, lookupURL, "Expected the original URL to match the lookup URL")
}

func (as *AdapterSuite) TestInvalidRequest() {
	req, err := http.NewRequest(http.MethodPut, as.srv.URL+"/somekey", nil)
	as.Require().NoError(err, "Failed to create a request")

	resp, err := as.srv.Client().Do(req)
	as.Require().NoError(err, "Failed to make request")
	defer resp.Body.Close()

	as.Equal(http.StatusBadRequest, resp.StatusCode, "Response status code")
}

func (as *AdapterSuite) TestDifferentKeys() {
	const url = "https://pkg.go.dev/cmp"
	const url2 = "https://pkg.go.dev/cmp/v2"

	key := as.invokeShortener(url)
	key2 := as.invokeShortener(url2)
	as.NotEqual(key, key2, "Expected different keys for different URLs")
}

func (as *AdapterSuite) TestShortenApi() {
	resp, err := as.srv.Client().Post(as.srv.URL+"/api/shorten",
		"application/json",
		strings.NewReader(`{"url": "https://pkg.go.dev/cmp"}`))
	as.Require().NoError(err, "Failed to POST")
	defer resp.Body.Close()

	as.Equal(http.StatusCreated, resp.StatusCode, "Response status code")

	bytes, err := io.ReadAll(resp.Body)
	as.Require().NoError(err, "Failed to read response body")

	body := string(bytes)
	as.Contains(body, DefaultBaseURL, "body")
}

func (as *AdapterSuite) invokeShortener(url string) string {
	req, err := http.NewRequest(http.MethodPost, as.srv.URL, strings.NewReader(url))
	as.Require().NoError(err, "Failed to create a request")

	resp, err := as.srv.Client().Do(req)
	as.Require().NoError(err, "Failed to make request")
	defer resp.Body.Close()

	as.Equal(http.StatusCreated, resp.StatusCode, "Response status code")

	bytes, err := io.ReadAll(resp.Body)
	as.Require().NoError(err, "Failed to read response body")

	body := string(bytes)

	as.Regexp("^"+DefaultBaseURL, body, "body")

	idx := strings.LastIndex(body, "/")
	key := body[idx+1:]
	as.GreaterOrEqual(len(key), 6, "Expected key length to be at least 6")

	return key
}

func (as *AdapterSuite) invokeLookup(key string) string {
	resp, err := as.srv.Client().Get(as.srv.URL + "/" + key)
	as.Require().NoError(err, "Failed to make request")
	defer resp.Body.Close()

	as.Equal(http.StatusTemporaryRedirect, resp.StatusCode, "Response status code")

	loc := resp.Header.Get("Location")
	as.NotEmpty(loc, "Expected Location header to be set")

	return loc
}

func (as *AdapterSuite) Info(msg string, v ...any) {
	as.T().Logf(msg, v...)
}

func (as *AdapterSuite) Error(msg string, err error, v ...any) {
	as.T().Logf(msg+": error "+err.Error(), v...)
}

func (as *AdapterSuite) TestReceivingGzip() {
	req, err := http.NewRequest(http.MethodPost, as.srv.URL+"/api/shorten",
		strings.NewReader(`{"url": "https://pkg.go.dev/cmp"}`))
	as.Require().NoError(err, "Failed to POST")

	req.Header.Add("Accept-Encoding", "gzip")

	resp, err := as.srv.Client().Do(req)
	as.Require().NoError(err, "Failed to POST")
	defer resp.Body.Close()

	as.Equal(http.StatusCreated, resp.StatusCode, "Response status code")
	as.Equal("gzip", resp.Header.Get("Content-Encoding"), "Content encoding")

	bytes, err := io.ReadAll(resp.Body)
	as.Require().NoError(err, "Failed to read response body")

	body := string(bytes)
	as.Contains(body, DefaultBaseURL, "body")
}

func (as *AdapterSuite) TestSendingGzip() {
	b, err := compress([]byte(`{"url": "https://pkg.go.dev/cmp"}`))
	as.Require().NoError(err, "Failed to compress")

	req, err := http.NewRequest(http.MethodPost, as.srv.URL+"/api/shorten",
		bytes.NewReader(b))
	as.Require().NoError(err, "Failed to POST")

	req.Header.Add("Content-Encoding", "gzip")

	resp, err := as.srv.Client().Do(req)
	as.Require().NoError(err, "Failed to POST")
	defer resp.Body.Close()

	as.Equal(http.StatusCreated, resp.StatusCode, "Response status code")

	bytes, err := io.ReadAll(resp.Body)
	as.Require().NoError(err, "Failed to read response body")

	body := string(bytes)
	as.Contains(body, DefaultBaseURL, "body")
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
