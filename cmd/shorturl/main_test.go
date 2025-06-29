package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/hrashk/shorturl/internal/app"
	"github.com/stretchr/testify/assert"
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
	srv      *mainServer
	cli      app.Client
}

func TestMainSuite(t *testing.T) {
	suite.Run(t, &MainSuite{})
}

func (ms *MainSuite) SetupSuite() {
	ms.origArgs = os.Args
	ms.srv = newServer(ms.T())
	ms.cli = app.NewClient(ms.T())
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
	ms.srv.wipeData()

	// avoid errors due to unknown flags from go test
	os.Args = []string{os.Args[0]}
}

func (ms *MainSuite) tearDown() {
	os.Args = ms.origArgs

	os.Unsetenv(addrSetting.envName)
	os.Unsetenv(baseURLSetting.envName)
	os.Unsetenv(storagePathSetting.envName)
	os.Unsetenv(dbSetting.envName)

	ms.srv.stop()
}

func (ms *MainSuite) startServer(expectedAddr string) {
	ms.srv.start()
	ms.cli.BaseURL = ms.srv.baseURL
	ms.Equal(expectedAddr, ms.srv.addr())
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
			ms.startServer(t.expected)

			ms.cli.Shorten(sampleURL, app.DefaultBaseURL)
			ms.cli.Ping()
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
			ms.startServer(app.DefaultServerAddress)

			ms.cli.Shorten(sampleURL, t.expected)
			ms.cli.Ping()
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
			ms.checkDataRestoredAfterRestart(app.DefaultServerAddress, app.DefaultBaseURL)
			ms.FileExists(t.expected)
			ms.cli.Ping()
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
			ms.checkDataWipedAfterRestart()
			ms.NoFileExists(app.DefaultStoragePath)
			ms.cli.Ping()
		})
	}
}

func (ms *MainSuite) TestCommandArgsWithFileStorage() {
	os.Args = []string{"", "-a", sampleAddr, "-b", sampleBaseURL,
		"-f", samplePath, "-d", ""}

	ms.checkDataRestoredAfterRestart(sampleAddr, sampleBaseURL)
	ms.NoFileExists(app.DefaultStoragePath)
	ms.FileExists(samplePath)
	ms.cli.Ping()
}

func (ms *MainSuite) TestCommandArgsWithDB() {
	os.Args = []string{"", "-a", sampleAddr, "-b", sampleBaseURL,
		"-f", samplePath, "-d", app.DefaultDatabaseDsn}

	ms.checkDataRestoredAfterRestart(sampleAddr, sampleBaseURL)
	ms.NoFileExists(app.DefaultStoragePath)
	ms.NoFileExists(samplePath)
	ms.cli.Ping()
}

func (ms *MainSuite) TestEnvVarsWithFileStorage() {
	os.Setenv(addrSetting.envName, sampleAddr)
	os.Setenv(baseURLSetting.envName, sampleBaseURL)
	os.Setenv(storagePathSetting.envName, samplePath)
	os.Setenv(dbSetting.envName, "")

	ms.checkDataRestoredAfterRestart(sampleAddr, sampleBaseURL)
	ms.NoFileExists(app.DefaultStoragePath)
	ms.FileExists(samplePath)
	ms.cli.Ping()
}

func (ms *MainSuite) TestEnvVarsWithDB() {
	os.Setenv(addrSetting.envName, sampleAddr)
	os.Setenv(baseURLSetting.envName, sampleBaseURL)
	os.Setenv(storagePathSetting.envName, samplePath)
	os.Setenv(dbSetting.envName, app.DefaultDatabaseDsn)

	ms.checkDataRestoredAfterRestart(sampleAddr, sampleBaseURL)
	ms.NoFileExists(app.DefaultStoragePath)
	ms.NoFileExists(samplePath)
	ms.cli.Ping()
}

func (ms *MainSuite) TestHelp() {
	os.Args = append(os.Args, "-h")

	main()

	ms.Nil(ms.srv.server)
}

func (ms *MainSuite) TestBatchShortener() {
	payload := app.BatchRequest{
		{
			CorrelationID: "123e4567-e89b-12d3-a456-426614174000",
			OriginalURL:   "https://example.com/path1",
		},
		{
			CorrelationID: "123e4567-e89b-12d3-a457-426614174000",
			OriginalURL:   "https://example.com/path2",
		},
		{
			CorrelationID: "123e4567-e89b-12d3-a458-426614174000",
			OriginalURL:   "https://example.com/path3",
		},
		{
			CorrelationID: "123e4567-e89b-12d3-a459-426614174000",
			OriginalURL:   "https://example.com/path4",
		},
	}

	tests := []struct {
		flag, value, expected string
	}{
		{"-f", "", app.DefaultStoragePath},
		{skip, skip, samplePath},
		{"-d", app.DefaultDatabaseDsn, samplePath},
	}

	for i, t := range tests {
		name := fmt.Sprintf("batch %d", i+1)
		ms.Run(name, func() {
			if t.flag != skip {
				os.Args = append(os.Args, t.flag, t.value)
			}
			ms.startServer(app.DefaultServerAddress)

			resp := ms.cli.Batch(payload)
			ms.Require().Equal(len(payload), len(resp))

			for i, req := range payload {
				ms.Equal(req.CorrelationID, resp[i].CorrelationID)

				original := ms.cli.LookUpByURL(resp[i].ShortURL)
				ms.Equal(req.OriginalURL, original)
			}
		})
	}
}

func (ms *MainSuite) TestDuplicateOriginalURL() {
	os.Args = append(os.Args, "-d", app.DefaultDatabaseDsn)
	ms.startServer(app.DefaultServerAddress)

	key := ms.cli.Shorten(sampleAddr, app.DefaultBaseURL)
	key2 := ms.cli.ShortenConflict(sampleAddr, app.DefaultBaseURL)
	assert.Equal(ms.T(), key, key2)
}

func (ms *MainSuite) TestDuplicateOriginalURLAPI() {
	os.Args = append(os.Args, "-d", app.DefaultDatabaseDsn)
	ms.startServer(app.DefaultServerAddress)

	key := ms.cli.ShortenAPI(sampleAddr, app.DefaultBaseURL)
	key2 := ms.cli.ShortenAPIConflict(sampleAddr, app.DefaultBaseURL)
	assert.Equal(ms.T(), key, key2)
}

func (ms *MainSuite) checkDataRestoredAfterRestart(addr, baseURL string) {
	ms.startServer(addr)
	key := ms.cli.Shorten(sampleURL, baseURL)

	ms.srv.stop()

	ms.startServer(addr)
	url := ms.cli.LookUp(key)
	ms.Equal(sampleURL, url)

	key2 := ms.cli.Shorten(anotherURL, baseURL)
	ms.NotEqual(key, key2, "duplicate key")
}

func (ms *MainSuite) checkDataWipedAfterRestart() {
	ms.startServer(app.DefaultServerAddress)
	key := ms.cli.Shorten(sampleURL, app.DefaultBaseURL)

	ms.srv.stop()

	ms.startServer(app.DefaultServerAddress)
	ms.cli.LookUpNotFound(key)
}
