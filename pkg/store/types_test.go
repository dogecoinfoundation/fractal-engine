package store_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"

	"dogecoin.org/fractal-engine/internal/test/support"
	"dogecoin.org/fractal-engine/pkg/store"
	"gotest.tools/assert"
)

func TestInvoiceGenerateHash(t *testing.T) {
	MintHash := support.GenerateDogecoinAddress(true)
	sellOfferAddress := support.GenerateDogecoinAddress(true)

	invoice := store.UnconfirmedInvoice{
		MintHash:      "buyOfferMintHash",
		Quantity:      100,
		Price:         20,
		BuyerAddress:  MintHash,
		SellerAddress: sellOfferAddress,
		PublicKey:     "publicKey",
	}

	inputHash := store.UnconfirmedInvoiceHash{
		MintHash:      invoice.MintHash,
		Quantity:      invoice.Quantity,
		Price:         invoice.Price,
		BuyerAddress:  invoice.BuyerAddress,
		SellerAddress: invoice.SellerAddress,
		PublicKey:     invoice.PublicKey,
	}

	invoiceBytes, err := json.Marshal(inputHash)
	assert.NilError(t, err)

	expectedHashBytes := sha256.Sum256(invoiceBytes)
	expectedHash := hex.EncodeToString(expectedHashBytes[:])

	hash, err := invoice.GenerateHash()
	assert.NilError(t, err)
	assert.Equal(t, expectedHash, hash)
}
