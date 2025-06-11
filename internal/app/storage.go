package app

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
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
	ticker := time.NewTicker(100 * time.Millisecond)
	encoder := json.NewEncoder(file)

	for {
		select {
		case rec := <-fs.ch:
			err := encoder.Encode(&rec)
			if err != nil {
				fs.log.Error(err, "writing record %v to file %s", rec, file.Name())
			}
		case <-ticker.C:
			if err := file.Sync(); err != nil {
				fs.log.Error(err, "syncing file %s to disc", file.Name())
			}
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

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		bytes := scanner.Bytes()

		rec := urlRec{}
		if err = json.Unmarshal(bytes, &rec); err != nil {
			err = fmt.Errorf("failed to parse %s: %w", string(bytes), err)
			return
		}

		id, e := strconv.ParseUint(rec.UUID, 10, 64)
		if e != nil {
			err = fmt.Errorf("failed to convert %s to int: %w", rec.UUID, e)
			return
		}
		if err = st.Store(shortKey{id, rec.ShortURL}, rec.OriginalURL); err != nil {
			return
		}

		if id > uuid {
			uuid = id
		}
	}
	err = scanner.Err()
	return
}
