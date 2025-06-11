package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"
)

type storage interface {
	Store(key shortKey, url string) error
	LookUp(shortURL string) (url string, err error)
}

type inMemStorage struct {
	data *sync.Map
}

type fileStorage struct {
	storage
	log logger
	ch  chan urlRec
}

func newStorage(cfg config) (st storage, uuid uint64, err error) {
	st = newInMemStorage()

	if cfg.storagePath != "" {
		uuid, err = readFile(st, cfg.storagePath)
		if err != nil {
			return
		}
		st, err = newFileStorage(st, cfg)
		if err != nil {
			return
		}
	}
	return
}

func newInMemStorage() inMemStorage {
	return inMemStorage{
		data: &sync.Map{},
	}
}
func (s inMemStorage) Store(key shortKey, url string) error {
	s.data.Store(key.shortURL, url)

	return nil
}
func (s inMemStorage) LookUp(shortURL string) (url string, err error) {
	v, ok := s.data.Load(shortURL)
	if !ok {
		return "", errors.New("short URL not found: " + shortURL)
	}
	return v.(string), nil
}

func newFileStorage(st storage, cfg config) (fileStorage, error) {
	fs := fileStorage{
		storage: st,
		log:     cfg.log,
	}

	f, err := os.OpenFile(cfg.storagePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return fs, fmt.Errorf("failed to open file %s: %w", cfg.storagePath, err)
	}
	fs.ch = make(chan urlRec, 100)
	go fs.storeRec(f)

	return fs, nil
}

func (fs fileStorage) storeRec(file *os.File) {
	encoder := json.NewEncoder(file)

	for {
		rec := <-fs.ch

		if err := encoder.Encode(&rec); err != nil {
			fs.log.Error(err, "writing record %v to file %s", rec, file.Name())
		}
		if err := file.Sync(); err != nil {
			fs.log.Error(err, "syncing file %s to disc", file.Name())
		}
	}
}

func (fs fileStorage) Store(key shortKey, url string) error {
	if err := fs.storage.Store(key, url); err != nil {
		return err
	}

	fs.ch <- urlRec{strconv.FormatUint(key.uuid, 10), key.shortURL, url}

	return nil
}

type urlRec struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func readFile(st storage, path string) (uuid uint64, err error) {
	file, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		err = fmt.Errorf("failed to open file %s: %w", path, err)
		return
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	for {
		rec := urlRec{}
		err = decoder.Decode(&rec)

		if err == io.EOF {
			err = nil
			break
		} else if err != nil {
			err = fmt.Errorf("failed to decode record at offset %d: %w", decoder.InputOffset(), err)
			return
		}

		var id uint64
		id, err = strconv.ParseUint(rec.UUID, 10, 64)
		if err != nil {
			err = fmt.Errorf("failed to convert %s to int: %w", rec.UUID, err)
			return
		}
		if err = st.Store(shortKey{id, rec.ShortURL}, rec.OriginalURL); err != nil {
			return
		}

		if id > uuid {
			uuid = id
		}
	}
	return
}
