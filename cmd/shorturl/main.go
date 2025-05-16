package main

import (
	"flag"
	"net/http"

	"github.com/hrashk/shorturl/internal/app"
)

func main() {
	readConfigFromArgs()

	err := http.ListenAndServe(app.GetListenAddr(), app.NewHandler())

	if err != nil {
		panic(err)
	}
}

func readConfigFromArgs() {
	listenAddr := flag.String("a", app.GetListenAddr(), "HTTP listen address")
	redirectBaseURL := flag.String("b", app.GetRedirectBaseURL(), "Base URL for redirects")
	flag.Parse()
	app.SetListenAddr(*listenAddr)
	app.SetRedirectBaseURL(*redirectBaseURL)
}
