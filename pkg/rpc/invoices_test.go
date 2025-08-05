package rpc_test

import (
	"testing"

	"dogecoin.org/fractal-engine/internal/test/support"
	"dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/rpc"
	"gotest.tools/assert"
)

func TestInvoices(t *testing.T) {
	tokenisationStore, dogenetClient, mux, feClient := SetupRpcTest(t)
	rpc.HandleInvoiceRoutes(tokenisationStore, dogenetClient, mux, config.NewConfig())

	paymentAddress := support.GenerateDogecoinAddress(true)
	buyOfferOffererAddress := support.GenerateDogecoinAddress(true)
	sellOfferAddress := support.GenerateDogecoinAddress(true)

	buyOfferMintHash := support.GenerateRandomHash()

	invoice := rpc.CreateInvoiceRequest{
		Payload: rpc.CreateInvoiceRequestPayload{
			PaymentAddress:         paymentAddress,
			BuyOfferOffererAddress: buyOfferOffererAddress,
			BuyOfferHash:           support.GenerateRandomHash(),
			BuyOfferMintHash:       buyOfferMintHash,
			BuyOfferQuantity:       10,
			BuyOfferPrice:          100,
			SellOfferAddress:       sellOfferAddress,
		},
	}

	invoiceResponse, err := feClient.CreateInvoice(&invoice)
	if err != nil {
		t.Fatalf("Failed to create invoice: %v", err)
	}

	invoices, err := tokenisationStore.GetUnconfirmedInvoices(0, 10, buyOfferMintHash, buyOfferOffererAddress)
	if err != nil {
		t.Fatalf("Failed to get invoices: %v", err)
	}

	assert.Equal(t, len(invoices), 1)
	assert.Equal(t, invoices[0].Hash, invoiceResponse.Hash)
	assert.Equal(t, invoices[0].PaymentAddress, invoice.Payload.PaymentAddress)
	assert.Equal(t, invoices[0].BuyOfferOffererAddress, invoice.Payload.BuyOfferOffererAddress)
	assert.Equal(t, invoices[0].BuyOfferHash, invoice.Payload.BuyOfferHash)
	assert.Equal(t, invoices[0].BuyOfferMintHash, invoice.Payload.BuyOfferMintHash)
	assert.Equal(t, invoices[0].BuyOfferQuantity, invoice.Payload.BuyOfferQuantity)
	assert.Equal(t, invoices[0].BuyOfferPrice, invoice.Payload.BuyOfferPrice)
	assert.Equal(t, invoices[0].SellOfferAddress, invoice.Payload.SellOfferAddress)

	assert.Equal(t, len(dogenetClient.invoices), 1)
	assert.Equal(t, dogenetClient.invoices[0].Hash, invoiceResponse.Hash)
	assert.Equal(t, dogenetClient.invoices[0].PaymentAddress, invoice.Payload.PaymentAddress)
	assert.Equal(t, dogenetClient.invoices[0].BuyOfferOffererAddress, invoice.Payload.BuyOfferOffererAddress)
	assert.Equal(t, dogenetClient.invoices[0].BuyOfferHash, invoice.Payload.BuyOfferHash)
	assert.Equal(t, dogenetClient.invoices[0].BuyOfferMintHash, invoice.Payload.BuyOfferMintHash)
	assert.Equal(t, dogenetClient.invoices[0].BuyOfferQuantity, invoice.Payload.BuyOfferQuantity)
	assert.Equal(t, dogenetClient.invoices[0].BuyOfferPrice, invoice.Payload.BuyOfferPrice)
	assert.Equal(t, dogenetClient.invoices[0].SellOfferAddress, invoice.Payload.SellOfferAddress)
}
