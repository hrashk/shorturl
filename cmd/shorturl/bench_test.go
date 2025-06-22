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
	zerolog.DefaultContextLogger = &logger
}

func setUpEmptyStorage(b *testing.B) app.Client {
	srv := newServer(b)
	cli := app.NewClient(b)

	srv.wipeData()

	srv.start()
	cli.BaseURL = srv.baseURL
	b.Cleanup(func() {
		srv.stop()
	})
	return cli
}

func BenchmarkShortenInMemory(b *testing.B) {
	os.Args = []string{os.Args[0], "-f", ""}
	cli := setUpEmptyStorage(b)

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			resp := cli.POST("", "text/plain", sampleURL)
			resp.Body.Close()
		}
	})
}

func BenchmarkShortenFileStorate(b *testing.B) {
	os.Args = []string{os.Args[0]}
	cli := setUpEmptyStorage(b)

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			resp := cli.POST("", "text/plain", sampleURL)
			resp.Body.Close()
		}
	})
}

func BenchmarkShortenPostgres(b *testing.B) {
	os.Args = []string{os.Args[0], "-d", app.DefaultDatabaseDsn}
	cli := setUpEmptyStorage(b)

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			resp := cli.POST("", "text/plain", sampleURL)
			resp.Body.Close()
		}
	})
}
