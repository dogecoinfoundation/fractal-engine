package store

import (
	"database/sql"
	"fmt"
	"net/url"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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
