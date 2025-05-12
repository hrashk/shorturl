package app

import (
	"io"
	"net/http"
)

type ShortURLController struct {
	Service Service
}

func NewShortURLController(service Service) ShortURLController {
	return ShortURLController{service}
}

func (c ShortURLController) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	raw, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	url := string(raw)
	shortURL, err := c.Service.CreateShortURL(url)

	if err != nil {
		http.Error(w, "Failed to store URL", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	io.WriteString(w, shortURL)
}

func (c ShortURLController) RedirectToOriginalURL(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[1:]

	url, err := c.Service.LookUp(key)
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
