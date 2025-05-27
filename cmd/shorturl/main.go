package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/hrashk/shorturl/internal/app"
)

func main() {
	readConfig()

	err := http.ListenAndServe(app.GetListenAddr(), app.NewHandler())

	if err != nil {
		panic(err)
	}
}

func readConfig() {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	var argListenAddr = fs.String("a", app.GetListenAddr(), "HTTP listen address")
	var argRedirectBaseURL = fs.String("b", app.GetRedirectBaseURL(), "Base URL for redirects")

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
