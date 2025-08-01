package store_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"

	"dogecoin.org/fractal-engine/pkg/store"
	"gotest.tools/assert"
)

func TestInvoiceGenerateHash(t *testing.T) {
	invoice := store.UnconfirmedInvoice{
		BuyOfferHash:           "buyOfferHash",
		BuyOfferMintHash:       "buyOfferMintHash",
		BuyOfferQuantity:       100,
		BuyOfferPrice:          20,
		BuyOfferOffererAddress: "buyOfferOffererAddress",
		SellOfferAddress:       "sellOfferAddress",
		BuyOfferValue:          30,
		PublicKey:              "publicKey",
	}

	inputHash := store.UnconfirmedInvoiceHash{
		BuyOfferHash:           invoice.BuyOfferHash,
		BuyOfferMintHash:       invoice.BuyOfferMintHash,
		BuyOfferQuantity:       invoice.BuyOfferQuantity,
		BuyOfferPrice:          invoice.BuyOfferPrice,
		BuyOfferOffererAddress: invoice.BuyOfferOffererAddress,
		SellOfferAddress:       invoice.SellOfferAddress,
		BuyOfferValue:          invoice.BuyOfferValue,
		PublicKey:              invoice.PublicKey,
	}

	invoiceBytes, err := json.Marshal(inputHash)
	assert.NilError(t, err)

	expectedHashBytes := sha256.Sum256(invoiceBytes)
	expectedHash := hex.EncodeToString(expectedHashBytes[:])

	hash, err := invoice.GenerateHash()
	assert.NilError(t, err)
	assert.Equal(t, expectedHash, hash)
}
