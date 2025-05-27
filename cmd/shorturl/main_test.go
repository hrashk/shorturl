package main

import (
	"os"
	"testing"

	"github.com/hrashk/shorturl/internal/app"
	"github.com/stretchr/testify/suite"
)

type MainSuite struct {
	suite.Suite
	origArgs     []string
	origListen   string
	origRedirect string
}

func TestMainSuite(t *testing.T) {
	suite.Run(t, &MainSuite{})
}

func (suite *MainSuite) SetupTest() {
	suite.origArgs = os.Args
	suite.origListen = app.GetListenAddr()
	suite.origRedirect = app.GetRedirectBaseURL()
}

func (suite *MainSuite) TearDownTest() {
	os.Args = suite.origArgs

	app.SetListenAddr(suite.origListen)
	app.SetRedirectBaseURL(suite.origRedirect)

	os.Unsetenv("SERVER_ADDRESS")
	os.Unsetenv("BASE_URL")
}

func (suite *MainSuite) Test_readConfig() {
	const listen = "localhost:9999"
	const redirect = "http://example.com:1024"
	os.Args = []string{"", "-a", listen, "-b", redirect}

	readConfig()

	suite.Equal(listen, app.GetListenAddr())
	suite.Equal(redirect, app.GetRedirectBaseURL())
}

func (suite *MainSuite) Test_readConfigWithDefaultListenAddress() {
	const redirect = "http://example.com:1024"
	os.Args = []string{"", "-b", redirect}

	readConfig()

	suite.Equal(suite.origListen, app.GetListenAddr())
	suite.Equal(redirect, app.GetRedirectBaseURL())
}

func (suite *MainSuite) Test_readConfigWithDefaultRedirectAddress() {
	const listen = "localhost:9999"
	os.Args = []string{"", "-a", listen}

	readConfig()

	suite.Equal(listen, app.GetListenAddr())
	suite.Equal(suite.origRedirect, app.GetRedirectBaseURL())
}

func (suite *MainSuite) Test_readConfig_Defaults() {
	os.Args = []string{""}

	readConfig()

	suite.Equal(suite.origListen, app.GetListenAddr())
	suite.Equal(suite.origRedirect, app.GetRedirectBaseURL())
}

func (suite *MainSuite) Test_readConfigWithEnvServerAddress() {
	const envListen = "127.0.0.1:8088"
	os.Setenv("SERVER_ADDRESS", envListen)

	const argListen = "localhost:9999"
	const argRedirect = "http://example.com:1024"
	os.Args = []string{"", "-a", argListen, "-b", argRedirect}

	readConfig()

	suite.Equal(envListen, app.GetListenAddr())
	suite.Equal(argRedirect, app.GetRedirectBaseURL())
}

func (suite *MainSuite) Test_readConfigWithEnvRedirectURL() {
	const envRedirect = "http://random.com:4201"
	os.Setenv("BASE_URL", envRedirect)

	const argListen = "localhost:9999"
	const argRedirect = "http://example.com:1024"
	os.Args = []string{"", "-a", argListen, "-b", argRedirect}

	readConfig()

	suite.Equal(argListen, app.GetListenAddr())
	suite.Equal(envRedirect, app.GetRedirectBaseURL())
}
