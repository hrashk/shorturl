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
	return NewHandlerWithLogger(NewZeroLogger())
}

func NewHandlerWithLogger(logger logger) http.Handler {
	c := NewInMemoryController()

	r := chi.NewRouter()
	r.Use(loggingMiddleware(logger))
	r.Use(newGzipDeflator())

	r.Get("/{key}", c.RedirectToOriginalURL)
	r.Post("/", c.CreateShortURL)
	r.Post("/api/shorten", c.ShortenAPI)
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Operation is not supported", http.StatusBadRequest)
	})

	return r
}
