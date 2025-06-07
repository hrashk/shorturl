package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/hrashk/shorturl/internal/app"
)

func main() {
	server, err := buildServer()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to build server: %v\n", err)
		os.Exit(2)
	} else if server == nil { // in case help was requested
		return
	}

	err = server.ListenAndServe()

	if err != nil {
		panic(err)
	}
}

func buildServer() (*http.Server, error) {
	if err := readConfig(); errors.Is(err, flag.ErrHelp) {
		return nil, nil // do not start server when asked for help
	} else if err != nil {
		return nil, err
	}

	return &http.Server{Addr: app.GetListenAddr(), Handler: app.NewHandler()}, nil
}

func readConfig() error {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	var argListenAddr = fs.String("a", app.DefaultServerAddress, "HTTP listen address")
	var argRedirectBaseURL = fs.String("b", app.DefaultBaseURL, "Base URL for redirects")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return fmt.Errorf("failed to parse command arguments: %w", err)
	}

	listenAddr := argOrEnv(argListenAddr, "SERVER_ADDRESS")
	redirectBaseURL := argOrEnv(argRedirectBaseURL, "BASE_URL")

	app.SetListenAddr(listenAddr)
	app.SetRedirectBaseURL(redirectBaseURL)

	return nil
}

func argOrEnv(argValue *string, envName string) string {
	if envValue, ok := os.LookupEnv(envName); ok {
		return envValue
	} else {
		return *argValue
	}
}
