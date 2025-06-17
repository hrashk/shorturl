package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/hrashk/shorturl/internal/app"
)

type setting struct {
	name     string
	usage    string
	defValue string
	envName  string
	cfg      func(string) app.Configurator
	value    *string
}

var addrSetting = setting{
	name:     "a",
	usage:    "the address HTTP server listens at. Related environment variable %s has higher priority.",
	defValue: app.DefaultServerAddress,
	envName:  "SERVER_ADDRESS",
	cfg:      app.WithServerAddress,
}

var baseURLSetting = setting{
	name:     "b",
	usage:    "base URL for redirects. Related environment variable %s has higher priority.",
	defValue: app.DefaultBaseURL,
	envName:  "BASE_URL",
	cfg:      app.WithBaseURL,
}

var storagePathSetting = setting{
	name: "f",
	usage: "file storage path. " +
		"Empty value means storage is in-memory only. " +
		"Related environment variable %s has higher priority.",
	defValue: app.DefaultStoragePath,
	envName:  "FILE_STORAGE_PATH",
	cfg:      app.WithStoragePath,
}

var dbSetting = setting{
	name: "d",
	usage: "database connection string. " +
		"Empty value falls back to final storage. " +
		"Related environment variable %s has higher priority.",
	defValue: app.DefaultDatabaseDsn,
	envName:  "DATABASE_DSN",
	cfg:      app.WithDatabaseDsn,
}

type settings struct {
	fs  *flag.FlagSet
	all []*setting
}

func newSettings() settings {
	ss := settings{
		fs:  flag.NewFlagSet(os.Args[0], flag.ContinueOnError),
		all: []*setting{&addrSetting, &baseURLSetting, &storagePathSetting, &dbSetting},
	}
	ss.declareAll()

	return ss
}

func (ss settings) declareAll() {
	for _, t := range ss.all {
		t.value =
			ss.fs.String(t.name, t.defValue, fmt.Sprintf(t.usage, t.envName))
	}
}

// mods == nil && err == nil when there is no need to start the server,
// e.g. when help is requested
func (ss settings) parse() (mods []app.Configurator, err error) {
	if err = ss.fs.Parse(os.Args[1:]); errors.Is(err, flag.ErrHelp) {
		err = nil
		return
	} else if err != nil {
		return nil, fmt.Errorf("failed to parse command arguments: %w", err)
	}

	for _, t := range ss.all {
		value := argOrEnv(t.value, t.envName)
		mods = append(mods, t.cfg(value))
	}

	return
}

func argOrEnv(argValue *string, envName string) string {
	if envValue, ok := os.LookupEnv(envName); ok {
		return envValue
	} else {
		return *argValue
	}
}
