package main

import (
	"os"
	"testing"

	"io"

	"github.com/hrashk/shorturl/internal/app"
	"github.com/rs/zerolog"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	logger := zerolog.New(io.Discard)
	zerolog.DefaultContextLogger = &logger // If you use context logger
}

func BenchmarkShortenInMemory(b *testing.B) {
	srv := newServer(b)
	cli := app.NewClient(b)

	srv.wipeData()

	os.Args = []string{os.Args[0], "-f", ""}
	srv.start()
	cli.BaseURL = srv.baseURL
	b.Cleanup(func() {
		srv.stop()
	})

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			resp := cli.POST("", "text/plain", sampleURL)
			resp.Body.Close()
		}
	})
}

func BenchmarkShortenFileStorate(b *testing.B) {
	srv := newServer(b)
	cli := app.NewClient(b)

	srv.wipeData()

	os.Args = []string{os.Args[0]}
	srv.start()
	cli.BaseURL = srv.baseURL
	b.Cleanup(func() {
		srv.stop()
	})

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			resp := cli.POST("", "text/plain", sampleURL)
			resp.Body.Close()
		}
	})
}

func BenchmarkShortenPostgres(b *testing.B) {
	srv := newServer(b)
	cli := app.NewClient(b)

	srv.wipeData()

	os.Args = []string{os.Args[0], "-d", app.DefaultDatabaseDsn}
	srv.start()
	cli.BaseURL = srv.baseURL
	b.Cleanup(func() {
		srv.stop()
	})

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			resp := cli.POST("", "text/plain", sampleURL)
			resp.Body.Close()
		}
	})
}
