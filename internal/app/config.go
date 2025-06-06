package app

import (
	"fmt"
	"net"
	"net/url"
)

const (
	DefaultServerAddress = ":8080"
	DefaultBaseURL       = "http://localhost:8080"
)

var config = struct {
	listenAddr      string
	redirectBaseURL string
}{
	listenAddr:      DefaultServerAddress,
	redirectBaseURL: DefaultBaseURL,
}

func GetListenAddr() string {
	return config.listenAddr
}

func SetListenAddr(addr string) {
	_, port, err := net.SplitHostPort(addr)
	if err != nil || port == "" {
		panic(fmt.Sprintf("Invalid server address %s", addr))
	}
	config.listenAddr = addr
}

func GetRedirectBaseURL() string {
	return config.redirectBaseURL
}

func SetRedirectBaseURL(baseURL string) {
	u, err := url.Parse(baseURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		panic(fmt.Sprintf("Invalid base URL %s.", baseURL))
	}
	config.redirectBaseURL = baseURL
}
