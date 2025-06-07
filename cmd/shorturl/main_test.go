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

const sampleURL = "https://pkg.go.dev/cmp"
const skip = "#"

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

	// avoid errors due to unknown flags from go test
	os.Args = []string{os.Args[0]}
}

func (ms *MainSuite) tearDown() {
	os.Args = ms.origArgs

	app.SetListenAddr(app.DefaultServerAddress)
	app.SetRedirectBaseURL(app.DefaultBaseURL)

	os.Unsetenv("SERVER_ADDRESS")
	os.Unsetenv("BASE_URL")

	if ms.server != nil {
		ms.server.Close()
	}
}

func (ms *MainSuite) TestServerAddress() {
	tests := []struct {
		env, arg, expected string
	}{
		{skip, skip, app.DefaultServerAddress},
		{skip, "localhost:8088", "localhost:8088"},
		{"localhost:8088", skip, "localhost:8088"},
		{"localhost:9099", "localhost:8088", "localhost:9099"},
	}

	for i, t := range tests {
		name := fmt.Sprintf("server address %d", i+1)
		ms.Run(name, func() {
			if t.env != skip {
				os.Setenv("SERVER_ADDRESS", t.env)
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

func (ms *MainSuite) TestCommandArgs() {
	const listen = "localhost:8088"
	const redirect = "http://example.com:1024"
	os.Args = []string{"", "-a", listen, "-b", redirect}

	ms.startServer()

	ms.Equal(listen, ms.server.Addr)
	ms.shorten(sampleURL, redirect)
}

func (ms *MainSuite) TestHelp() {
	os.Args = append(os.Args, "-h")
	srv, err := buildServer()
	ms.NoError(err)
	ms.Nil(srv)
}

func (ms *MainSuite) Test_readConfigWithDefaultListenAddress() {
	const baseURL = "http://example.com:1024"
	os.Args = []string{"", "-b", baseURL}

	ms.startServer()

	ms.Equal(app.DefaultServerAddress, ms.server.Addr)
	ms.shorten(sampleURL, baseURL)
}

func (ms *MainSuite) Test_readConfigWithEnvServerAddress() {
	const envListen = "127.0.0.1:8088"
	os.Setenv("SERVER_ADDRESS", envListen)

	const argListen = "localhost:8099"
	const argRedirect = "http://example.com:1024"
	os.Args = []string{"", "-a", argListen, "-b", argRedirect}

	ms.startServer()

	ms.Equal(envListen, ms.server.Addr)
	ms.shorten(sampleURL, argRedirect)
}

func (ms *MainSuite) Test_readConfigWithEnvRedirectURL() {
	const envRedirect = "http://random.com:4201"
	os.Setenv("BASE_URL", envRedirect)

	const argListen = "localhost:8099"
	const argRedirect = "http://example.com:1024"
	os.Args = []string{"", "-a", argListen, "-b", argRedirect}

	ms.startServer()

	ms.Equal(argListen, ms.server.Addr)
	ms.shorten(sampleURL, envRedirect)
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

	ms.Equal(http.StatusCreated, resp.StatusCode, "Response status code")

	return ms.readBody(resp)
}

func (ms *MainSuite) readBody(resp *http.Response) string {
	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)
	ms.Require().NoError(err, "Failed to read response body")

	return string(bytes)
}

func (ms *MainSuite) post(contentType string, body string) *http.Response {
	resp, err := http.Post(ms.serverAddress(), contentType, strings.NewReader(body))
	ms.Require().NoError(err, "Failed to POST")

	return resp
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
	srv, err := buildServer()
	ms.Require().NoError(err, "Failed building server")

	ms.server = srv
	go ms.server.ListenAndServe()

	ms.waitForPort()
}

func (ms *MainSuite) waitForPort() {
	const timeout = time.Second
	const pollInterval = 50 * time.Millisecond
	addr := ms.server.Addr

	start := time.Now()
	for {
		conn, err := net.DialTimeout("tcp", addr, timeout)
		if err == nil {
			conn.Close()
			return // Port is open
		}
		if time.Since(start) > timeout {
			ms.Require().Fail("timed out connecting to " + addr)
			return
		}
		time.Sleep(pollInterval)
	}
}
