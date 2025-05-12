package app

import "errors"

type InMemStorage struct {
	data map[string]string
}

func NewInMemStorage() *InMemStorage {
	return &InMemStorage{
		data: make(map[string]string),
	}
}
func (s *InMemStorage) Store(key string, url string) error {
	s.data[key] = url

	return nil
}
func (s *InMemStorage) LookUp(key string) (url string, err error) {
	url, ok := s.data[key]
	if !ok {
		return "", errors.New("key not found: " + key)
	}
	return url, nil
}
