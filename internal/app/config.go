package app

import (
	"fmt"
	"net"
	"net/url"
)

const (
	DefaultServerAddress = ":8080"
	DefaultBaseURL       = "http://localhost:8080"
	DefaultStoragePath   = "/tmp/short-url-db.json"
)

type config struct {
	serverAddress string
	baseURL       string
	log           logger
	storagePath   string
}

func newConfig(modifiers ...Configurator) (config, error) {
	cfg := config{
		serverAddress: DefaultServerAddress,
		baseURL:       DefaultBaseURL,
		storagePath:   DefaultStoragePath,
	}

	for _, m := range modifiers {
		if err := m(&cfg); err != nil {
			return cfg, err
		}
	}

	if cfg.log == nil {
		cfg.log = newZeroLogger()
	}

	return cfg, nil
}

type Configurator func(*config) error

func WithServerAddress(addr string) Configurator {
	return func(cfg *config) error {
		_, port, err := net.SplitHostPort(addr)
		if err != nil || port == "" {
			return fmt.Errorf("invalid server address %s", addr)
		}
		cfg.serverAddress = addr
		return nil
	}
}

func WithBaseURL(baseURL string) Configurator {
	return func(cfg *config) error {
		u, err := url.Parse(baseURL)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("invalid base URL %s", baseURL)
		}
		cfg.baseURL = baseURL
		return nil
	}
}

func WithLogger(log logger) Configurator {
	return func(cfg *config) error {
		cfg.log = log
		return nil
	}
}

func WithStoragePath(path string) Configurator {
	return func(cfg *config) error {
		cfg.storagePath = path
		return nil
	}
}

func WithMemoryStorage() Configurator {
	return func(cfg *config) error {
		cfg.storagePath = ""
		return nil
	}
}
