package store

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

type StringInterfaceMap map[string]interface{}

func (m *StringInterfaceMap) Scan(src interface{}) error {
	var source []byte
	switch src := src.(type) {
	case string:
		source = []byte(src)
	case []byte:
		source = src
	case nil:
		*m = nil
		return nil
	default:
		return fmt.Errorf("unsupported type: %T", src)
	}
	return json.Unmarshal(source, m)
}

type MintWithoutID struct {
	Hash            string             `json:"hash"`
	Title           string             `json:"title"`
	FractionCount   int                `json:"fraction_count"`
	Description     string             `json:"description"`
	Tags            StringArray        `json:"tags"`
	Metadata        StringInterfaceMap `json:"metadata"`
	TransactionHash sql.NullString     `json:"transaction_hash"`
	BlockHeight     int64              `json:"block_height"`
	CreatedAt       time.Time          `json:"created_at"`
	Requirements    StringInterfaceMap `json:"requirements"`
	LockupOptions   StringInterfaceMap `json:"lockup_options"`
	Gossiped        bool               `json:"gossiped"`
	FeedURL         string             `json:"feed_url"`
}

type MintHash struct {
	Title         string             `json:"title"`
	FractionCount int                `json:"fraction_count"`
	Description   string             `json:"description"`
	Tags          StringArray        `json:"tags"`
	Metadata      StringInterfaceMap `json:"metadata"`
	Requirements  StringInterfaceMap `json:"requirements"`
	LockupOptions StringInterfaceMap `json:"lockup_options"`
	OwnerAddress  string             `json:"owner_address"`
}

type OnChainTransaction struct {
	Id            string  `json:"id"`
	TxHash        string  `json:"tx_hash"`
	Height        int64   `json:"height"`
	ActionType    uint8   `json:"action_type"`
	ActionVersion uint8   `json:"action_version"`
	ActionData    []byte  `json:"action_data"`
	Address       string  `json:"address"`
	Value         float64 `json:"value"`
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
	Id           string `json:"id"`
	OwnerAddress string `json:"owner_address"`
}

type OnChainMint struct {
	MintId          string `json:"mint_id"`
	TransactionHash string `json:"transaction_hash"`
	Address         string `json:"address"`
}

type OfferType int32

const (
	OfferTypeBuy  OfferType = 0
	OfferTypeSell OfferType = 1
)

type OfferWithoutID struct {
	Hash           string    `json:"hash"`
	MintHash       string    `json:"mint_hash"`
	Type           OfferType `json:"type"`
	OffererAddress string    `json:"offerer_address"`
	Quantity       int       `json:"quantity"`
	Price          int       `json:"price"`
	CreatedAt      time.Time `json:"created_at"`
}

type OfferHash struct {
	Type     OfferType `json:"type"`
	MintHash string    `json:"mint_hash"`
	Quantity int       `json:"quantity"`
	Price    int       `json:"price"`
}

func (o *OfferWithoutID) GenerateHash() (string, error) {
	input := OfferHash{
		Type:     o.Type,
		MintHash: o.MintHash,
		Quantity: o.Quantity,
		Price:    o.Price,
	}

	jsonBytes, err := json.Marshal(input)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(jsonBytes)

	return hex.EncodeToString(hash[:]), nil
}

type Offer struct {
	OfferWithoutID
	Id string `json:"id"`
}

type UnconfirmedInvoice struct {
	Id                     string    `json:"id"`
	Hash                   string    `json:"hash"`
	BuyOfferOffererAddress string    `json:"buy_offer_offerer_address"`
	BuyOfferHash           string    `json:"buy_offer_hash"`
	BuyOfferMintHash       string    `json:"buy_offer_mint_hash"`
	BuyOfferQuantity       int       `json:"buy_offer_quantity"`
	BuyOfferPrice          int       `json:"buy_offer_price"`
	CreatedAt              time.Time `json:"created_at"`
	PaymentAddress         string    `json:"payment_address"`
	SellOfferAddress       string    `json:"sell_offer_address"`
	BuyOfferValue          float64   `json:"buy_offer_value"`
}

func (u *UnconfirmedInvoice) GenerateHash() (string, error) {
	input := UnconfirmedInvoiceHash{
		BuyOfferHash:           u.BuyOfferHash,
		BuyOfferMintHash:       u.BuyOfferMintHash,
		BuyOfferQuantity:       u.BuyOfferQuantity,
		BuyOfferPrice:          u.BuyOfferPrice,
		BuyOfferOffererAddress: u.BuyOfferOffererAddress,
		SellOfferAddress:       u.SellOfferAddress,
		BuyOfferValue:          u.BuyOfferValue,
	}

	jsonBytes, err := json.Marshal(input)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(jsonBytes)

	return hex.EncodeToString(hash[:]), nil
}

type UnconfirmedInvoiceHash struct {
	BuyOfferHash           string  `json:"buy_offer_hash"`
	BuyOfferMintHash       string  `json:"buy_offer_mint_hash"`
	BuyOfferQuantity       int     `json:"buy_offer_quantity"`
	BuyOfferPrice          int     `json:"buy_offer_price"`
	BuyOfferOffererAddress string  `json:"buy_offer_offerer_address"`
	PaymentAddress         string  `json:"payment_address"`
	SellOfferAddress       string  `json:"sell_offer_address"`
	BuyOfferValue          float64 `json:"buy_offer_value"`
}

type InvoiceHash struct {
	BuyOfferHash     string  `json:"buy_offer_hash"`
	BuyOfferMintHash string  `json:"buy_offer_mint_hash"`
	BuyOfferQuantity int     `json:"buy_offer_quantity"`
	BuyOfferPrice    int     `json:"buy_offer_price"`
	PaymentAddress   string  `json:"payment_address"`
	SellOfferAddress string  `json:"sell_offer_address"`
	BuyOfferValue    float64 `json:"buy_offer_value"`
}

type Invoice struct {
	Id                     string    `json:"id"`
	Hash                   string    `json:"hash"`
	PaymentAddress         string    `json:"payment_address"`
	BuyOfferOffererAddress string    `json:"buy_offer_offerer_address"`
	BuyOfferHash           string    `json:"buy_offer_hash"`
	BuyOfferMintHash       string    `json:"buy_offer_mint_hash"`
	BuyOfferQuantity       int       `json:"buy_offer_quantity"`
	BuyOfferPrice          int       `json:"buy_offer_price"`
	CreatedAt              time.Time `json:"created_at"`
	SellOfferAddress       string    `json:"sell_offer_address"`
	BuyOfferValue          float64   `json:"buy_offer_value"`
}

func (i *Invoice) GenerateHash() (string, error) {
	input := InvoiceHash{
		BuyOfferHash:     i.BuyOfferHash,
		BuyOfferMintHash: i.BuyOfferMintHash,
		BuyOfferQuantity: i.BuyOfferQuantity,
		BuyOfferPrice:    i.BuyOfferPrice,
		PaymentAddress:   i.PaymentAddress,
		SellOfferAddress: i.SellOfferAddress,
		BuyOfferValue:    i.BuyOfferValue,
	}

	jsonBytes, err := json.Marshal(input)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(jsonBytes)

	return hex.EncodeToString(hash[:]), nil
}

type TokenBalance struct {
	MintHash  string    `json:"mint_hash"`
	Address   string    `json:"address"`
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
