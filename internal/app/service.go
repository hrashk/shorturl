package app

type ShortKeyGenerator interface {
	Generate(url string) (key string)
}

type Storage interface {
	Store(key string, url string) error
	LookUp(key string) (url string, err error)
}

type ShortURLService struct {
	KeyGenerator ShortKeyGenerator
	Storage      Storage
}

func NewShortURLService(keyGenerator ShortKeyGenerator, storage Storage) ShortURLService {
	return ShortURLService{KeyGenerator: keyGenerator, Storage: storage}
}

func (s ShortURLService) CreateShortURL(url string) (shortURL string, err error) {
	key := s.KeyGenerator.Generate(url)
	if err := s.Storage.Store(key, url); err != nil {
		return "", err
	}
	shortURL = "http://localhost:8080/" + key
	return shortURL, nil
}

func (s ShortURLService) LookUp(key string) (url string, err error) {
	url, err = s.Storage.LookUp(key)
	if err != nil {
		return "", err
	}
	return url, nil
}
