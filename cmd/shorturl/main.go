package main

import (
	"net/http"

	"github.com/hrashk/shorturl/internal/app"
)

func main() {
	c := app.NewShortURLController(app.NewBase62Generator())

	err := http.ListenAndServe(`:8080`, http.HandlerFunc(c.RouteRequest))
	if err != nil {
		panic(err)
	}
}
