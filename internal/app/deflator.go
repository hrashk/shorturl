package app

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// New compression algorithms have to implement this interface
// and provide a middleware constructor. Follow the example for gzip below.
type deflator interface {
	accepts(r *http.Request) bool
	wrap(rw http.ResponseWriter) *wrappedResponse
}

type wrappedResponse struct {
	http.ResponseWriter
	wc io.WriteCloser
}

func (wr *wrappedResponse) Write(b []byte) (int, error) {
	return wr.wc.Write(b)
}

func (wr *wrappedResponse) Close() {
	wr.wc.Close()
}

func compressingMiddleware(d deflator) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			effectiveWriter := w

			if d.accepts(r) {
				wrapped := d.wrap(w)
				defer wrapped.Close()

				effectiveWriter = wrapped
			}

			next.ServeHTTP(effectiveWriter, r)
		})
	}
}

func newGzipDeflator() func(next http.Handler) http.Handler {
	return compressingMiddleware(&gzipDeflator{})
}

type gzipDeflator struct{}

func (gd *gzipDeflator) wrap(w http.ResponseWriter) *wrappedResponse {
	gz, _ := gzip.NewWriterLevel(w, gzip.BestSpeed)

	wrapped := &wrappedResponse{w, gz}
	wrapped.Header().Set("Content-Encoding", "gzip")

	return wrapped
}

func (gd *gzipDeflator) accepts(r *http.Request) bool {
	values := r.Header.Values("Accept-Encoding")

	for _, v := range values {
		if strings.Contains(v, "gzip") {
			return true
		}
	}

	return false
}
