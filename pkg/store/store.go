package store

import (
	"database/sql"
	"log"

	"dogecoin.org/fractal-engine/pkg/config"
	_ "github.com/lib/pq"
)

type TokenisationStore struct {
	db *sql.DB
}

func (t *TokenisationStore) GetMints(start int, end int, verified bool) ([]Mint, error) {
	query := `
		SELECT * FROM mints
		WHERE verified = $1
		LIMIT $2 OFFSET $3
	`

	rows, err := t.db.Query(query, verified, end, start)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	mints := []Mint{}

	for rows.Next() {
		var mint Mint
		err := rows.Scan(&mint.Id, &mint.Hash, &mint.Title, &mint.FractionCount, &mint.Description, &mint.Tags, &mint.Metadata, &mint.TransactionHash, &mint.Verified, &mint.OutputAddress, &mint.CreatedAt)
		if err != nil {
			return nil, err
		}
		mints = append(mints, mint)
	}

	return mints, nil
}

func NewTokenisationStore(cfg *config.Config) *TokenisationStore {
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	return &TokenisationStore{db: db}
}

func (t *TokenisationStore) Close() error {
	return t.db.Close()
}
