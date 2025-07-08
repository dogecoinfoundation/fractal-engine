package rpc

import (
	"fmt"
	"strings"
	"time"

	"dogecoin.org/fractal-engine/pkg/store"
)

type CreateMintRequest struct {
	Title         string                   `json:"title"`
	FractionCount int                      `json:"fraction_count"`
	Description   string                   `json:"description"`
	Tags          store.StringArray        `json:"tags"`
	Metadata      store.StringInterfaceMap `json:"metadata"`
	Requirements  store.StringInterfaceMap `json:"requirements"`
	LockupOptions store.StringInterfaceMap `json:"lockup_options"`
	FeedURL       string                   `json:"feed_url"`
	OwnerAddress  string                   `json:"owner_address"`
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
	Type           store.OfferType `json:"type"`
	OffererAddress string          `json:"offerer_address"`
	MintHash       string          `json:"mint_hash"`
	Quantity       int             `json:"quantity"`
	Price          int             `json:"price"`
}

func (req *CreateOfferRequest) Validate() error {
	var missing []string

	if req.Type == store.OfferType(0) {
		missing = append(missing, "type")
	}
	if req.OffererAddress == "" {
		missing = append(missing, "offerer_address")
	}
	if req.MintHash == "" {
		missing = append(missing, "mint_hash")
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

type CreateInvoiceRequest struct {
	PaymentAddress         string `json:"payment_address"`
	BuyOfferOffererAddress string `json:"buy_offer_offerer_address"`
	BuyOfferHash           string `json:"buy_offer_hash"`
	BuyOfferMintHash       string `json:"buy_offer_mint_hash"`
	BuyOfferQuantity       int    `json:"buy_offer_quantity"`
	BuyOfferPrice          int    `json:"buy_offer_price"`
	SellOfferAddress       string `json:"sell_offer_address"`
}

func (req *CreateInvoiceRequest) Validate() error {
	var missing []string

	if req.PaymentAddress == "" {
		missing = append(missing, "payment_address")
	}

	if req.BuyOfferOffererAddress == "" {
		missing = append(missing, "buy_offer_offerer_address")
	}
	if req.BuyOfferHash == "" {
		missing = append(missing, "buy_offer_hash")
	}
	if req.BuyOfferMintHash == "" {
		missing = append(missing, "buy_offer_mint_hash")
	}
	if req.SellOfferAddress == "" {
		missing = append(missing, "sell_offer_address")
	}
	if req.BuyOfferQuantity <= 0 {
		missing = append(missing, "buy_offer_quantity (must be > 0)")
	}
	if req.BuyOfferPrice <= 0 {
		missing = append(missing, "buy_offer_price (must be > 0)")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing or invalid fields: %s", strings.Join(missing, ", "))
	}

	return nil
}

type GetInvoicesResponse struct {
	Invoices []store.Invoice `json:"invoices"`
	Total    int             `json:"total"`
	Page     int             `json:"page"`
	Limit    int             `json:"limit"`
}

type CreateInvoiceResponse struct {
	EncodedTransactionBody string `json:"encoded_transaction_body"`
	TransactionHash        string `json:"transaction_hash"`
	Id                     string `json:"id"`
}

type GetHealthResponse struct {
	CurrentBlockHeight int64     `json:"current_block_height"`
	LatestBlockHeight  int64     `json:"latest_block_height"`
	UpdatedAt          time.Time `json:"updated_at"`
}
