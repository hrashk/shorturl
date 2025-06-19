package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type adapter struct {
	svc service
	log logger
}

func newAdapter(cfg config) (adapter, error) {
	s, err := newService(cfg)

	return adapter{s, cfg.log}, err
}

func (a adapter) handler() http.Handler {
	r := chi.NewRouter()
	r.Use(loggingMiddleware(a.log))
	r.Use(newGzipDeflator())
	r.Use(newGzipInflator())

	r.Get("/ping", a.Ping)
	r.Get("/{key}", a.RedirectToOriginalURL)
	r.Post("/", a.CreateShortURL)
	r.Post("/api/shorten", a.ShortenAPI)
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Operation is not supported", http.StatusBadRequest)
	})

	return r
}

func (a adapter) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	raw, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, fmt.Sprintf("failed to read request body: %v", err), http.StatusBadRequest)
		return
	}

	url := string(raw)
	shortURL, err := a.svc.CreateShortURL(r.Context(), url)

	if err != nil {
		http.Error(w, fmt.Sprintf("failed to store URL: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	io.WriteString(w, shortURL)
}

func (a adapter) RedirectToOriginalURL(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[1:]

	url, err := a.svc.LookUp(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

type ShortURLRequest struct {
	URL string `json:"url"`
}

type ShortURLResponse struct {
	Result string `json:"result"`
}

func (a adapter) ShortenAPI(w http.ResponseWriter, r *http.Request) {
	var req ShortURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("failed to read body: %v", err), http.StatusBadRequest)
		return
	}

	shortURL, err := a.svc.CreateShortURL(r.Context(), req.URL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to store URL: %v", err), http.StatusInternalServerError)
		return
	}

	resp := ShortURLResponse{shortURL}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err = json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, fmt.Sprintf("failed to write response: %v", err), http.StatusInternalServerError)
		return
	}
}

func (a adapter) Ping(w http.ResponseWriter, r *http.Request) {
	if err := a.svc.PingDB(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
