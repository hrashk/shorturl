package app

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// New compression algorithms have to implement this interface
// and provide a middleware constructor. Follow the example for gzip below.
type inflator interface {
	isCompressed(r *http.Request) bool
	wrap(rc io.ReadCloser) (io.ReadCloser, error)
}

func inflatorMiddleware(inflator inflator) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if inflator.isCompressed(r) {
				body, err := inflator.wrap(r.Body)
				if err != nil {
					http.Error(w, fmt.Sprintf("failed to decompress body: %v", err), http.StatusBadRequest)
					return
				}
				r.Body = body
			}

			next.ServeHTTP(w, r)
		})
	}
}

func newGzipInflator() func(next http.Handler) http.Handler {
	return inflatorMiddleware(&gzipInflator{})
}

type gzipInflator struct{}

func (gi *gzipInflator) isCompressed(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Content-Encoding"), "gzip")
}

func (gi *gzipInflator) wrap(rc io.ReadCloser) (io.ReadCloser, error) {
	return gzip.NewReader(rc)
}
