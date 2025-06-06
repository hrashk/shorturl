package main

import (
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

type MainSuite struct {
	suite.Suite
	origArgs []string
	server   *http.Server
}

func TestMainSuite(t *testing.T) {
	suite.Run(t, &MainSuite{})
}

func (ms *MainSuite) SetupTest() {
	ms.origArgs = os.Args
}

func (ms *MainSuite) TearDownTest() {
	os.Args = ms.origArgs

	app.SetListenAddr(app.DefaultServerAddress)
	app.SetRedirectBaseURL(app.DefaultBaseURL)

	os.Unsetenv("SERVER_ADDRESS")
	os.Unsetenv("BASE_URL")

	if ms.server != nil {
		ms.server.Close()
	}
}

func (ms *MainSuite) Test_readConfig() {
	const listen = "localhost:8088"
	const redirect = "http://example.com:1024"
	os.Args = []string{"", "-a", listen, "-b", redirect}

	ms.startServer()

	ms.Equal(listen, ms.server.Addr)
	ms.invokeShortener(sampleURL, redirect)
}

func (ms *MainSuite) Test_readConfigWithDefaultListenAddress() {
	const baseURL = "http://example.com:1024"
	os.Args = []string{"", "-b", baseURL}

	ms.startServer()

	ms.Equal(app.DefaultServerAddress, ms.server.Addr)
	ms.invokeShortener(sampleURL, baseURL)
}

func (ms *MainSuite) Test_readConfigWithDefaultRedirectAddress() {
	const listen = "localhost:8099"
	os.Args = []string{"", "-a", listen}

	ms.startServer()

	ms.Equal(listen, ms.server.Addr)
	ms.invokeShortener(sampleURL, app.DefaultBaseURL)
}

func (ms *MainSuite) Test_readConfig_Defaults() {
	os.Args = []string{""}

	ms.startServer()

	ms.Equal(app.DefaultServerAddress, ms.server.Addr)
	ms.invokeShortener(sampleURL, app.DefaultBaseURL)
}

func (ms *MainSuite) Test_readConfigWithEnvServerAddress() {
	const envListen = "127.0.0.1:8088"
	os.Setenv("SERVER_ADDRESS", envListen)

	const argListen = "localhost:8099"
	const argRedirect = "http://example.com:1024"
	os.Args = []string{"", "-a", argListen, "-b", argRedirect}

	ms.startServer()

	ms.Equal(envListen, ms.server.Addr)
	ms.invokeShortener(sampleURL, argRedirect)
}

func (ms *MainSuite) Test_readConfigWithEnvRedirectURL() {
	const envRedirect = "http://random.com:4201"
	os.Setenv("BASE_URL", envRedirect)

	const argListen = "localhost:8099"
	const argRedirect = "http://example.com:1024"
	os.Args = []string{"", "-a", argListen, "-b", argRedirect}

	ms.startServer()

	ms.Equal(argListen, ms.server.Addr)
	ms.invokeShortener(sampleURL, envRedirect)
}

func (ms *MainSuite) invokeShortener(url, baseURL string) string {
	resp, err := http.Post(ms.serverAddress(), "text/plain", strings.NewReader(url))
	ms.Require().NoError(err, "Failed to make request")
	defer resp.Body.Close()

	ms.Equal(http.StatusCreated, resp.StatusCode, "Response status code")

	bytes, err := io.ReadAll(resp.Body)
	ms.Require().NoError(err, "Failed to read response body")

	body := string(bytes)

	ms.Regexp("^"+baseURL, body, "Redirect URL")

	idx := strings.LastIndex(body, "/")
	key := body[idx+1:]
	ms.GreaterOrEqual(len(key), 6, "Expected key length to be at least 6")

	return key
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
	ms.server = buildServer()

	go ms.server.ListenAndServe()

	ms.waitForPort(ms.server.Addr)
}

func (ms *MainSuite) waitForPort(address string) {
	start := time.Now()
	timeout := time.Second

	for {
		conn, err := net.DialTimeout("tcp", address, timeout)
		if err == nil {
			conn.Close()
			return // Port is open
		}
		if time.Since(start) > timeout {
			ms.Require().NoError(err, "timed out connecting to "+address)
			return
		}
		time.Sleep(50 * time.Millisecond) // Wait before retrying
	}
}
