package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type storage interface {
	Store(ctx context.Context, key shortKey, url string) error
	LookUp(ctx context.Context, shortURL string) (url string, err error)
	Ping(ctx context.Context) error
	StoreBatch(ctx context.Context, batch urlBatch) error
	LookUpKey(ctx context.Context, url string) (shortKey, error)
}

type urlBatch []struct {
	shortKey
	originalURL string
}

type inMemStorage struct {
	data *sync.Map
}

type fileStorage struct {
	storage
	log logger
	ch  chan urlRec
}

type pgsqlStorage struct {
	db *sql.DB
}

func newStorage(cfg config) (st storage, uuid uint64, err error) {
	st = newInMemStorage()

	if strings.HasPrefix(cfg.dbDsn, "postgresql") {
		st, uuid, err = newPgsqlStorage(cfg)
	} else if cfg.storagePath != "" {
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
func (s inMemStorage) Store(ctx context.Context, key shortKey, url string) error {
	s.data.Store(key.shortURL, url)

	return nil
}
func (s inMemStorage) StoreBatch(ctx context.Context, batch urlBatch) error {
	var err error
	for _, b := range batch {
		err = s.Store(ctx, b.shortKey, b.originalURL)
		if err != nil {
			break
		}
	}

	return err
}
func (s inMemStorage) LookUp(ctx context.Context, shortURL string) (url string, err error) {
	v, ok := s.data.Load(shortURL)
	if !ok {
		return "", errors.New("short URL not found: " + shortURL)
	}
	return v.(string), nil
}

func (s inMemStorage) Ping(ctx context.Context) error {
	return nil
}

func (s inMemStorage) LookUpKey(ctx context.Context, url string) (shortKey, error) {
	return shortKey{}, nil
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

func (fs fileStorage) Store(ctx context.Context, key shortKey, url string) error {
	if err := fs.storage.Store(ctx, key, url); err != nil {
		return err
	}

	fs.ch <- urlRec{key.uuid, key.shortURL, url}

	return nil
}

func (fs fileStorage) StoreBatch(ctx context.Context, batch urlBatch) error {
	if err := fs.storage.StoreBatch(ctx, batch); err != nil {
		return err
	}

	for _, b := range batch {
		fs.ch <- urlRec{b.uuid, b.shortURL, b.originalURL}
	}

	return nil
}

func (fs fileStorage) LookUpKey(ctx context.Context, url string) (shortKey, error) {
	return shortKey{}, nil
}

type urlRec struct {
	UUID        uint64 `json:"uuid,string"`
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
		err = st.Store(context.TODO(), shortKey{rec.UUID, rec.ShortURL}, rec.OriginalURL)
		if err != nil {
			return
		}

		if rec.UUID > uuid {
			uuid = rec.UUID
		}
	}
	return
}

func newPgsqlStorage(cfg config) (pst pgsqlStorage, uuid uint64, err error) {
	pst = pgsqlStorage{nil}

	pst.db, err = sql.Open("pgx", cfg.dbDsn)
	if err != nil {
		return
	}

	err = pst.Ping(context.Background())
	if err != nil {
		return
	}

	err = pst.createTables(context.Background())
	if err != nil {
		return
	}

	uuid, err = pst.fetchLastID(context.Background())

	return
}

func (pst pgsqlStorage) Ping(ctx context.Context) error {
	ctx, stop := context.WithTimeout(ctx, 5*time.Second)
	defer stop()

	err := pst.db.PingContext(ctx)
	if err != nil {
		err = fmt.Errorf("failed to ping db: %w", err)
	}
	return err
}

const createURLTable = `
CREATE TABLE IF NOT EXISTS urls (
	uuid BIGINT PRIMARY KEY,
	short_url TEXT NOT NULL,
	original_url TEXT NOT NULL UNIQUE
);`

func (pst pgsqlStorage) createTables(ctx context.Context) error {
	_, err := pst.db.ExecContext(ctx, createURLTable)
	if err != nil {
		return err
	}

	const indexQry = "CREATE INDEX IF NOT EXISTS short_url_idx ON urls (short_url)"
	_, err = pst.db.ExecContext(ctx, indexQry)

	return err
}

func (pst pgsqlStorage) fetchLastID(ctx context.Context) (uint64, error) {
	const maxQuery = "select coalesce(max(uuid),0) maxid from urls"
	var uuid uint64

	err := pst.db.QueryRowContext(ctx, maxQuery).Scan(&uuid)
	if errors.Is(err, sql.ErrNoRows) {
		uuid = 0
		err = nil
	}

	return uuid, err
}

func (pst pgsqlStorage) LookUp(ctx context.Context, shortURL string) (string, error) {
	const query = "SELECT original_url FROM urls WHERE short_url = $1"
	var originalURL string
	err := pst.db.QueryRowContext(ctx, query, shortURL).Scan(&originalURL)
	if errors.Is(err, sql.ErrNoRows) {
		err = fmt.Errorf("short URL %s not found: %w", shortURL, err)
	}
	return originalURL, err
}

var ErrConflict = errors.New("data conflict")

func (pst pgsqlStorage) Store(ctx context.Context, key shortKey, url string) error {
	const query = `
		INSERT INTO urls (uuid, short_url, original_url)
		VALUES ($1, $2, $3)
	`
	_, err := pst.db.ExecContext(ctx, query, key.uuid, key.shortURL, url)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
		err = ErrConflict
	}

	return err
}

func (pst pgsqlStorage) StoreBatch(ctx context.Context, batch urlBatch) error {
	const query = `
		INSERT INTO urls (uuid, short_url, original_url)
		VALUES ($1, $2, $3)
	`

	tx, err := pst.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, b := range batch {
		_, err = stmt.ExecContext(ctx, b.uuid, b.shortURL, b.originalURL)
		if err != nil {
			break
		}
	}

	if err == nil {
		stmt.Close()
		err = tx.Commit()
	}

	return err
}

func (pst pgsqlStorage) LookUpKey(ctx context.Context, url string) (shortKey, error) {
	const query = "SELECT uuid, short_url FROM urls WHERE original_url = $1"
	var uuid uint64
	var shortURL string
	err := pst.db.QueryRowContext(ctx, query, url).Scan(&uuid, &shortURL)
	if errors.Is(err, sql.ErrNoRows) {
		err = fmt.Errorf("original URL %s not found: %w", url, err)
	}
	return shortKey{uuid, shortURL}, err
}
