package app

import (
	"io"
	"net/http"
)

type ShortKeyGenerator interface {
	Generate(url string) (key string)
}

type Storage interface {
	Store(key string, url string) error
	LookUp(key string) (url string, err error)
}

type ShortURLController struct {
	KeyGenerator ShortKeyGenerator
	Storage      Storage
}

func NewShortURLController(keyGenerator ShortKeyGenerator, storage Storage) ShortURLController {
	return ShortURLController{KeyGenerator: keyGenerator, Storage: storage}
}

func (c ShortURLController) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	raw, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	url := string(raw)
	key := c.KeyGenerator.Generate(url)

	if err := c.Storage.Store(key, url); err != nil {
		http.Error(w, "Failed to store URL", http.StatusInternalServerError)
		return
	}

	shortURL := "http://localhost:8080/" + key

	w.WriteHeader(http.StatusCreated)
	io.WriteString(w, shortURL)
}

func (c ShortURLController) RedirectToOriginalURL(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[1:]

	url, err := c.Storage.LookUp(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (c ShortURLController) RouteRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		c.CreateShortURL(w, r)
	} else if r.Method == http.MethodGet && len(r.URL.Path) > 1 {
		c.RedirectToOriginalURL(w, r)
	} else {
		http.Error(w, "Operation is not supported", http.StatusBadRequest)
	}
}
