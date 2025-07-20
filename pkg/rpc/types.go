package rpc

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"dogecoin.org/fractal-engine/pkg/doge"
	"dogecoin.org/fractal-engine/pkg/store"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
)

type SignedRequest struct {
	PublicKey string `json:"public_key"`
	Signature string `json:"signature"`
}

func (req *SignedRequest) ValidateSignature(payloadBytes []byte) error {
	if req.PublicKey == "" {
		return fmt.Errorf("public_key is required")
	}
	if req.Signature == "" {
		return fmt.Errorf("signature is required")
	}

	// 2. Hash message
	hash := sha256.Sum256(payloadBytes)

	// 3. Decode public key
	pubKeyBytes, err := hex.DecodeString(req.PublicKey)
	if err != nil {
		return errors.New("invalid public key format")
	}

	pubKey, err := btcec.ParsePubKey(pubKeyBytes)
	if err != nil {
		return errors.New("failed to parse public key")
	}

	// 4. Decode signature
	sigBytes, err := hex.DecodeString(req.Signature)
	if err != nil {
		return errors.New("invalid signature encoding")
	}

	signature, err := ecdsa.ParseDERSignature(sigBytes)
	if err != nil {
		return errors.New("failed to parse DER signature")
	}

	// 5. Verify signature
	if !signature.Verify(hash[:], pubKey) {
		return errors.New("signature verification failed")
	}

	return nil
}

type PrepareMintRequest struct {
	Payload CreateMintRequestPayload `json:"payload"`
}

func (req *PrepareMintRequest) Validate() error {
	var missing []string

	if req.Payload.Title == "" {
		missing = append(missing, "title")
	}
	if req.Payload.FractionCount <= 0 {
		missing = append(missing, "fraction_count (must be > 0)")
	}
	if req.Payload.Description == "" {
		missing = append(missing, "description")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing or invalid fields: %s", strings.Join(missing, ", "))
	}
	return nil
}

type CreateMintRequest struct {
	Address   string                   `json:"address"`
	PublicKey string                   `json:"public_key"`
	Payload   CreateMintRequestPayload `json:"payload"`
	Signature string                   `json:"signature"`
}

type CreateMintRequestPayload struct {
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

	if req.Address == "" {
		missing = append(missing, "address")
	}
	if req.PublicKey == "" {
		missing = append(missing, "public_key")
	}

	payloadBytes, err := json.Marshal(req.Payload)
	if err != nil {
		return fmt.Errorf("invalid payload: %w", err)
	}

	if err := doge.ValidateSignature(payloadBytes, req.PublicKey, req.Signature); err != nil {
		return err
	}

	if req.Payload.Title == "" {
		missing = append(missing, "title")
	}
	if req.Payload.FractionCount <= 0 {
		missing = append(missing, "fraction_count (must be > 0)")
	}
	if req.Payload.Description == "" {
		missing = append(missing, "description")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing or invalid fields: %s", strings.Join(missing, ", "))
	}
	return nil
}

