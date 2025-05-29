package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Logger interface {
	Info(msg string, fields ...any)
}

func NewInMemoryController() ShortURLController {
	s := NewShortURLService(NewBase62Generator(), NewInMemStorage())

	return NewShortURLController(s)
}

func NewHandler() http.Handler {
	return NewHandlerWithLogger(NewZeroLogger())
}

func NewHandlerWithLogger(logger Logger) http.Handler {
	c := NewInMemoryController()

	r := chi.NewRouter()
	r.Use(wrapper{logger}.middleware)

	r.Get("/{key}", c.RedirectToOriginalURL)
	r.Post("/", c.CreateShortURL)
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Operation is not supported", http.StatusBadRequest)
	})

	return r
}

type wrapper struct {
	Logger
}

func (lw wrapper) middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lw.Info("Received request: %s %s", r.Method, r.URL)
		h.ServeHTTP(w, r)
	})
}
