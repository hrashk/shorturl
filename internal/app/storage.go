package app

import (
	"errors"
	"sync"
)

type InMemStorage struct {
	data *sync.Map
}

func NewInMemStorage() InMemStorage {
	return InMemStorage{
		data: &sync.Map{},
	}
}
func (s InMemStorage) Store(key string, url string) error {
	s.data.Store(key, url)

	return nil
}
func (s InMemStorage) LookUp(key string) (url string, err error) {
	v, ok := s.data.Load(key)
	if !ok {
		return "", errors.New("key not found: " + key)
	}
	return v.(string), nil
}
