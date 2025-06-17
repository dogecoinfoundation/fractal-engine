package store

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"time"
)

type MintWithoutID struct {
	Hash            string                 `json:"hash"`
	Title           string                 `json:"title"`
	FractionCount   int                    `json:"fraction_count"`
	Description     string                 `json:"description"`
	Tags            StringArray            `json:"tags"`
	Metadata        map[string]interface{} `json:"metadata"`
	TransactionHash sql.NullString         `json:"transaction_hash"`
	BlockHeight     int64                  `json:"block_height"`
	Verified        bool                   `json:"verified"`
	CreatedAt       time.Time              `json:"created_at"`
	Requirements    map[string]interface{} `json:"requirements"`
	LockupOptions   map[string]interface{} `json:"lockup_options"`
	Gossiped        bool                   `json:"gossiped"`
	FeedURL         string                 `json:"feed_url"`
}

type MintHash struct {
	Title         string                 `json:"title"`
	FractionCount int                    `json:"fraction_count"`
	Description   string                 `json:"description"`
	Tags          StringArray            `json:"tags"`
	Metadata      map[string]interface{} `json:"metadata"`
	Requirements  map[string]interface{} `json:"requirements"`
	LockupOptions map[string]interface{} `json:"lockup_options"`
}

type OnChainTransaction struct {
	Id            string `json:"id"`
	TxHash        string `json:"tx_hash"`
	Height        int64  `json:"height"`
	ActionType    uint8  `json:"action_type"`
	ActionVersion uint8  `json:"action_version"`
	ActionData    []byte `json:"action_data"`
}

func (m *MintWithoutID) GenerateHash() (string, error) {
	input := MintHash{
		Title:         m.Title,
		FractionCount: m.FractionCount,
		Description:   m.Description,
		Tags:          m.Tags,
		Metadata:      m.Metadata,
		Requirements:  m.Requirements,
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

type OnChainMint struct {
	MintId          string `json:"mint_id"`
	TransactionHash string `json:"transaction_hash"`
	Address         string `json:"address"`
}
