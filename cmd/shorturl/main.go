package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/hrashk/shorturl/internal/app"
)

func main() {
	server := buildServer()

	err := server.ListenAndServe()

	if err != nil {
		panic(err)
	}
}

func buildServer() *http.Server {
	readConfig()

	server := &http.Server{Addr: app.GetListenAddr(), Handler: app.NewHandler()}
	return server
}

func readConfig() {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	var argListenAddr = fs.String("a", app.DefaultServerAddress, "HTTP listen address")
	var argRedirectBaseURL = fs.String("b", app.DefaultBaseURL, "Base URL for redirects")

	fs.Parse(os.Args[1:])

	listenAddr := argOrEnv(argListenAddr, "SERVER_ADDRESS")
	redirectBaseURL := argOrEnv(argRedirectBaseURL, "BASE_URL")

	app.SetListenAddr(listenAddr)
	app.SetRedirectBaseURL(redirectBaseURL)
}

func argOrEnv(argValue *string, envName string) string {
	if envValue, ok := os.LookupEnv(envName); ok {
		return envValue
	} else {
		return *argValue
	}
}
