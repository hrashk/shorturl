# ShortURL

ShortURL is a URL shortening service written in Go. It provides a simple and efficient way to generate short, shareable links for long URLs.

## Features

- Generate short URLs for long links.
- Redirect users from short URLs to their original destinations.
- Simple and lightweight implementation.
- Easy to deploy and extend.

## Requirements

- Go 1.24 or later
- A database (e.g., PostgreSQL, MySQL, or SQLite)

## Installation

1. Build the project:
```bash
go build -C ./cmd/shorturl
```

2. Start the service:
```bash
./cmd/shorturl/shorturl
```

## Configuration

You may specify either environment variables
```bash
SERVER_ADDRESS=":9999" BASE_URL="http://example.com" ./cmd/shorturl/shorturl
```

and/or command-line args
```bash
./cmd/shorturl/shorturl -a ":9999" -b "http://example.com"
```

## Usage

Example of shortening a URL
```bash
curl -X POST http://localhost:8080/ -d "https://pkg.go.dev/cmp"
```

Example of redirecting a short URL
```bash
curl -v http://localhost:8080/aaaaab
```

Example of calling REST API
```bash
curl -X POST http://localhost:8080/api/shorten -d '{"url": "https://pkg.go.dev/cmp"}'
```

Example of calling REST API with gzip output
```bash
curl -X POST http://localhost:8080/api/shorten \
-H "Accept-Encoding: gzip" \
-d '{"url": "https://pkg.go.dev/cmp"}' \
-o some.gz
```

## License

This project is licensed under the BSD License.
