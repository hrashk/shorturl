package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/hrashk/shorturl/internal/app"
	"github.com/stretchr/testify/suite"
)

const (
	sampleURL  = "https://pkg.go.dev/cmp"
	anotherURL = "https://pkg.go.dev/errors"

	samplePath  = "/tmp/sample-urls.json"
	anotherPath = "/tmp/other-urls.json"

	sampleAddr  = "localhost:8088"
	anotherAddr = "localhost:9099"

	sampleBaseURL  = "http://example.com:1024"
	anotherBaseURL = "http://example.com:4201"

	skip = "#"
)

type MainSuite struct {
	suite.Suite
	origArgs []string
	server   *mainServer
}

func TestMainSuite(t *testing.T) {
	suite.Run(t, &MainSuite{})
}

func (ms *MainSuite) SetupSuite() {
	ms.origArgs = os.Args
	ms.server = newServer(&ms.Suite)

	http.DefaultClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
}

func (ms *MainSuite) SetupTest() {
	ms.setUp()
}

func (ms *MainSuite) SetupSubTest() {
	ms.setUp()
}

func (ms *MainSuite) TearDownTest() {
	ms.tearDown()
}

func (ms *MainSuite) TearDownSubTest() {
	ms.tearDown()
}

func (ms *MainSuite) setUp() {
	ms.deleteFile(app.DefaultStoragePath)
	ms.deleteFile(samplePath)

	// avoid errors due to unknown flags from go test
	os.Args = []string{os.Args[0]}
}

func (ms *MainSuite) deleteFile(path string) {
	if err := os.Remove(path); err != nil {
		ms.ErrorIs(err, os.ErrNotExist, "failed to delete file %s", path)
	}
}

func (ms *MainSuite) tearDown() {
	os.Args = ms.origArgs

	os.Unsetenv(addrSetting.envName)
	os.Unsetenv(baseURLSetting.envName)
	os.Unsetenv(storagePathSetting.envName)

	ms.server.stop()
}

func (ms *MainSuite) TestServerAddress() {
	tests := []struct {
		env, arg, expected string
	}{
		{skip, skip, app.DefaultServerAddress},
		{skip, sampleAddr, sampleAddr},
		{sampleAddr, skip, sampleAddr},
		{anotherAddr, sampleAddr, anotherAddr},
	}

	for i, t := range tests {
		name := fmt.Sprintf("server address %d", i+1)
		ms.Run(name, func() {
			if t.env != skip {
				os.Setenv(addrSetting.envName, t.env)
			}
			if t.arg != skip {
				os.Args = append(os.Args, "-a", t.arg)
			}
			ms.server.start()

			ms.Equal(t.expected, ms.server.addr())
			ms.shorten(sampleURL, app.DefaultBaseURL)
		})
	}
}

func (ms *MainSuite) TestBaseURL() {
	tests := []struct {
		env, arg, expected string
	}{
		{skip, skip, app.DefaultBaseURL},
		{skip, sampleBaseURL, sampleBaseURL},
		{sampleBaseURL, skip, sampleBaseURL},
		{anotherBaseURL, sampleBaseURL, anotherBaseURL},
	}

	for i, t := range tests {
		name := fmt.Sprintf("base URL %d", i+1)
		ms.Run(name, func() {
			if t.env != skip {
				os.Setenv(baseURLSetting.envName, t.env)
			}
			if t.arg != skip {
				os.Args = append(os.Args, "-b", t.arg)
			}
			ms.server.start()

			ms.Equal(app.DefaultServerAddress, ms.server.addr())
			ms.shorten(sampleURL, t.expected)
		})
	}
}

func (ms *MainSuite) TestFileStoragePath() {
	tests := []struct {
		env, arg, expected string
	}{
		{skip, skip, app.DefaultStoragePath},
		{skip, samplePath, samplePath},
		{samplePath, skip, samplePath},
		{anotherPath, samplePath, anotherPath},
	}

	for i, t := range tests {
		name := fmt.Sprintf("storage path %d", i+1)
		ms.Run(name, func() {
			if t.env != skip {
				os.Setenv(storagePathSetting.envName, t.env)
			}
			if t.arg != skip {
				os.Args = append(os.Args, "-f", t.arg)
			}
			ms.checkFileStorage(app.DefaultServerAddress, app.DefaultBaseURL, t.expected)
		})
	}
}

