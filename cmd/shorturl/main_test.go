package main

import (
	"os"
	"testing"

	"github.com/hrashk/shorturl/internal/app"
	"github.com/stretchr/testify/suite"
)

type MainSuite struct {
	suite.Suite
	origArgs []string
}

func TestMainSuite(t *testing.T) {
	suite.Run(t, &MainSuite{})
}

func (suite *MainSuite) SetupTest() {
	suite.origArgs = os.Args
}

func (suite *MainSuite) TearDownTest() {
	os.Args = suite.origArgs

	app.SetListenAddr(app.DefaultServerAddress)
	app.SetRedirectBaseURL(app.DefaultBaseURL)

	os.Unsetenv("SERVER_ADDRESS")
	os.Unsetenv("BASE_URL")
}

func (suite *MainSuite) Test_readConfig() {
	const listen = "localhost:9999"
	const redirect = "http://example.com:1024"
	os.Args = []string{"", "-a", listen, "-b", redirect}

	server := buildServer()

	suite.Equal(listen, server.Addr)
	suite.Equal(redirect, app.GetRedirectBaseURL())
}

func (suite *MainSuite) Test_readConfigWithDefaultListenAddress() {
	const redirect = "http://example.com:1024"
	os.Args = []string{"", "-b", redirect}

	server := buildServer()

	suite.Equal(app.DefaultServerAddress, server.Addr)
	suite.Equal(redirect, app.GetRedirectBaseURL())
}

func (suite *MainSuite) Test_readConfigWithDefaultRedirectAddress() {
	const listen = "localhost:9999"
	os.Args = []string{"", "-a", listen}

	server := buildServer()

	suite.Equal(listen, server.Addr)
	suite.Equal(app.DefaultBaseURL, app.GetRedirectBaseURL())
}

func (suite *MainSuite) Test_readConfig_Defaults() {
	os.Args = []string{""}

	server := buildServer()

	suite.Equal(app.DefaultServerAddress, server.Addr)
	suite.Equal(app.DefaultBaseURL, app.GetRedirectBaseURL())
}

func (suite *MainSuite) Test_readConfigWithEnvServerAddress() {
	const envListen = "127.0.0.1:8088"
	os.Setenv("SERVER_ADDRESS", envListen)

	const argListen = "localhost:9999"
	const argRedirect = "http://example.com:1024"
	os.Args = []string{"", "-a", argListen, "-b", argRedirect}

	server := buildServer()

	suite.Equal(envListen, server.Addr)
	suite.Equal(argRedirect, app.GetRedirectBaseURL())
}

func (suite *MainSuite) Test_readConfigWithEnvRedirectURL() {
	const envRedirect = "http://random.com:4201"
	os.Setenv("BASE_URL", envRedirect)

	const argListen = "localhost:9999"
	const argRedirect = "http://example.com:1024"
	os.Args = []string{"", "-a", argListen, "-b", argRedirect}

	server := buildServer()

	suite.Equal(argListen, server.Addr)
	suite.Equal(envRedirect, app.GetRedirectBaseURL())
}
