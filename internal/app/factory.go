package app

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type Logger interface {
	Info(msg string, fields ...any)
}

func NewInMemoryController() ShortURLController {
	s := NewShortURLService(NewBase62Generator(), NewInMemStorage())

	return NewShortURLController(s)
}

func NewHandler() http.Handler {
	return NewHandlerWithLogger(NewZeroLogger())
}

func NewHandlerWithLogger(logger Logger) http.Handler {
	c := NewInMemoryController()

	r := chi.NewRouter()
	r.Use(wrapper{logger}.middleware)

	r.Get("/{key}", c.RedirectToOriginalURL)
	r.Post("/", c.CreateShortURL)
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Operation is not supported", http.StatusBadRequest)
	})

	return r
}

type wrapper struct {
	Logger
}

func (lw wrapper) middleware(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := responseData{
			status: 0,
			size:   0,
		}
		rw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		h.ServeHTTP(&rw, r)

		duration := time.Since(start)

		lw.Info("",
			"uri", r.RequestURI,
			"method", r.Method,
			"status", responseData.status,
			"duration", duration,
			"size", responseData.size,
		)
	}

	return http.HandlerFunc(logFn)
}

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}
