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

type config struct {
	serverAddress string
	baseURL       string
	log           logger
	storage       storage
}

func newConfig(modifiers ...CfgModifier) (*config, error) {
	cfg := &config{
		serverAddress: DefaultServerAddress,
		baseURL:       DefaultBaseURL,
	}

	for _, m := range modifiers {
		if err := m(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.log == nil {
		cfg.log = newZeroLogger()
	}
	if cfg.storage == nil {
		cfg.storage = newInMemStorage()
	}

	return cfg, nil
}

type CfgModifier func(*config) error

func ServerAddress(addr string) CfgModifier {
	return func(cfg *config) error {
		_, port, err := net.SplitHostPort(addr)
		if err != nil || port == "" {
			return fmt.Errorf("invalid server address %s", addr)
		}
		cfg.serverAddress = addr
		return nil
	}
}

func BaseURL(baseURL string) CfgModifier {
	return func(cfg *config) error {
		u, err := url.Parse(baseURL)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("invalid base URL %s", baseURL)
		}
		cfg.baseURL = baseURL
		return nil
	}
}

func Logger(log logger) CfgModifier {
	return func(cfg *config) error {
		cfg.log = log
		return nil
	}
}

func Storage(s storage) CfgModifier {
	return func(cfg *config) error {
		cfg.storage = s
		return nil
	}
}
