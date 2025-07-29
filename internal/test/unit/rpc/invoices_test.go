package test_rpc

import (
	"testing"

	"dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/rpc"
	"gotest.tools/assert"
)

func TestInvoices(t *testing.T) {
	tokenisationStore, dogenetClient, mux, feClient := SetupRpcTest(t)
	rpc.HandleInvoiceRoutes(tokenisationStore, dogenetClient, mux, config.NewConfig())

	invoice := rpc.CreateInvoiceRequest{
		Payload: rpc.CreateInvoiceRequestPayload{
			PaymentAddress:         "0x122122121212121",
			BuyOfferOffererAddress: "0x122122121213333",
			BuyOfferHash:           "0x122122121212121",
			BuyOfferMintHash:       "0x122122121212XXX",
			BuyOfferQuantity:       10,
			BuyOfferPrice:          100,
			SellOfferAddress:       "0x122122121212121",
		},
	}

	invoiceResponse, err := feClient.CreateInvoice(&invoice)
	if err != nil {
		t.Fatalf("Failed to create invoice: %v", err)
	}

	invoices, err := tokenisationStore.GetUnconfirmedInvoices(0, 10, "0x122122121212XXX", "0x122122121213333")
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
