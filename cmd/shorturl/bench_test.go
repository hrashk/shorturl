package main

import (
	"os"
	"strconv"
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

func BenchmarkShortenBatchPostgres(b *testing.B) {
	os.Args = []string{os.Args[0], "-d", app.DefaultDatabaseDsn}
	cli := setUpEmptyStorage(b)

	payload := make(app.BatchRequest, 100)
	const prefix = "123e4567-e89b-12d3-a456-4266141740"
	for i := range payload {
		istr := strconv.Itoa(i)
		payload[i].CorrelationID = prefix + istr
		payload[i].OriginalURL = sampleURL + "/path" + istr
	}

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			cli.Batch(payload)
		}
	})
}
