package app

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
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
	h := InMemoryHandler()

	r := chi.NewRouter()
	r.Use(suite.loggingMiddleware)
	r.Mount("/", h)

	suite.srv = httptest.NewServer(r)
	suite.srv.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
}

func (suite *ControllerTestSuite) TearDownTest() {
	if suite.srv != nil {
		suite.srv.Close()
	}
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

func (suite *ControllerTestSuite) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suite.T().Logf("Request: %s %s", r.Method, r.URL)

		rec := httptest.NewRecorder()
		next.ServeHTTP(rec, r)

		suite.T().Logf("Response: %d %s", rec.Code, rec.Body.String())

		for k, v := range rec.Header() {
			w.Header()[k] = v
		}
		w.WriteHeader(rec.Code)
		_, _ = io.Copy(w, rec.Body)
	})
}
