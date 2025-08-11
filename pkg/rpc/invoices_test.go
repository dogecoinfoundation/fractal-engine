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
	MintHash := support.GenerateDogecoinAddress(true)
	sellOfferAddress := support.GenerateDogecoinAddress(true)

	buyOfferMintHash := support.GenerateRandomHash()

	invoice := rpc.CreateInvoiceRequest{
		Payload: rpc.CreateInvoiceRequestPayload{
			PaymentAddress: paymentAddress,
			BuyerAddress:   MintHash,
			MintHash:       buyOfferMintHash,
			Quantity:       10,
			Price:          100,
			SellerAddress:  sellOfferAddress,
		},
	}

	invoiceResponse, err := feClient.CreateInvoice(&invoice)
	if err != nil {
		t.Fatalf("Failed to create invoice: %v", err)
	}

	invoices, err := tokenisationStore.GetUnconfirmedInvoices(0, 10, buyOfferMintHash, MintHash)
	if err != nil {
		t.Fatalf("Failed to get invoices: %v", err)
	}

	assert.Equal(t, len(invoices), 1)
	assert.Equal(t, invoices[0].Hash, invoiceResponse.Hash)
	assert.Equal(t, invoices[0].PaymentAddress, invoice.Payload.PaymentAddress)
	assert.Equal(t, invoices[0].BuyerAddress, invoice.Payload.BuyerAddress)
	assert.Equal(t, invoices[0].MintHash, invoice.Payload.MintHash)
	assert.Equal(t, invoices[0].Quantity, invoice.Payload.Quantity)
	assert.Equal(t, invoices[0].Price, invoice.Payload.Price)
	assert.Equal(t, invoices[0].SellerAddress, invoice.Payload.SellerAddress)

	assert.Equal(t, len(dogenetClient.invoices), 1)
	assert.Equal(t, dogenetClient.invoices[0].Hash, invoiceResponse.Hash)
	assert.Equal(t, dogenetClient.invoices[0].PaymentAddress, invoice.Payload.PaymentAddress)
	assert.Equal(t, dogenetClient.invoices[0].BuyerAddress, invoice.Payload.BuyerAddress)
	assert.Equal(t, dogenetClient.invoices[0].MintHash, invoice.Payload.MintHash)
	assert.Equal(t, dogenetClient.invoices[0].Quantity, invoice.Payload.Quantity)
	assert.Equal(t, dogenetClient.invoices[0].Price, invoice.Payload.Price)
	assert.Equal(t, dogenetClient.invoices[0].SellerAddress, invoice.Payload.SellerAddress)
}
