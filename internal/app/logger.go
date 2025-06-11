package app

import (
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"
)

type logger interface {
	Info(msg string, v ...any)
	Error(err error, msg string, v ...any)
}

type zeroLogger struct {
	logger zerolog.Logger
}

func newZeroLogger() zeroLogger {
	zl := zerolog.New(os.Stdout).With().Timestamp().Logger()

	return zeroLogger{logger: zl}
}

func (zl zeroLogger) Info(msg string, v ...any) {
	zl.logger.Info().Msgf(msg, v...)
}

func (zl zeroLogger) Error(err error, msg string, v ...any) {
	zl.logger.Err(err).Msgf(msg, v...)
}

func loggingMiddleware(l logger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		logFn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := loggingResponseWriter{
				ResponseWriter: w,
			}
			h.ServeHTTP(&rw, r)

			duration := time.Since(start)

			l.Info("uri=%v, method=%v, status=%v, duration=%v, size=%v",
				r.RequestURI,
				r.Method,
				rw.status,
				duration,
				rw.size,
			)
		}

		return http.HandlerFunc(logFn)
	}
}

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (w *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := w.ResponseWriter.Write(b)
	w.size += size
	return size, err
}

func (w *loggingResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.status = statusCode
}
