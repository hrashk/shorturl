package app

import (
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"
)

type logger interface {
	Info(msg string, fields ...any)
}

type ZeroLogger struct {
	logger zerolog.Logger
}

func NewZeroLogger() ZeroLogger {
	zl := zerolog.New(os.Stdout).With().Timestamp().Logger()

	return ZeroLogger{logger: zl}
}

func (zl ZeroLogger) Info(msg string, fields ...any) {
	zl.logger.Info().Fields(fields).Msg(msg)
}

func loggingMiddleware(logger logger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		logFn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := loggingResponseWriter{
				ResponseWriter: w,
			}
			h.ServeHTTP(&rw, r)

			duration := time.Since(start)

			logger.Info("",
				"uri", r.RequestURI,
				"method", r.Method,
				"status", rw.status,
				"duration", duration,
				"size", rw.size,
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
