package app

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ControllerTestSuite struct {
	suite.Suite
	srv *httptest.Server
}

func TestExampleTestSuite(t *testing.T) {
	suite.Run(t, &ControllerTestSuite{})
}

func (suite *ControllerTestSuite) SetupTest() {
	suite.srv = httptest.NewServer(loggingMiddleware(InMemoryHandler()))
	suite.srv.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
}

func (suite *ControllerTestSuite) TearDownTest() {
	suite.srv.Close()
}

func (suite *ControllerTestSuite) TestCreatingShortURL() {
	const url = "https://pkg.go.dev/cmp"

	key := suite.invokeShortener(url)
	lookupURL := suite.invokeLookup(key)

	suite.Equal(url, lookupURL, "Expected the original URL to match the lookup URL")
}

func (suite *ControllerTestSuite) TestInvalidRequest() {
	req, err := http.NewRequest(http.MethodPut, suite.srv.URL+"/somekey", nil)
	suite.Require().NoError(err, "Failed to create a request")

	resp, err := suite.srv.Client().Do(req)
	suite.Require().NoError(err, "Failed to make request")
	defer resp.Body.Close()

	suite.Equal(http.StatusBadRequest, resp.StatusCode, "Response status code")
}

func (suite *ControllerTestSuite) TestDifferentKeys() {
	const url = "https://pkg.go.dev/cmp"
	const url2 = "https://pkg.go.dev/cmp/v2"

	key := suite.invokeShortener(url)
	key2 := suite.invokeShortener(url2)
	suite.NotEqual(key, key2, "Expected different keys for different URLs")
}

func (suite *ControllerTestSuite) invokeShortener(url string) string {
	req, err := http.NewRequest(http.MethodPost, suite.srv.URL, strings.NewReader(url))
	suite.Require().NoError(err, "Failed to create a request")

	resp, err := suite.srv.Client().Do(req)
	suite.Require().NoError(err, "Failed to make request")
	defer resp.Body.Close()

	suite.Equal(http.StatusCreated, resp.StatusCode, "Response status code")

	bytes, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err, "Failed to read response body")

	body := string(bytes)

	suite.Regexp(`^http://localhost:8080/`, body, "Expected body to start with http://localhost:8080/")

	idx := strings.LastIndex(body, "/")
	key := body[idx+1:]
	suite.GreaterOrEqual(len(key), 6, "Expected key length to be at least 6")

	return key
}

func (suite *ControllerTestSuite) invokeLookup(key string) string {
	resp, err := suite.srv.Client().Get(suite.srv.URL + "/" + key)
	suite.Require().NoError(err, "Failed to make request")
	defer resp.Body.Close()

	suite.Equal(http.StatusTemporaryRedirect, resp.StatusCode, "Response status code")

	loc := resp.Header.Get("Location")
	suite.NotEmpty(loc, "Expected Location header to be set")

	return loc
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log the incoming request
		log.Printf("Request: %s %s", r.Method, r.URL)

		// Capture the response using a ResponseRecorder
		rec := httptest.NewRecorder()
		next.ServeHTTP(rec, r)

		// Log the response
		log.Printf("Response: %d %s", rec.Code, rec.Body.String())

		// Copy the recorded response to the actual response writer
		for k, v := range rec.Header() {
			w.Header()[k] = v
		}
		w.WriteHeader(rec.Code)
		_, _ = io.Copy(w, rec.Body)
	})
}
