package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func newInMemoryController() shortURLController {
	s := newShortURLService(newBase62Generator(), newInMemStorage())

	return newShortURLController(s)
}

func NewHandler() http.Handler {
	return newHandlerWithLogger(newZeroLogger())
}

func newHandlerWithLogger(logger logger) http.Handler {
	c := newInMemoryController()

	r := chi.NewRouter()
	r.Use(loggingMiddleware(logger))
	r.Use(newGzipDeflator())
	r.Use(newGzipInflator())

	r.Get("/{key}", c.RedirectToOriginalURL)
	r.Post("/", c.CreateShortURL)
	r.Post("/api/shorten", c.ShortenAPI)
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Operation is not supported", http.StatusBadRequest)
	})

	return r
}
