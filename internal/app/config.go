package app

import (
	"net"
	"net/url"
)

var config = struct {
	listenAddr      string
	redirectBaseURL string
}{
	listenAddr:      ":8080",
	redirectBaseURL: "http://localhost:8080",
}

func GetListenAddr() string {
	return config.listenAddr
}

func SetListenAddr(addr string) {
	_, port, err := net.SplitHostPort(addr)
	if err != nil || port == "" {
		panic("listen address must be in the form host:port")
	}
	config.listenAddr = addr
}

func GetRedirectBaseURL() string {
	return config.redirectBaseURL
}

func SetRedirectBaseURL(baseURL string) {
	u, err := url.Parse(baseURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		panic("redirect base URL must be a valid absolute URL (e.g., http://example.com)")
	}
	config.redirectBaseURL = baseURL
}
