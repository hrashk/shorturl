package main

import (
	"io"
	"net/http"
)

func mainPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		w.WriteHeader(http.StatusCreated)
		io.WriteString(w, "http://localhost:8080/EwHXdJfB")
	} else if r.Method == http.MethodGet && len(r.URL.Path) > 1 {
		w.Header().Set("Location", "https://pkg.go.dev/cmp")
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		http.Error(w, "Operation is not supported", http.StatusBadRequest)
	}
}

func main() {
	err := http.ListenAndServe(`:8080`, http.HandlerFunc(mainPage))
	if err != nil {
		panic(err)
	}
}
