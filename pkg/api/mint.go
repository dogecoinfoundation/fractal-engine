package api

import (
	"encoding/json"

	"dogecoin.org/fractal-engine/pkg/protocol"
)

type CreateMintRequest struct {
	protocol.MintWithoutID
}

type CreateMintResponse struct {
	Id   string `json:"id"`
	Hash string `json:"hash"`
}

type GetMintsResponse struct {
	Mints []protocol.Mint `json:"mints"`
	Total int             `json:"total"`
	Page  int             `json:"page"`
	Limit int             `json:"limit"`
}

func NewCreateMintRequest(mint protocol.MintWithoutID) CreateMintRequest {
	return CreateMintRequest{
		MintWithoutID: mint,
	}
}

func (m *CreateMintRequest) Deserialize(data []byte) error {
	return json.Unmarshal(data, m)
}

func (m *CreateMintRequest) Serialize() ([]byte, error) {
	return json.Marshal(m)
}

func (m *CreateMintResponse) Deserialize(data []byte) error {
	return json.Unmarshal(data, m)
}

func (m *CreateMintResponse) Serialize() ([]byte, error) {
	return json.Marshal(m)
}
