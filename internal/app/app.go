package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewServer(modifiers ...CfgModifier) (*http.Server, error) {
	cfg, err := newConfig(modifiers...)
	if err != nil {
		return nil, err
	}
	h, err := newHandler(cfg)
	if err != nil {
		return nil, err
	}

	return &http.Server{Addr: cfg.serverAddress, Handler: h}, nil
}

func newHandler(cfg *config) (http.Handler, error) {
	s, err := newService(cfg)
	if err != nil {
		return nil, err
	}
	c := shortURLController{s}

	r := chi.NewRouter()
	r.Use(loggingMiddleware(cfg.log))
	r.Use(newGzipDeflator())
	r.Use(newGzipInflator())

	r.Get("/{key}", c.RedirectToOriginalURL)
	r.Post("/", c.CreateShortURL)
	r.Post("/api/shorten", c.ShortenAPI)
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Operation is not supported", http.StatusBadRequest)
	})

	return r, nil
}
