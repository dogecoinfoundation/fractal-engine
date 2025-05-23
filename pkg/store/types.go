package store

import (
	"database/sql"
	"time"
)

type MintWithoutID struct {
	Hash            string         `json:"hash"`
	Title           string         `json:"title"`
	FractionCount   int            `json:"fraction_count"`
	Description     string         `json:"description"`
	Tags            StringArray    `json:"tags"`
	Metadata        interface{}    `json:"metadata"`
	TransactionHash sql.NullString `json:"transaction_hash"`
	Verified        bool           `json:"verified"`
	OutputAddress   string         `json:"output_address"`
	CreatedAt       time.Time      `json:"created_at"`
}

type Mint struct {
	MintWithoutID
	Id string `json:"id"`
}
