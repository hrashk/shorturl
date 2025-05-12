package app

import (
	"io"
	"net/http"
)

type ShortKeyGenerator interface {
	Generate(url string) (key string)
}

type ShortURLController struct {
	KeyGenerator ShortKeyGenerator
}

func NewShortURLController(keyGenerator ShortKeyGenerator) *ShortURLController {
	return &ShortURLController{keyGenerator}
}

func (c *ShortURLController) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	raw, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	url := string(raw)
	shortUrl := c.KeyGenerator.Generate(url)

	w.WriteHeader(http.StatusCreated)
	io.WriteString(w, "http://localhost:8080/"+shortUrl)
}

func (c *ShortURLController) RedirectToOriginalURL(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://pkg.go.dev/cmp", http.StatusTemporaryRedirect)
}

func (c *ShortURLController) RouteRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		c.CreateShortURL(w, r)
	} else if r.Method == http.MethodGet && len(r.URL.Path) > 1 {
		c.RedirectToOriginalURL(w, r)
	} else {
		http.Error(w, "Operation is not supported", http.StatusBadRequest)
	}
}
