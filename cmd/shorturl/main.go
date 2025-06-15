package main

import (
	"errors"
	"net/http"

	"github.com/hrashk/shorturl/internal/app"
)

func main() {
	mods, err := newSettings().parse()
	if err != nil {
		panic(err)
	} else if mods == nil {
		return
	}

	server, err := app.NewServer(mods...)
	if err != nil {
		panic(err)
	}

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}
}
