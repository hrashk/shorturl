package app

import "fmt"

type shortKeyGenerator interface {
	Generate(url string) (key string)
}

type storage interface {
	Store(key string, url string) error
	LookUp(key string) (url string, err error)
}

type shortURLService struct {
	keyGenerator shortKeyGenerator
	storage      storage
	baseURL      string
}

func newShortURLService(kg shortKeyGenerator, st storage, baseURL string) shortURLService {
	return shortURLService{kg, st, baseURL}
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
