package rpc_test

import (
	"testing"

	"dogecoin.org/fractal-engine/internal/test/support"
	"dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/rpc"
	"dogecoin.org/fractal-engine/pkg/store"
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
	assert.Equal(t, invoices[0].Status, "draft")

	assert.Equal(t, len(dogenetClient.invoices), 1)
	assert.Equal(t, dogenetClient.invoices[0].Hash, invoiceResponse.Hash)
	assert.Equal(t, dogenetClient.invoices[0].PaymentAddress, invoice.Payload.PaymentAddress)
	assert.Equal(t, dogenetClient.invoices[0].BuyerAddress, invoice.Payload.BuyerAddress)
	assert.Equal(t, dogenetClient.invoices[0].MintHash, invoice.Payload.MintHash)
	assert.Equal(t, dogenetClient.invoices[0].Quantity, invoice.Payload.Quantity)
	assert.Equal(t, dogenetClient.invoices[0].Price, invoice.Payload.Price)
	assert.Equal(t, dogenetClient.invoices[0].SellerAddress, invoice.Payload.SellerAddress)
}

func TestInvoicesWithSignatureRequired(t *testing.T) {
	tokenisationStore, dogenetClient, mux, feClient := SetupRpcTest(t)
	rpc.HandleInvoiceRoutes(tokenisationStore, dogenetClient, mux, config.NewConfig())

	paymentAddress := support.GenerateDogecoinAddress(true)
	sellOfferAddress := support.GenerateDogecoinAddress(true)
	buyerAddress := support.GenerateDogecoinAddress(true)

	// Save a confirmed mint
	confirmedMint := &store.MintWithoutID{
		Title:                    "Confirmed Mint",
		FractionCount:            100,
		Description:              "Confirmed mint",
		Tags:                     store.StringArray{},
		Metadata:                 store.StringInterfaceMap{},
		Requirements:             store.StringInterfaceMap{},
		LockupOptions:            store.StringInterfaceMap{},
		PublicKey:                "testPubKey",
		TransactionHash:          "txHash",
		SignatureRequirementType: store.SignatureRequirementType_ALL_SIGNATURES,
		MinSignatures:            1,
		AssetManagers: store.AssetManagers{
			{
				Name:      "asset manager",
				PublicKey: "publicKey123",
				URL:       "https://example.com/assetManager",
			},
		},
	}
	var err error
	confirmedMint.Hash, err = confirmedMint.GenerateHash()
	assert.NilError(t, err)

	_, err = tokenisationStore.SaveMint(confirmedMint, "owner")
	assert.NilError(t, err)

	invoice := rpc.CreateInvoiceRequest{
		Payload: rpc.CreateInvoiceRequestPayload{
			PaymentAddress: paymentAddress,
			BuyerAddress:   buyerAddress,
			MintHash:       confirmedMint.Hash,
			Quantity:       10,
			Price:          100,
			SellerAddress:  sellOfferAddress,
		},
	}

	invoiceResponse, err := feClient.CreateInvoice(&invoice)
	if err != nil {
		t.Fatalf("Failed to create invoice: %v", err)
	}

	invoices, err := tokenisationStore.GetUnconfirmedInvoices(0, 10, confirmedMint.Hash, buyerAddress)
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
	assert.Equal(t, invoices[0].Status, "pending_signatures")

	assert.Equal(t, len(dogenetClient.invoices), 1)
	assert.Equal(t, dogenetClient.invoices[0].Hash, invoiceResponse.Hash)
	assert.Equal(t, dogenetClient.invoices[0].PaymentAddress, invoice.Payload.PaymentAddress)
	assert.Equal(t, dogenetClient.invoices[0].BuyerAddress, invoice.Payload.BuyerAddress)
	assert.Equal(t, dogenetClient.invoices[0].MintHash, invoice.Payload.MintHash)
	assert.Equal(t, dogenetClient.invoices[0].Quantity, invoice.Payload.Quantity)
	assert.Equal(t, dogenetClient.invoices[0].Price, invoice.Payload.Price)
	assert.Equal(t, dogenetClient.invoices[0].SellerAddress, invoice.Payload.SellerAddress)
}

func TestCreateInvoiceSignature(t *testing.T) {
	tokenisationStore, dogenetClient, mux, feClient := SetupRpcTest(t)
	rpc.HandleInvoiceRoutes(tokenisationStore, dogenetClient, mux, config.NewConfig())

	invoiceHash := support.GenerateRandomHash()
	signature := support.GenerateRandomHash()
	pubKey := support.GenerateRandomHash()

	createInvoiceSignatureRequest := rpc.CreateInvoiceSignatureRequest{
		Payload: rpc.CreateInvoiceSignatureRequestPayload{
			InvoiceHash: invoiceHash,
			Signature:   signature,
			PublicKey:   pubKey,
		},
	}

	createInvoiceSignatureResponse, err := feClient.CreateInvoiceSignature(&createInvoiceSignatureRequest)
	if err != nil {
		t.Fatalf("Failed to create invoice signature: %v", err)
	}

	var savedInvoiceHash string
	tokenisationStore.DB.QueryRow("SELECT invoice_hash FROM invoice_signatures WHERE id = $1", createInvoiceSignatureResponse.Id).Scan(&savedInvoiceHash)
	if err != nil {
		t.Fatalf("Failed to get invoice signature: %v", err)
	}

	assert.Equal(t, savedInvoiceHash, invoiceHash)
}
