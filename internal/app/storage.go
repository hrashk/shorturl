package app

import (
	"errors"
	"sync"
)

type inMemStorage struct {
	data *sync.Map
}

func newInMemStorage() inMemStorage {
	return inMemStorage{
		data: &sync.Map{},
	}
}
func (s inMemStorage) Store(key string, url string) error {
	s.data.Store(key, url)

	return nil
}
func (s inMemStorage) LookUp(key string) (url string, err error) {
	v, ok := s.data.Load(key)
	if !ok {
		return "", errors.New("key not found: " + key)
	}
	return v.(string), nil
}
