package store

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
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
	CreatedAt       time.Time      `json:"created_at"`
	Requirements    interface{}    `json:"requirements"`
	Resellable      bool           `json:"resellable"`
	LockupOptions   interface{}    `json:"lockup_options"`
}

type MintHash struct {
	Title         string      `json:"title"`
	FractionCount int         `json:"fraction_count"`
	Description   string      `json:"description"`
	Tags          StringArray `json:"tags"`
	Metadata      interface{} `json:"metadata"`
	Requirements  interface{} `json:"requirements"`
	Resellable    bool        `json:"resellable"`
	LockupOptions interface{} `json:"lockup_options"`
}

func (m *MintWithoutID) GenerateHash() (string, error) {
	input := MintHash{
		Title:         m.Title,
		FractionCount: m.FractionCount,
		Description:   m.Description,
		Tags:          m.Tags,
		Metadata:      m.Metadata,
		Requirements:  m.Requirements,
		Resellable:    m.Resellable,
		LockupOptions: m.LockupOptions,
	}

	// Serialize to JSON with sorted keys
	jsonBytes, err := json.Marshal(input)
	if err != nil {
		return "", err
	}

	// Generate SHA-256 hash (32 bytes)
	hash := sha256.Sum256(jsonBytes)

	// Return as byte slice (length 32)
	return hex.EncodeToString(hash[:]), nil
}

type Mint struct {
	MintWithoutID
	Id string `json:"id"`
}
