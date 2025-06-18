package app

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type service interface {
	CreateShortURL(url string) (shortURL string, err error)
	LookUp(key string) (url string, err error)
	PingDB(ctx context.Context) error
}

type shortURLService struct {
	keyGenerator keyGenerator
	storage      storage
	baseURL      string
	db           *sql.DB
}

func newService(cfg config) (s service, err error) {
	var (
		st   storage
		uuid uint64
		db   *sql.DB
	)
	st, uuid, err = newStorage(cfg)
	if err != nil {
		return
	}
	if cfg.dbDsn != "" {
		db, err = sql.Open("pgx", cfg.dbDsn)
		if err != nil {
			return
		}
		err = ping(db, context.Background())
		if err != nil {
			return
		}
	}
	kg := newBase62Generator(uuid + 1)
	s = shortURLService{kg, st, cfg.baseURL, db}

	return
}

func (s shortURLService) CreateShortURL(url string) (shortURL string, err error) {
	key := s.keyGenerator.Generate(url)
	if err := s.storage.Store(key, url); err != nil {
		return "", fmt.Errorf("failed to store key %v: [%w]", key, err)
	}
	shortURL = s.baseURL + "/" + key.shortURL
	return shortURL, nil
}

func (s shortURLService) LookUp(key string) (url string, err error) {
	url, err = s.storage.LookUp(key)
	if err != nil {
		return "", fmt.Errorf("key %v not found: [%w]", key, err)
	}
	return url, nil
}

func (s shortURLService) PingDB(ctx context.Context) error {
	if s.db == nil {
		return fmt.Errorf("the service is running without a db")
	}

	return ping(s.db, ctx)
}

func ping(db *sql.DB, ctx context.Context) error {
	ctx, stop := context.WithTimeout(ctx, 5*time.Second)
	defer stop()

	err := db.PingContext(ctx)
	if err != nil {
		err = fmt.Errorf("failed to ping db: %w", err)
	}
	return err
}
