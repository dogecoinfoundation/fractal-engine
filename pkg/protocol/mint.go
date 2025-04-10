package protocol

import (
	"encoding/json"
	"time"
)

type MintWithoutID struct {
	Title         string      `json:"title"`
	FractionCount int         `json:"fraction_count"`
	Description   string      `json:"description"`
	Tags          []string    `json:"tags"`
	Metadata      interface{} `json:"metadata"`
	Verified      bool        `json:"verified"`
	CreatedAt     time.Time   `json:"created_at"`
}

type Mint struct {
	MintWithoutID
	Id string `json:"id"`
}

func (m *Mint) Deserialize(data []byte) error {
	return json.Unmarshal(data, m)
}

func (m *Mint) Serialize() ([]byte, error) {
	return json.Marshal(m)
}

func (m *MintWithoutID) Deserialize(data []byte) error {
	return json.Unmarshal(data, m)
}

func (m *MintWithoutID) Serialize() ([]byte, error) {
	return json.Marshal(m)
}