type CreateMintResponse struct {
	Hash string `json:"hash"`
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

type CreateBuyOfferRequest struct {
	SignedRequest
	Payload CreateBuyOfferRequestPayload `json:"payload"`
}

type CreateBuyOfferRequestPayload struct {
	OffererAddress string `json:"offerer_address"`
	SellerAddress  string `json:"seller_address"`
	MintHash       string `json:"mint_hash"`
	Quantity       int    `json:"quantity"`
	Price          int    `json:"price"`
}

type DeleteBuyOfferRequest struct {
	SignedRequest
	Payload DeleteBuyOfferRequestPayload `json:"payload"`
}

type DeleteBuyOfferRequestPayload struct {
	OfferHash string `json:"offer_hash"`
}

type DeleteSellOfferRequest struct {
	SignedRequest
	Payload DeleteSellOfferRequestPayload `json:"payload"`
}

type DeleteSellOfferRequestPayload struct {
	OfferHash string `json:"offer_hash"`
}

func (req *DeleteBuyOfferRequest) Validate() error {
	var missing []string

	if req.PublicKey == "" {
		missing = append(missing, "public_key")
	}
	if req.Signature == "" {
		missing = append(missing, "signature")
	}

	if req.Payload.OfferHash == "" {
		missing = append(missing, "offer_hash")
	}

	payloadBytes, err := json.Marshal(req.Payload)
	if err != nil {
		return fmt.Errorf("invalid payload: %w", err)
	}

	if err := req.ValidateSignature(payloadBytes); err != nil {
		return err
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing or invalid fields: %s", strings.Join(missing, ", "))
	}

	return nil
}

func (req *DeleteSellOfferRequest) Validate() error {
	var missing []string

	if req.PublicKey == "" {
		missing = append(missing, "public_key")
	}
	if req.Signature == "" {
		missing = append(missing, "signature")
	}

	if req.Payload.OfferHash == "" {
		missing = append(missing, "offer_hash")
	}

	payloadBytes, err := json.Marshal(req.Payload)
	if err != nil {
		return fmt.Errorf("invalid payload: %w", err)
	}

	if err := req.ValidateSignature(payloadBytes); err != nil {
		return err
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing or invalid fields: %s", strings.Join(missing, ", "))
	}

	return nil
}

func (req *CreateBuyOfferRequest) Validate() error {
	var missing []string

	if req.PublicKey == "" {
		missing = append(missing, "public_key")
	}
	if req.Signature == "" {
		missing = append(missing, "signature")
	}
	if req.Payload.OffererAddress == "" {
		missing = append(missing, "offerer_address")
	}
	if req.Payload.SellerAddress == "" {
		missing = append(missing, "seller_address")
	}
	if req.Payload.MintHash == "" {
		missing = append(missing, "mint_hash")
	}
	if req.Payload.Quantity <= 0 {
		missing = append(missing, "quantity (must be > 0)")
	}
	if req.Payload.Price <= 0 {
		missing = append(missing, "price (must be > 0)")
	}

	payloadBytes, err := json.Marshal(req.Payload)
	if err != nil {
		return fmt.Errorf("invalid payload: %w", err)
	}

	if err := req.ValidateSignature(payloadBytes); err != nil {
		return err
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing or invalid fields: %s", strings.Join(missing, ", "))
	}
	return nil
}

type CreateSellOfferRequest struct {
	SignedRequest
	Payload CreateSellOfferRequestPayload `json:"payload"`
}

type CreateSellOfferRequestPayload struct {
	OffererAddress string `json:"offerer_address"`
	MintHash       string `json:"mint_hash"`
	Quantity       int    `json:"quantity"`
	Price          int    `json:"price"`
}

func (req *CreateSellOfferRequest) Validate() error {
	var missing []string

	if req.PublicKey == "" {
		missing = append(missing, "public_key")
	}
	if req.Signature == "" {
		missing = append(missing, "signature")
	}
	if req.Payload.OffererAddress == "" {
		missing = append(missing, "offerer_address")
	}
	if req.Payload.MintHash == "" {
		missing = append(missing, "mint_hash")
	}
	if req.Payload.Quantity <= 0 {
		missing = append(missing, "quantity (must be > 0)")
	}
	if req.Payload.Price <= 0 {
		missing = append(missing, "price (must be > 0)")
	}

	payloadBytes, err := json.Marshal(req.Payload)
	if err != nil {
		return fmt.Errorf("invalid payload: %w", err)
	}

	if err := req.ValidateSignature(payloadBytes); err != nil {
		return err
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing or invalid fields: %s", strings.Join(missing, ", "))
	}
	return nil
}

type CreateOfferResponse struct {
	Id   string `json:"id"`
	Hash string `json:"hash"`
}

type GetSellOffersResponse struct {
	Offers []store.SellOffer `json:"offers"`
	Total  int               `json:"total"`
	Page   int               `json:"page"`
	Limit  int               `json:"limit"`
}

type GetBuyOffersResponse struct {
	Offers []store.BuyOffer `json:"offers"`
	Total  int              `json:"total"`
	Page   int              `json:"page"`
	Limit  int              `json:"limit"`
}

type CreateInvoiceRequest struct {
	SignedRequest
	Payload CreateInvoiceRequestPayload `json:"payload"`
}

type CreateInvoiceRequestPayload struct {
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

	if req.PublicKey == "" {
		missing = append(missing, "public_key")
	}
	if req.Signature == "" {
		missing = append(missing, "signature")
	}
	if req.Payload.PaymentAddress == "" {
		missing = append(missing, "payment_address")
	}

	if req.Payload.BuyOfferOffererAddress == "" {
		missing = append(missing, "buy_offer_offerer_address")
	}
	if req.Payload.BuyOfferHash == "" {
		missing = append(missing, "buy_offer_hash")
	}
	if req.Payload.BuyOfferMintHash == "" {
		missing = append(missing, "buy_offer_mint_hash")
	}
	if req.Payload.SellOfferAddress == "" {
		missing = append(missing, "sell_offer_address")
	}
	if req.Payload.BuyOfferQuantity <= 0 {
		missing = append(missing, "buy_offer_quantity (must be > 0)")
	}
	if req.Payload.BuyOfferPrice <= 0 {
		missing = append(missing, "buy_offer_price (must be > 0)")
	}

	payloadBytes, err := json.Marshal(req.Payload)
	if err != nil {
		return fmt.Errorf("invalid payload: %w", err)
	}

	if err := req.ValidateSignature(payloadBytes); err != nil {
		return err
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
	Chain              string    `json:"chain"`
	WalletsEnabled     bool      `json:"wallets_enabled"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type Address struct {
	Address    string `json:"address"`
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
	Label      string `json:"label"`
}

type SignTxRequest struct {
	Payload                PrepareMintRequest `json:"payload"`
	Signature              string             `json:"signature"`
	PublicKey              string             `json:"public_key"`
	EncodedTransactionBody string             `json:"encoded_transaction_body"`
}
