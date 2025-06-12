package app

import (
	"net/http"
)

var NewServer = func(modifiers ...Configurator) (*http.Server, error) {
	cfg, err := newConfig(modifiers...)
	if err != nil {
		return nil, err
	}
	a, err := newAdapter(cfg)
	if err != nil {
		return nil, err
	}

	return &http.Server{Addr: cfg.serverAddress, Handler: a.handler()}, nil
}
