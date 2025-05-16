package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewInMemoryController() ShortURLController {
	s := NewShortURLService(NewBase62Generator(), NewInMemStorage())

	return NewShortURLController(s)
}

func NewHandler() http.Handler {
	c := NewInMemoryController()

	r := chi.NewRouter()
	r.Get("/{key}", c.RedirectToOriginalURL)
	r.Post("/", c.CreateShortURL)
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Operation is not supported", http.StatusBadRequest)
	})

	return r
}
