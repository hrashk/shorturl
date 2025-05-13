package main

import (
	"net/http"

	"github.com/hrashk/shorturl/internal/app"
)

func main() {
	err := http.ListenAndServe(`:8080`, app.InMemoryHandler())

	if err != nil {
		panic(err)
	}
}
