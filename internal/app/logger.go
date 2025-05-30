package app

import (
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"
)

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
