package app

type ShortKeyGenerator interface {
	Generate(url string) (key string)
}

type Storage interface {
	Store(key string, url string) error
	LookUp(key string) (url string, err error)
}

type ShortURLService struct {
	keyGenerator ShortKeyGenerator
	storage      Storage
}

func NewShortURLService(keyGenerator ShortKeyGenerator, storage Storage) ShortURLService {
	return ShortURLService{keyGenerator, storage}
}

func (s ShortURLService) CreateShortURL(url string) (shortURL string, err error) {
	key := s.keyGenerator.Generate(url)
	if err := s.storage.Store(key, url); err != nil {
		return "", err
	}
	shortURL = config.redirectBaseURL + "/" + key
	return shortURL, nil
}

func (s ShortURLService) LookUp(key string) (url string, err error) {
	url, err = s.storage.LookUp(key)
	if err != nil {
		return "", err
	}
	return url, nil
}
