package app

import (
	"encoding/json"
	"errors"
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
	r.Post("/api/shorten/batch", a.ShortenBatch)
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Operation is not supported", http.StatusBadRequest)
	})

	return r
}

func (a adapter) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	url, err := originalURL(r)
	if err != nil {
		badRequest(w, err)
		return
	}

	shortURL, err := a.svc.CreateShortURL(r.Context(), url)

	if errors.Is(err, ErrConflict) {
		conflict(w, shortURL)
	} else if err != nil {
		serverError(w, err)
	} else {
		created(w, shortURL)
	}
}

func conflict(w http.ResponseWriter, shortURL string) {
	w.WriteHeader(http.StatusConflict)
	io.WriteString(w, shortURL)
}

func created(w http.ResponseWriter, shortURL string) {
	w.WriteHeader(http.StatusCreated)
	io.WriteString(w, shortURL)
}

func serverError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func badRequest(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusBadRequest)
}

func notFound(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusNotFound)
}

func originalURL(r *http.Request) (string, error) {
	raw, err := io.ReadAll(r.Body)

	if err != nil {
		return "", fmt.Errorf("failed to read request body: %w", err)
	}

	return string(raw), nil
}

func (a adapter) RedirectToOriginalURL(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[1:]

	url, err := a.svc.LookUp(r.Context(), key)
	if err != nil {
		notFound(w, err)
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
	req, err := bind(r)
	if err != nil {
		badRequest(w, err)
		return
	}

	shortURL, err := a.svc.CreateShortURL(r.Context(), req.URL)

	if errors.Is(err, ErrConflict) {
		err = conflictAPI(w, shortURL)
	} else if err == nil {
		err = createdAPI(w, shortURL)
	}

	if err != nil {
		serverError(w, err)
	}
}

func bind(r *http.Request) (ShortURLRequest, error) {
	var req ShortURLRequest
	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		err = fmt.Errorf("unable to decode body: %w", err)
	}
	return req, err
}

func createdAPI(w http.ResponseWriter, shortURL string) error {
	resp := ShortURLResponse{shortURL}
	return writeJSON(w, resp, http.StatusCreated)
}

func conflictAPI(w http.ResponseWriter, shortURL string) error {
	resp := ShortURLResponse{shortURL}
	return writeJSON(w, resp, http.StatusConflict)
}

func writeJSON(w http.ResponseWriter, resp any, status int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		err = fmt.Errorf("unable to write response: %w", err)
	}
	return err
}

func (a adapter) Ping(w http.ResponseWriter, r *http.Request) {
	if err := a.svc.PingDB(r.Context()); err != nil {
		serverError(w, err)
	}
}

func (a adapter) ShortenBatch(w http.ResponseWriter, r *http.Request) {
	req, err := bindBatch(r)
	if err != nil {
		badRequest(w, err)
		return
	}

	resp, err := a.svc.ShortenBatch(r.Context(), req)
	if err == nil {
		err = writeJSON(w, resp, http.StatusCreated)
	}

	if err != nil {
		serverError(w, err)
	}
}

func bindBatch(r *http.Request) (BatchRequest, error) {
	var req BatchRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		err = fmt.Errorf("unable to decode batch request: %w", err)
	}
	return req, err
}