func (ms *MainSuite) TestInMemStorage() {
	tests := []struct {
		env, arg string
	}{
		{skip, ""},
		{"", skip},
		{"", ""},
		{"", app.DefaultStoragePath},
	}

	for i, t := range tests {
		name := fmt.Sprintf("storage path %d", i+1)
		ms.Run(name, func() {
			if t.env != skip {
				os.Setenv(storagePathSetting.envName, t.env)
			}
			if t.arg != skip {
				os.Args = append(os.Args, "-f", t.arg)
			}
			ms.checkURLNotKeptAfterRestart()
		})
	}
}

func (ms *MainSuite) checkURLNotKeptAfterRestart() {
	ms.server.start()

	ms.Equal(app.DefaultServerAddress, ms.server.addr())
	key := ms.shorten(sampleURL, app.DefaultBaseURL)

	ms.server.stop()
	ms.NoFileExists(app.DefaultStoragePath)

	ms.server.start()

	resp, _ := ms.httpGet("/" + key)
	defer resp.Body.Close()
	ms.Equal(http.StatusNotFound, resp.StatusCode, "Response status code")
}

func (ms *MainSuite) TestCommandArgs() {
	os.Args = []string{"", "-a", sampleAddr, "-b", sampleBaseURL, "-f", samplePath}

	ms.checkFileStorage(sampleAddr, sampleBaseURL, samplePath)
	ms.NoFileExists(app.DefaultStoragePath)
}

func (ms *MainSuite) TestEnvVars() {
	os.Setenv(addrSetting.envName, sampleAddr)
	os.Setenv(baseURLSetting.envName, sampleBaseURL)
	os.Setenv(storagePathSetting.envName, samplePath)

	ms.checkFileStorage(sampleAddr, sampleBaseURL, samplePath)
	ms.NoFileExists(app.DefaultStoragePath)
}

func (ms *MainSuite) checkFileStorage(addr string, baseURL string, filePath string) {
	ms.server.start()

	ms.Equal(addr, ms.server.addr())
	key := ms.shorten(sampleURL, baseURL)

	ms.server.stop()
	ms.FileExists(filePath)

	ms.server.start()

	url := ms.lookUp(key)
	ms.Equal(sampleURL, url)

	key2 := ms.shorten(anotherURL, baseURL)
	ms.NotEqual(key, key2, "duplicate key")
}

func (ms *MainSuite) TestHelp() {
	os.Args = append(os.Args, "-h")

	main()

	ms.Nil(ms.server.server)
}

func (ms *MainSuite) shorten(url, baseURL string) string {
	body := ms.callShortener(url)

	return ms.extractKey(baseURL, body)
}

func (ms *MainSuite) extractKey(baseURL string, body string) string {
	ms.Regexp("^"+baseURL, body, "Redirect URL")

	idx := strings.LastIndex(body, "/")
	key := body[idx+1:]
	ms.GreaterOrEqual(len(key), 6, "Expected key length to be at least 6")

	return key
}

func (ms *MainSuite) callShortener(url string) string {
	resp := ms.post("text/plain", url)
	defer resp.Body.Close()

	ms.Equal(http.StatusCreated, resp.StatusCode, "Response status code")

	return ms.readBody(resp.Body)
}

func (ms *MainSuite) readBody(body io.ReadCloser) string {
	defer body.Close()

	bytes, err := io.ReadAll(body)
	ms.Require().NoError(err, "Failed to read response body")

	return string(bytes)
}

func (ms *MainSuite) post(contentType string, body string) *http.Response {
	resp, err := http.Post(ms.server.baseURL, contentType, strings.NewReader(body))
	ms.Require().NoError(err, "Failed to POST")

	return resp
}

func (ms *MainSuite) lookUp(shortURL string) string {
	resp, _ := ms.httpGet("/" + shortURL)
	defer resp.Body.Close()

	ms.Equal(http.StatusTemporaryRedirect, resp.StatusCode, "Response status code")

	loc := resp.Header.Get("Location")
	ms.NotEmpty(loc, "Expected Location header to be set")

	return loc
}

func (ms *MainSuite) httpGet(query string) (*http.Response, string) {
	resp, err := http.Get(ms.server.baseURL + query)
	ms.Require().NoError(err, "Failed to make request")
	body := ms.readBody(resp.Body)

	return resp, body
}
