package rpc

import (
	"dogecoin.org/fractal-engine/pkg/store"
)

type CreateMintRequest struct {
	TokenID string `json:"token_id"`
}

type CreateMintResponse struct {
	UnsignedTransaction string `json:"unsigned_transaction"`
}

type GetMintsResponse struct {
	Mints []store.Mint `json:"mints"`
	Total int          `json:"total"`
	Page  int          `json:"page"`
	Limit int          `json:"limit"`
}
