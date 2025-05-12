package app

import (
	"io"
	"net/http"
)

type ShortURLController struct{}

func NewShortURLController() *ShortURLController {
	return &ShortURLController{}
}

func (c *ShortURLController) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	io.WriteString(w, "http://localhost:8080/EwHXdJfB")
}

func (c *ShortURLController) RedirectToOriginalURL(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://pkg.go.dev/cmp", http.StatusTemporaryRedirect)
}

func (c *ShortURLController) RouteRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		c.CreateShortURL(w, r)
	} else if r.Method == http.MethodGet && len(r.URL.Path) > 1 {
		c.RedirectToOriginalURL(w, r)
	} else {
		http.Error(w, "Operation is not supported", http.StatusBadRequest)
	}
}
