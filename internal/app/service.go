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

func newService(cfg *config) (s service, err error) {
	var st storage
	var uuid uint64

	st = newInMemStorage()

	if cfg.StoragePath != "" {
		uuid, err = readFile(st, cfg.StoragePath)
		if err != nil {
			return
		}
		fmt.Println("creating file storage")
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
