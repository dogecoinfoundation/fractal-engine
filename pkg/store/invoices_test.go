package store

import (
	"testing"
	"time"

	"dogecoin.org/fractal-engine/pkg/config"

	"gotest.tools/assert"
)

func TestSaveAndGetInvoices(t *testing.T) {
	tokenisationStore, err := NewTokenisationStore("memory://invoices.db", config.Config{
		MigrationsPath: "../../db/migrations",
	})
	if err != nil {
		panic(err)
	}

	err = tokenisationStore.Migrate()
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	invoice := Invoice{
		Id:                     "myId",
		Hash:                   "myHash",
		PaymentAddress:         "myPaymentAddress",
		BuyOfferOffererAddress: "myOffererAddress",
		BuyOfferHash:           "myBuyOfferHash",
		BuyOfferMintHash:       "myMintHash",
		BuyOfferQuantity:       10,
		BuyOfferPrice:          25,
		CreatedAt:              time.Now(),
	}

	id, err := tokenisationStore.SaveInvoice(&invoice)
	if err != nil {
		t.Fatalf("Failed to save invoice: %v", err)
	}

	invoices, err := tokenisationStore.GetInvoices(0, 10, "myMintHash", "myOffererAddress")
	if err != nil {
		t.Fatalf("Failed to get offer: %v", err)
	}

	assert.Equal(t, len(invoices), 1)
	assert.Equal(t, invoices[0].Id, id, "failed to match invoice id")
	assert.Equal(t, invoices[0].Hash, invoice.Hash, "failed to match invoice hash")
	assert.Equal(t, invoices[0].PaymentAddress, invoice.PaymentAddress, "failed to match invoice payment address")
	assert.Equal(t, invoices[0].BuyOfferOffererAddress, invoice.BuyOfferOffererAddress, "failed to match invoice offerer address")
	assert.Equal(t, invoices[0].BuyOfferHash, invoice.BuyOfferHash, "failed to match invoice buy offer hash")
	assert.Equal(t, invoices[0].BuyOfferMintHash, invoice.BuyOfferMintHash, "failed to match invoice buy offer mint hash")
	assert.Equal(t, invoices[0].BuyOfferQuantity, invoice.BuyOfferQuantity, "failed to match invoice buy offer quantity")
	assert.Equal(t, invoices[0].BuyOfferPrice, invoice.BuyOfferPrice, "failed to match invoice buy offer price")
	assert.Equal(t, invoices[0].CreatedAt.Unix(), invoice.CreatedAt.Unix(), "failed to match invoice created at")
}
