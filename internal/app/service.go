package app

import (
	"context"
	"fmt"
)

type service interface {
	CreateShortURL(ctx context.Context, url string) (shortURL string, err error)
	LookUp(ctx context.Context, key string) (url string, err error)
	PingDB(ctx context.Context) error
}

type shortURLService struct {
	keyGenerator keyGenerator
	storage      storage
	baseURL      string
}

func newService(cfg config) (s service, err error) {
	var (
		st   storage
		uuid uint64
	)
	st, uuid, err = newStorage(cfg)
	if err != nil {
		return
	}
	kg := newBase62Generator(uuid + 1)
	s = shortURLService{kg, st, cfg.baseURL}

	return
}

func (s shortURLService) CreateShortURL(ctx context.Context, url string) (shortURL string, err error) {
	key := s.keyGenerator.Generate(url)
	if err := s.storage.Store(ctx, key, url); err != nil {
		return "", fmt.Errorf("failed to store key %v: [%w]", key, err)
	}
	shortURL = s.baseURL + "/" + key.shortURL
	return shortURL, nil
}

func (s shortURLService) LookUp(ctx context.Context, key string) (url string, err error) {
	url, err = s.storage.LookUp(ctx, key)
	if err != nil {
		return "", fmt.Errorf("key %v not found: [%w]", key, err)
	}
	return url, nil
}

func (s shortURLService) PingDB(ctx context.Context) error {
	return s.storage.Ping(ctx)
}
