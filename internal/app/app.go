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

	return &http.Server{Addr: cfg.serverAddress, Handler: newHandler(cfg)}, nil
}

func newInMemoryController(cfg *config) shortURLController {
	s := newShortURLService(newBase62Generator(), newInMemStorage(), cfg.baseURL)

	return newShortURLController(s)
}

func newHandler(cfg *config) http.Handler {
	c := newInMemoryController(cfg)

	r := chi.NewRouter()
	r.Use(loggingMiddleware(newZeroLogger()))
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
