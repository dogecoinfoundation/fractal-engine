package protocol

import (
	"crypto/sha256"
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"
)

type StringArray []string

// Scan implements the sql.Scanner interface
func (s *StringArray) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		str, ok := value.(string)
		if !ok {
			return errors.New("type assertion to string failed")
		}
		bytes = []byte(str)
	}
	return json.Unmarshal(bytes, s)
}

// Value implements the driver.Valuer interface (optional if you want to write back to DB)
func (s StringArray) Value() (driver.Value, error) {
	return json.Marshal(s)
}

type MintWithoutID struct {
	Hash            string      `json:"hash"`
	Title           string      `json:"title"`
	FractionCount   int         `json:"fraction_count"`
	Description     string      `json:"description"`
	Tags            StringArray `json:"tags"`
	Metadata        interface{} `json:"metadata"`
	TransactionHash string      `json:"transaction_hash"`
	Verified        bool        `json:"verified"`
	CreatedAt       time.Time   `json:"created_at"`
}

type MintHash struct {
	Title         string      `json:"title"`
	FractionCount int         `json:"fraction_count"`
	Description   string      `json:"description"`
	Tags          StringArray `json:"tags"`
	Metadata      interface{} `json:"metadata"`
}

type Mint struct {
	MintWithoutID
	Id string `json:"id"`
}

func NewMint(mintWithoutID MintWithoutID) (MintWithoutID, error) {
	hash, err := mintWithoutID.GenerateHash()
	if err != nil {
		return MintWithoutID{}, err
	}

	mintWithoutID.Hash = hash

	return mintWithoutID, nil
}

func (m *Mint) Deserialize(data []byte) error {
	return json.Unmarshal(data, m)
}

func (m *Mint) Serialize() ([]byte, error) {
	return json.Marshal(m)
}

func (m *Mint) SerializeToMintTransaction() []byte {
	return []byte(m.Id)
}

func (m *MintWithoutID) GenerateHash() (string, error) {
	input := MintHash{
		Title:         m.Title,
		FractionCount: m.FractionCount,
		Description:   m.Description,
		Tags:          m.Tags,
		Metadata:      m.Metadata,
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
