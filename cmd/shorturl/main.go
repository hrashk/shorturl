package main

import (
	"net/http"

	"github.com/hrashk/shorturl/internal/app"
)

func main() {
	err := http.ListenAndServe(`:8080`, app.NewHandler())

	if err != nil {
		panic(err)
	}
}
