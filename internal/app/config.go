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
	serverAddress string
	baseURL       string
}{
	serverAddress: DefaultServerAddress,
	baseURL:       DefaultBaseURL,
}

func GetListenAddr() string {
	return config.serverAddress
}

func SetListenAddr(addr string) {
	_, port, err := net.SplitHostPort(addr)
	if err != nil || port == "" {
		panic(fmt.Sprintf("Invalid server address %s", addr))
	}
	config.serverAddress = addr
}

func GetRedirectBaseURL() string {
	return config.baseURL
}

func SetRedirectBaseURL(baseURL string) {
	u, err := url.Parse(baseURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		panic(fmt.Sprintf("Invalid base URL %s.", baseURL))
	}
	config.baseURL = baseURL
}
