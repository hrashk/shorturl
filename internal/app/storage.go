package app

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
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

type fileStorage struct {
	storage
	file *os.File
}

func newFileStorage(st storage, path string) (fileStorage, error) {
	fs := fileStorage{
		storage: st,
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return fs, fmt.Errorf("failed to open file %s: %w", path, err)
	}
	fs.file = f

	return fs, nil
}

func (fs fileStorage) Store(key string, url string) error {
	if err := fs.storage.Store(key, url); err != nil {
		return err
	}

	rec := urlRec{strconv.Itoa(0), key, url} // todo fix

	return json.NewEncoder(fs.file).Encode(&rec)
}

type urlRec struct {
	UUID, ShortURL, OriginalURL string
}

func readFile(st storage, path string) (uuid uint64, err error) {
	file, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		err = fmt.Errorf("failed to open file %s: %w", path, err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		bytes := scanner.Bytes()

		rec := urlRec{}
		if err = json.Unmarshal(bytes, &rec); err != nil {
			err = fmt.Errorf("failed to parse %s: %w", string(bytes), err)
			return
		}

		if err = st.Store(rec.ShortURL, rec.OriginalURL); err != nil {
			return
		}

		id, e := strconv.ParseUint(rec.UUID, 10, 64)
		if e != nil {
			err = fmt.Errorf("failed to convert %s to int: %w", rec.UUID, e)
			return
		}
		if id > uuid {
			uuid = id
		}
	}
	err = scanner.Err()
	return
}
