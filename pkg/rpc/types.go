package rpc

import (
	"fmt"
	"strings"

	"dogecoin.org/fractal-engine/pkg/store"
)

type CreateMintRequest struct {
	store.MintWithoutID
}

func (req *CreateMintRequest) Validate() error {
	var missing []string

	if req.Title == "" {
		missing = append(missing, "title")
	}
	if req.FractionCount <= 0 {
		missing = append(missing, "fraction_count (must be > 0)")
	}
	if req.Description == "" {
		missing = append(missing, "description")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing or invalid fields: %s", strings.Join(missing, ", "))
	}
	return nil
}

type CreateMintResponse struct {
	EncodedTransactionBody string `json:"encoded_transaction_body"`
	TransactionHash        string `json:"transaction_hash"`
	Id                     string `json:"id"`
}

type GetMintsResponse struct {
	Mints []store.Mint `json:"mints"`
	Total int          `json:"total"`
	Page  int          `json:"page"`
	Limit int          `json:"limit"`
}

type GetStatsResponse struct {
	Stats map[string]int `json:"stats"`
}

type CreateOfferRequest struct {
	store.OfferWithoutID
}

func (req *CreateOfferRequest) Validate() error {
	var missing []string

	if req.Type == store.OfferType(0) {
		missing = append(missing, "type")
	}
	if req.OffererAddress == "" {
		missing = append(missing, "offerer_address")
	}
	if req.Hash == "" {
		missing = append(missing, "hash")
	}
	if req.Quantity <= 0 {
		missing = append(missing, "quantity (must be > 0)")
	}
	if req.Price <= 0 {
		missing = append(missing, "price (must be > 0)")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing or invalid fields: %s", strings.Join(missing, ", "))
	}
	return nil
}

type CreateOfferResponse struct {
	Id string `json:"id"`
}

type GetOffersResponse struct {
	Offers []store.Offer `json:"offers"`
	Total  int           `json:"total"`
	Page   int           `json:"page"`
	Limit  int           `json:"limit"`
}
