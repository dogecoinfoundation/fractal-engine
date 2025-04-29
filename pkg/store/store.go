package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"dogecoin.org/fractal-engine/pkg/protocol"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	DB      *sql.DB
	backend string
}

func NewStore(dbUrl string) (*Store, error) {
	u, err := url.Parse(dbUrl)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "memory" {
		sqlite, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			return nil, err
		}

		return &Store{DB: sqlite, backend: "sqlite"}, nil
	} else if u.Scheme == "sqlite" {
		sqlite, err := sql.Open("sqlite3", u.Host)
		if err != nil {
			return nil, err
		}

		return &Store{DB: sqlite, backend: "sqlite"}, nil
	} else if u.Scheme == "postgres" {
		postgres, err := sql.Open("postgres", dbUrl)
		if err != nil {
			return nil, err
		}
		return &Store{DB: postgres, backend: "postgres"}, nil
	}

	return nil, fmt.Errorf("unsupported database scheme: %s", u.Scheme)
}

func (s *Store) Migrate() error {
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

func (s *Store) getMigrationDriver() (database.Driver, error) {
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

func (s *Store) GetMint(id string) (*protocol.Mint, error) {
	var mint protocol.Mint
	err := s.DB.QueryRow("SELECT * FROM mints WHERE id = $1", id).Scan(&mint)
	if err != nil {
		return nil, err
	}

	return &mint, nil
}

func (s *Store) RemoveOnchainMint(id string) error {
	_, err := s.DB.Exec("DELETE FROM onchain_mints WHERE id = $1", id)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) GetMints(limit int, offset int, verified bool) ([]protocol.Mint, error) {
	rows, err := s.DB.Query("SELECT id, created_at, title, description, fraction_count, tags, metadata, hash, verified FROM mints WHERE verified = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3", verified, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mints []protocol.Mint
	for rows.Next() {
		var m protocol.Mint
		if err := rows.Scan(&m.Id, &m.CreatedAt, &m.Title, &m.Description, &m.FractionCount, &m.Tags, &m.Metadata, &m.Hash, &m.Verified); err != nil {
			return nil, err
		}
		mints = append(mints, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return mints, nil
}

func (s *Store) GetUnverifiedOnchainMints() ([]protocol.Mint, error) {
	var mints []protocol.Mint
	err := s.DB.QueryRow("SELECT * FROM onchain_mints WHERE verified = false").Scan(&mints)
	if err != nil {
		return nil, err
	}

	return mints, nil
}

func (s *Store) GetUnsyncedMints() ([]protocol.Mint, error) {
	var mints []protocol.Mint
	err := s.DB.QueryRow("SELECT * FROM mints WHERE synced = false").Scan(&mints)
	if err != nil {
		return nil, err
	}

	return mints, nil
}

func (s *Store) SetMintSynced(mint protocol.Mint) error {
	_, err := s.DB.Exec("UPDATE mints SET synced = true WHERE id = $1", mint.Id)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) VerifyMint(id string, transactionHash string) error {
	_, err := s.DB.Exec("UPDATE mints SET verified = true, transaction_hash = $1 WHERE id = $2", transactionHash, id)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) SaveMint(mint *protocol.MintWithoutID) (string, error) {
	id := uuid.New().String()

	metadata, err := json.Marshal(mint.Metadata)
	if err != nil {
		return "", err
	}

	tags, err := json.Marshal(mint.Tags)
	if err != nil {
		return "", err
	}

	_, err = s.DB.Exec(`
	INSERT INTO mints (id, title, description, fraction_count, tags, metadata, hash)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, id, mint.Title, mint.Description, mint.FractionCount, string(tags), string(metadata), mint.Hash)

	return id, err
}

func (s *Store) CreateOnchainMint(mint protocol.Mint, transactionHash string) error {
	_, err := s.DB.Exec(`
	INSERT INTO onchain_mints (id, hash, transaction_hash)
	VALUES ($1, $2, $3)
	`, mint.Id, mint.Hash, transactionHash)

	return err
}

func (s *Store) GetChainPosition() (int64, string, error) {
	var blockHeight int64
	var blockHash string

	err := s.DB.QueryRow("SELECT block_height, block_hash FROM chain_position").Scan(&blockHeight, &blockHash)
	if err == sql.ErrNoRows {
		return 0, "", nil
	}

	if err != nil {
		return 0, "", err
	}

	return blockHeight, blockHash, nil
}

func (s *Store) UpsertChainPosition(blockHeight int64, blockHash string) error {
	_, err := s.DB.Exec(`
	INSERT INTO chain_position (id, block_height, block_hash)
	VALUES (1, $1, $2)
	ON CONFLICT (id)
	DO UPDATE SET block_height = EXCLUDED.block_height,
				  block_hash = EXCLUDED.block_hash
`, blockHeight, blockHash)

	return err
}
