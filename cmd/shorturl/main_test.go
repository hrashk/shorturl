package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

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
	server   *http.Server
}

func TestMainSuite(t *testing.T) {
	suite.Run(t, &MainSuite{})
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
	ms.origArgs = os.Args

	http.DefaultClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

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

	os.Unsetenv(serverAddressEnv)
	os.Unsetenv(baseURLEnv)
	os.Unsetenv(fileStoragePathEnv)

	if ms.server != nil {
		ms.server.Close()
	}
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
				os.Setenv(serverAddressEnv, t.env)
			}
			if t.arg != skip {
				os.Args = append(os.Args, "-a", t.arg)
			}
			ms.startServer()

			ms.Equal(t.expected, ms.server.Addr)
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
				os.Setenv(baseURLEnv, t.env)
			}
			if t.arg != skip {
				os.Args = append(os.Args, "-b", t.arg)
			}
			ms.startServer()

			ms.Equal(app.DefaultServerAddress, ms.server.Addr)
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
				os.Setenv(fileStoragePathEnv, t.env)
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
				os.Setenv(fileStoragePathEnv, t.env)
			}
			if t.arg != skip {
				os.Args = append(os.Args, "-f", t.arg)
			}
			ms.checkURLNotKeptAfterRestart()
		})
	}
}

func (ms *MainSuite) checkURLNotKeptAfterRestart() {
	ms.startServer()

	ms.Equal(app.DefaultServerAddress, ms.server.Addr)
	key := ms.shorten(sampleURL, app.DefaultBaseURL)

	ms.server.Close()
	ms.NoFileExists(app.DefaultStoragePath)

	ms.startServer()

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
	os.Setenv(serverAddressEnv, sampleAddr)
	os.Setenv(baseURLEnv, sampleBaseURL)
	os.Setenv(fileStoragePathEnv, samplePath)

	ms.checkFileStorage(sampleAddr, sampleBaseURL, samplePath)
	ms.NoFileExists(app.DefaultStoragePath)
}

func (ms *MainSuite) checkFileStorage(addr string, baseURL string, filePath string) {
	ms.startServer()

	ms.Equal(addr, ms.server.Addr)
	key := ms.shorten(sampleURL, baseURL)

	ms.server.Close()
	ms.FileExists(filePath)

	ms.startServer()

	url := ms.lookUp(key)
	ms.Equal(sampleURL, url)

	key2 := ms.shorten(anotherURL, baseURL)
	ms.NotEqual(key, key2, "duplicate key")
}

func (ms *MainSuite) TestHelp() {
	os.Args = append(os.Args, "-h")

	srv, _ := startServer()

	ms.Nil(srv)
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
	resp, err := http.Post(ms.serverAddress(), contentType, strings.NewReader(body))
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
	resp, err := http.Get(ms.serverAddress() + query)
	ms.Require().NoError(err, "Failed to make request")
	body := ms.readBody(resp.Body)

	return resp, body
}

func (ms *MainSuite) serverAddress() string {
	if ms.server.Addr[0] == ':' {
		return "http://localhost" + ms.server.Addr
	} else if !strings.Contains(ms.server.Addr, "://") {
		return "http://" + ms.server.Addr
	} else {
		return ms.server.Addr
	}
}

func (ms *MainSuite) startServer() {
	ms.server, _ = startServer()
	ms.Require().NotNil(ms.server, "failed to start server")

	ms.waitForPort()
}

func (ms *MainSuite) waitForPort() {
	const timeout = time.Second
	const pollInterval = 50 * time.Millisecond
	addr := ms.server.Addr

	var timer = time.NewTimer(timeout)
	var ticker = time.NewTicker(pollInterval)

	for {
		select {
		case <-timer.C:
			ms.Require().Fail("timed out connecting to " + addr)
			return
		case <-ticker.C:
			conn, err := net.DialTimeout("tcp", addr, timeout)
			if err == nil {
				conn.Close() // the port is open
				return
			}
		}
	}
}
