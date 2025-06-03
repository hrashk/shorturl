package app

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ControllerSuite struct {
	suite.Suite
	srv *httptest.Server
}

func TestControllerSuite(t *testing.T) {
	suite.Run(t, &ControllerSuite{})
}

func (suite *ControllerSuite) SetupTest() {
	config.redirectBaseURL = "http://localhost:8888"

	h := NewHandlerWithLogger(suite)

	suite.srv = httptest.NewServer(h)
	suite.srv.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
}

func (suite *ControllerSuite) TearDownTest() {
	if suite.srv != nil {
		suite.srv.Close()
	}
}

func (suite *ControllerSuite) TestCreatingShortURL() {
	const url = "https://pkg.go.dev/cmp"

	key := suite.invokeShortener(url)
	lookupURL := suite.invokeLookup(key)

	suite.Equal(url, lookupURL, "Expected the original URL to match the lookup URL")
}

func (suite *ControllerSuite) TestInvalidRequest() {
	req, err := http.NewRequest(http.MethodPut, suite.srv.URL+"/somekey", nil)
	suite.Require().NoError(err, "Failed to create a request")

	resp, err := suite.srv.Client().Do(req)
	suite.Require().NoError(err, "Failed to make request")
	defer resp.Body.Close()

	suite.Equal(http.StatusBadRequest, resp.StatusCode, "Response status code")
}

func (suite *ControllerSuite) TestDifferentKeys() {
	const url = "https://pkg.go.dev/cmp"
	const url2 = "https://pkg.go.dev/cmp/v2"

	key := suite.invokeShortener(url)
	key2 := suite.invokeShortener(url2)
	suite.NotEqual(key, key2, "Expected different keys for different URLs")
}

func (suite *ControllerSuite) TestShortenApi() {
	resp, err := suite.srv.Client().Post(suite.srv.URL+"/api/shorten",
		"application/json",
		strings.NewReader(`{"url": "https://pkg.go.dev/cmp"}`))
	suite.Require().NoError(err, "Failed to POST")
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode, "Response status code")

	bytes, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err, "Failed to read response body")

	body := string(bytes)
	suite.Contains(body, config.redirectBaseURL, "Expected body to start with http://localhost:8080/")
}

func (suite *ControllerSuite) invokeShortener(url string) string {
	req, err := http.NewRequest(http.MethodPost, suite.srv.URL, strings.NewReader(url))
	suite.Require().NoError(err, "Failed to create a request")

	resp, err := suite.srv.Client().Do(req)
	suite.Require().NoError(err, "Failed to make request")
	defer resp.Body.Close()

	suite.Equal(http.StatusCreated, resp.StatusCode, "Response status code")

	bytes, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err, "Failed to read response body")

	body := string(bytes)

	suite.Regexp("^"+config.redirectBaseURL, body, "Expected body to start with http://localhost:8080/")

	idx := strings.LastIndex(body, "/")
	key := body[idx+1:]
	suite.GreaterOrEqual(len(key), 6, "Expected key length to be at least 6")

	return key
}

func (suite *ControllerSuite) invokeLookup(key string) string {
	resp, err := suite.srv.Client().Get(suite.srv.URL + "/" + key)
	suite.Require().NoError(err, "Failed to make request")
	defer resp.Body.Close()

	suite.Equal(http.StatusTemporaryRedirect, resp.StatusCode, "Response status code")

	loc := resp.Header.Get("Location")
	suite.NotEmpty(loc, "Expected Location header to be set")

	return loc
}

func (suite *ControllerSuite) Info(msg string, fields ...any) {
	suite.T().Logf(msg, fields...)
}
