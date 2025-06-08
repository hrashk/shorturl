package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type service interface {
	CreateShortURL(url string) (shortURL string, err error)
	LookUp(key string) (url string, err error)
}

type shortURLController struct {
	Service service
}

func (c shortURLController) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	raw, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, fmt.Sprintf("failed to read request body: %v", err), http.StatusBadRequest)
		return
	}

	url := string(raw)
	shortURL, err := c.Service.CreateShortURL(url)

	if err != nil {
		http.Error(w, fmt.Sprintf("failed to store URL: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	io.WriteString(w, shortURL)
}

func (c shortURLController) RedirectToOriginalURL(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[1:]

	url, err := c.Service.LookUp(key)
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

func (c shortURLController) ShortenAPI(w http.ResponseWriter, r *http.Request) {
	var req ShortURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("failed to read body: %v", err), http.StatusBadRequest)
		return
	}

	shortURL, err := c.Service.CreateShortURL(req.URL)
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
