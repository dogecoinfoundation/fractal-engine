package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"

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
	db      *sql.DB
	backend string
}

func NewStore(dbUrl string) (*Store, error) {
	u, err := url.Parse(dbUrl)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "sqlite" {
		sqlite, err := sql.Open("sqlite3", u.Host)
		if err != nil {
			return nil, err
		}

		return &Store{db: sqlite, backend: "sqlite"}, nil
	} else if u.Scheme == "postgres" {
		postgres, err := sql.Open("postgres", dbUrl)
		if err != nil {
			return nil, err
		}
		return &Store{db: postgres, backend: "postgres"}, nil
	}

	return nil, fmt.Errorf("unsupported database scheme: %s", u.Scheme)
}

func (s *Store) Migrate() error {
	driver, err := s.getMigrationDriver()
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance("file://db/migrations", s.backend, driver)
	if err != nil {
		return err
	}

	return m.Up()
}

func (s *Store) getMigrationDriver() (database.Driver, error) {
	if s.backend == "postgres" {
		driver, err := postgres.WithInstance(s.db, &postgres.Config{})
		if err != nil {
			return nil, err
		}

		return driver, nil
	}

	if s.backend == "sqlite" {
		driver, err := sqlite.WithInstance(s.db, &sqlite.Config{})
		if err != nil {
			return nil, err
		}

		return driver, nil
	}

	return nil, fmt.Errorf("unsupported database scheme: %s", s.backend)
}

func (s *Store) GetMint(id string) (*protocol.Mint, error) {
	var mint protocol.Mint
	err := s.db.QueryRow("SELECT * FROM mints WHERE id = $1", id).Scan(&mint)
	if err != nil {
		return nil, err
	}

	return &mint, nil
}

func (s *Store) GetUnsyncedMints() ([]protocol.Mint, error) {
	var mints []protocol.Mint
	err := s.db.QueryRow("SELECT * FROM mints WHERE synced = false").Scan(&mints)
	if err != nil {
		return nil, err
	}

	return mints, nil
}

func (s *Store) SetMintSynced(mint protocol.Mint) error {
	_, err := s.db.Exec("UPDATE mints SET synced = true WHERE id = $1", mint.Id)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) VerifyMint(id string) error {
	_, err := s.db.Exec("UPDATE mints SET verified = true WHERE id = $1", id)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) SaveMint(mint *protocol.Mint) (string, error) {
	id := uuid.New().String()

	metadata, err := json.Marshal(mint.Metadata)
	if err != nil {
		return "", err
	}

	tags, err := json.Marshal(mint.Tags)
	if err != nil {
		return "", err
	}

	_, err = s.db.Exec(`
	INSERT INTO mints (id, title, description, fraction_count, tags, metadata)
	VALUES ($1, $2, $3, $4, $5, $6)
	`, id, mint.Title, mint.Description, mint.FractionCount, string(tags), string(metadata))

	return id, err
}

func (s *Store) GetChainPosition() (int64, string, error) {
	var blockHeight int64
	var blockHash string

	err := s.db.QueryRow("SELECT block_height, block_hash FROM chain_position").Scan(&blockHeight, &blockHash)
	if err == sql.ErrNoRows {
		return 0, "", nil
	}

	if err != nil {
		return 0, "", err
	}

	return blockHeight, blockHash, nil
}

func (s *Store) UpsertChainPosition(blockHeight int64, blockHash string) error {
	_, err := s.db.Exec(`
	INSERT INTO chain_position (id, block_height, block_hash)
	VALUES (1, $1, $2)
	ON CONFLICT (id)
	DO UPDATE SET block_height = EXCLUDED.block_height,
				  block_hash = EXCLUDED.block_hash
`, blockHeight, blockHash)

	return err
}
