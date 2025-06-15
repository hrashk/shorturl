package main

import (
	"errors"
	"net/http"

	"github.com/hrashk/shorturl/internal/app"
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
	mods, err := newSettings().parse()

	if mods == nil {
		return nil, err
	}

	return app.NewServer(mods...)
}
