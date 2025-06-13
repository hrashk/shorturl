package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/hrashk/shorturl/internal/app"
)

const (
	serverAddressEnv   = "SERVER_ADDRESS"
	baseURLEnv         = "BASE_URL"
	fileStoragePathEnv = "FILE_STORAGE_PATH"
)

func main() {
	_, ch := startServer()

	if err := <-ch; err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}
}

func startServer() (*http.Server, chan error) {
	// do not block in case of error
	ch := make(chan error, 1)

	server, err := buildServer()
	if err != nil || server == nil {
		ch <- err
		return nil, ch
	}

	go func() {
		ch <- server.ListenAndServe()
	}()

	return server, ch
}

func buildServer() (*http.Server, error) {
	mods, err := readConfig()

	if errors.Is(err, flag.ErrHelp) {
		return nil, nil // do not start server when asked for help
	} else if err != nil {
		return nil, fmt.Errorf("failed to read configuration: %w", err)
	}

	return app.NewServer(mods...)
}

func readConfig() ([]app.Configurator, error) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	var argAddr = fs.String("a", app.DefaultServerAddress,
		fmt.Sprintf("HTTP listen address. Alternatively, use %s environment variable.", serverAddressEnv))
	var argBaseURL = fs.String("b", app.DefaultBaseURL,
		fmt.Sprintf("Base URL for redirects. Alternatively, use %s environment variable.", baseURLEnv))
	var argFilePath = fs.String("f", app.DefaultStoragePath,
		fmt.Sprintf("File storage path. Alternatively, use %s environment variable.", fileStoragePathEnv))

	if err := fs.Parse(os.Args[1:]); err != nil {
		return nil, fmt.Errorf("failed to parse command arguments: %w", err)
	}

	addr := argOrEnv(argAddr, serverAddressEnv)
	baseURL := argOrEnv(argBaseURL, baseURLEnv)
	filePath := argOrEnv(argFilePath, fileStoragePathEnv)

	return []app.Configurator{
		app.WithServerAddress(addr),
		app.WithBaseURL(baseURL),
		app.WithStoragePath(filePath),
	}, nil
}

func argOrEnv(argValue *string, envName string) string {
	if envValue, ok := os.LookupEnv(envName); ok {
		return envValue
	} else {
		return *argValue
	}
}
