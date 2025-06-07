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
	uuid int
	file *os.File
}

func newFileStorage(path string) (*fileStorage, error) {
	fs := &fileStorage{
		storage: newInMemStorage(),
	}

	if err := fs.readFile(path); err != nil {
		return nil, err
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", path, err)
	}
	fs.file = f

	return fs, nil
}

type urlRec struct {
	Uuid, Short_url, Original_url string
}

func (fs *fileStorage) readFile(path string) error {
	file, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		bytes := scanner.Bytes()

		rec := urlRec{}
		if err = json.Unmarshal(bytes, &rec); err != nil {
			return fmt.Errorf("failed to parse %s: %w", string(bytes), err)
		}

		if err = fs.storage.Store(rec.Short_url, rec.Original_url); err != nil {
			return err
		}

		uuid, err := strconv.Atoi(rec.Uuid)
		if err != nil {
			return fmt.Errorf("failed to convert %s to int: %w", rec.Uuid, err)
		}
		fs.uuid = uuid
	}

	return scanner.Err()
}

func (fs *fileStorage) Store(key string, url string) error {
	if err := fs.storage.Store(key, url); err != nil {
		return err
	}

	fs.uuid++
	rec := urlRec{strconv.Itoa(fs.uuid), key, url}

	return json.NewEncoder(fs.file).Encode(&rec)
}
