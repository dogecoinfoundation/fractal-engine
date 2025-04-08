package protocol

import "encoding/json"

type Mint struct {
	Title         string `json:"title"`
	FractionCount int    `json:"fraction_count"`
	Description   string `json:"description"`
}

func (m *Mint) Deserialize(data []byte) error {
	return json.Unmarshal(data, m)
}

func (m *Mint) Serialize() ([]byte, error) {
	return json.Marshal(m)
}
