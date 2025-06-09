package app

import "fmt"

type service interface {
	CreateShortURL(url string) (shortURL string, err error)
	LookUp(key string) (url string, err error)
}

type shortURLService struct {
	keyGenerator keyGenerator
	storage      storage
	baseURL      string
}

func newService(cfg *config) (s service, err error) {
	var st storage
	var uuid uint64

	st = newInMemStorage()

	if cfg.StoragePath != "" {
		uuid, err = readFile(st, cfg.StoragePath)
		if err != nil {
			return
		}
		st, err = newFileStorage(st, cfg.StoragePath)
		if err != nil {
			return
		}
	}

	kg := newBase62Generator(uuid + 1)
	s = &shortURLService{kg, st, cfg.baseURL}

	return
}

func (s shortURLService) CreateShortURL(url string) (shortURL string, err error) {
	key := s.keyGenerator.Generate(url)
	if err := s.storage.Store(key, url); err != nil {
		return "", fmt.Errorf("failed to store key %s: [%w]", key, err)
	}
	shortURL = s.baseURL + "/" + key
	return shortURL, nil
}

func (s shortURLService) LookUp(key string) (url string, err error) {
	url, err = s.storage.LookUp(key)
	if err != nil {
		return "", fmt.Errorf("key %v not found: [%w]", key, err)
	}
	return url, nil
}
