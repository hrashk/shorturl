package app

import (
	"context"
	"errors"
	"fmt"
)

type service interface {
	CreateShortURL(ctx context.Context, url string) (shortURL string, err error)
	LookUp(ctx context.Context, key string) (url string, err error)
	PingDB(ctx context.Context) error
	ShortenBatch(ctx context.Context, req BatchRequest) (BatchResponse, error)
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

	err = s.storage.Store(ctx, key, url)
	if errors.Is(err, ErrConflict) {
		key, err = s.storage.LookUpKey(ctx, url)
		if err == nil {
			err = ErrConflict
		}
	} else if err != nil {
		err = fmt.Errorf("failed to store key %v: [%w]", key, err)
	}
	shortURL = s.baseURL + "/" + key.shortURL
	return shortURL, err
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

type BatchRequest []struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchResponse []struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func (s shortURLService) ShortenBatch(ctx context.Context, req BatchRequest) (BatchResponse, error) {
	resp := make(BatchResponse, len(req))
	data := make(urlBatch, len(req))

	for i, r := range req {
		key := s.keyGenerator.Generate(r.OriginalURL)
		data[i].shortKey = key
		data[i].originalURL = r.OriginalURL

		resp[i].CorrelationID = r.CorrelationID
		resp[i].ShortURL = s.baseURL + "/" + key.shortURL
	}

	if err := s.storage.StoreBatch(ctx, data); err != nil {
		return nil, fmt.Errorf("failed to store batch of urls: [%w]", err)
	}

	return resp, nil
}
