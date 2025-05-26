package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type TokenisationStore struct {
	DB      *sql.DB
	backend string
}

func NewTokenisationStore(dbUrl string) (*TokenisationStore, error) {
	u, err := url.Parse(dbUrl)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "memory" {
		sqlite, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			return nil, err
		}

		return &TokenisationStore{DB: sqlite, backend: "sqlite"}, nil
	} else if u.Scheme == "sqlite" {
		var url string
		if u.Host == "" {
			url = u.Path
		} else {
			url = u.Host
		}

		sqlite, err := sql.Open("sqlite3", url)
		if err != nil {
			return nil, err
		}

		return &TokenisationStore{DB: sqlite, backend: "sqlite"}, nil
	} else if u.Scheme == "postgres" {
		postgres, err := sql.Open("postgres", dbUrl)
		if err != nil {
			return nil, err
		}
		return &TokenisationStore{DB: postgres, backend: "postgres"}, nil
	}

	return nil, fmt.Errorf("unsupported database scheme: %s", u.Scheme)
}

func (s *TokenisationStore) Migrate() error {
	driver, err := s.getMigrationDriver()
	if err != nil {
		return err
	}

	path, err := MigrationsPath()
	if err != nil {
		return err
	}

	path = strings.ReplaceAll(path, "\\", "/")

	m, err := migrate.NewWithDatabaseInstance("file://"+path, s.backend, driver)
	if err != nil {
		return err
	}

	return m.Up()
}

func ProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		// Check if go.mod exists in this directory
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory, cannot find go.mod
			return "", os.ErrNotExist
		}
		dir = parent
	}
}

func MigrationsPath() (string, error) {
	root, err := ProjectRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "db", "migrations"), nil
}

func (s *TokenisationStore) getMigrationDriver() (database.Driver, error) {
	if s.backend == "postgres" {
		driver, err := postgres.WithInstance(s.DB, &postgres.Config{})
		if err != nil {
			return nil, err
		}

		return driver, nil
	}

	if s.backend == "sqlite" {
		driver, err := sqlite.WithInstance(s.DB, &sqlite.Config{})
		if err != nil {
			return nil, err
		}

		return driver, nil
	}

	return nil, fmt.Errorf("unsupported database scheme: %s", s.backend)
}

func (s *TokenisationStore) ClearMints() error {
	_, err := s.DB.Exec("DELETE FROM mints")
	if err != nil {
		return err
	}
	return nil
}

func (s *TokenisationStore) GetMints(limit int, offset int, verified bool) ([]Mint, error) {
	fmt.Println("Getting mints:", limit, offset, verified)

	rows, err := s.DB.Query("SELECT id, created_at, title, description, fraction_count, tags, metadata, hash, verified, transaction_hash, requirements, resellable, lockup_options FROM mints WHERE verified = $1", verified)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mints []Mint
	for rows.Next() {
		var m Mint
		if err := rows.Scan(&m.Id, &m.CreatedAt, &m.Title, &m.Description, &m.FractionCount, &m.Tags, &m.Metadata, &m.Hash, &m.Verified, &m.TransactionHash, &m.Requirements, &m.Resellable, &m.LockupOptions); err != nil {
			return nil, err
		}
		mints = append(mints, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return mints, nil
}

func (s *TokenisationStore) GetUnsyncedMints() ([]Mint, error) {
	var mints []Mint
	err := s.DB.QueryRow("SELECT * FROM mints WHERE synced = false").Scan(&mints)
	if err != nil {
		return nil, err
	}

	return mints, nil
}

func (s *TokenisationStore) SetMintSynced(mint Mint) error {
	_, err := s.DB.Exec("UPDATE mints SET synced = true WHERE id = $1", mint.Id)
	if err != nil {
		return err
	}

	return nil
}

func (s *TokenisationStore) VerifyMint(id string, transactionHash string) error {
	fmt.Println("Verifying mint:", id, transactionHash)

	_, err := s.DB.Exec("UPDATE mints SET verified = true, transaction_hash = $1 WHERE id = $2", transactionHash, id)
	if err != nil {
		return err
	}
	return nil
}

func (s *TokenisationStore) SaveMint(mint *MintWithoutID) (string, error) {
	fmt.Println("Saving mint:", mint.Hash)

	id := uuid.New().String()

	metadata, err := json.Marshal(mint.Metadata)
	if err != nil {
		return "", err
	}

	requirements, err := json.Marshal(mint.Requirements)
	if err != nil {
		return "", err
	}

	lockupOptions, err := json.Marshal(mint.LockupOptions)
	if err != nil {
		return "", err
	}

	tags, err := json.Marshal(mint.Tags)
	if err != nil {
		return "", err
	}

	_, err = s.DB.Exec(`
	INSERT INTO mints (id, title, description, fraction_count, tags, metadata, hash, verified, requirements, resellable, lockup_options)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, id, mint.Title, mint.Description, mint.FractionCount, string(tags), string(metadata), mint.Hash, false, string(requirements), mint.Resellable, string(lockupOptions))

	return id, err
}

func (s *TokenisationStore) Close() error {
	return s.DB.Close()
}
