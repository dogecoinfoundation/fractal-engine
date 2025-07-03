package test_service

import (
	"database/sql"
	"testing"
	"time"

	test_support "dogecoin.org/fractal-engine/internal/test/support"
	"dogecoin.org/fractal-engine/pkg/protocol"
	"dogecoin.org/fractal-engine/pkg/service"
	"dogecoin.org/fractal-engine/pkg/store"
	"google.golang.org/protobuf/proto"
	"gotest.tools/assert"
)

func TestMintMatch(t *testing.T) {
	tokenisationStore := test_support.SetupTestDB(t)

	hash := SetupUnconfirmedMint(t, tokenisationStore)

	// tokenisationStore.DebugPrintStore()

	processor := service.NewFractalEngineProcessor(tokenisationStore)
	processor.Process()

	// tokenisationStore.DebugPrintStore()

	mints, err := tokenisationStore.GetMints(0, 100)
	if err != nil {
		t.Fatalf("Failed to get mints: %v", err)
	}

	// tokenisationStore.DebugPrintStore()

	assert.Equal(t, 1, len(mints))
	assert.Equal(t, "Test Mint", mints[0].Title)

	tokenBalance, err := tokenisationStore.GetTokenBalance("ownerAddress", hash)
	if err != nil {
		t.Fatalf("Failed to get token balance: %v", err)
	}

	assert.Equal(t, 100, tokenBalance)
}

func TestInvoiceMatch(t *testing.T) {
	tokenisationStore := test_support.SetupTestDB(t)

	hash := SetupUnconfirmedMint(t, tokenisationStore)

	processor := service.NewFractalEngineProcessor(tokenisationStore)
	processor.Process()

	tokenBalance, err := tokenisationStore.GetTokenBalance("ownerAddress", hash)
	if err != nil {
		t.Fatalf("Failed to get token balance: %v", err)
	}

	assert.Equal(t, 100, tokenBalance)

	message := protocol.OnChainInvoiceMessage{
		SellOfferAddress: "ownerAddress",
		InvoiceHash:      "invoiceHash",
		MintHash:         hash,
		Quantity:         50,
	}

	encodedMessage, err := proto.Marshal(&message)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}

	_, err = tokenisationStore.SaveOnChainTransaction("txHash002", 1, 1, protocol.ACTION_INVOICE, protocol.DEFAULT_VERSION, encodedMessage, "ownerAddress", 100)
	if err != nil {
		t.Fatalf("Failed to save on chain transaction: %v", err)
	}

	processor.Process()

	tokenBalance, err = tokenisationStore.GetTokenBalance("ownerAddress", hash)
	if err != nil {
		t.Fatalf("Failed to get token balance: %v", err)
	}

	assert.Equal(t, 50, tokenBalance)

	message2 := protocol.OnChainInvoiceMessage{
		SellOfferAddress: "ownerAddress",
		InvoiceHash:      "invoiceHash",
		MintHash:         hash,
		Quantity:         90,
	}

	encodedMessage2, err := proto.Marshal(&message2)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}

	_, err = tokenisationStore.SaveOnChainTransaction("txHash003", 1, 1, protocol.ACTION_INVOICE, protocol.DEFAULT_VERSION, encodedMessage2, "ownerAddress", 100)
	if err != nil {
		t.Fatalf("Failed to save on chain transaction: %v", err)
	}

	processor.Process()

	_, err = tokenisationStore.SaveUnconfirmedInvoice(&store.UnconfirmedInvoice{
		Hash:                   "invoiceHash",
		PaymentAddress:         "ownerAddress",
		BuyOfferOffererAddress: "buyerAddress",
		BuyOfferHash:           "buyOfferHash",
		BuyOfferMintHash:       hash,
		BuyOfferQuantity:       50,
		BuyOfferPrice:          100,
		BuyOfferValue:          100,
		CreatedAt:              time.Now(),
		SellOfferAddress:       "ownerAddress",
	})

	if err != nil {
		t.Fatalf("Failed to save invoice: %v", err)
	}

	processor.Process()

	tokenBalance, err = tokenisationStore.GetTokenBalance("ownerAddress", hash)
	if err != nil {
		t.Fatalf("Failed to get token balance: %v", err)
	}

	assert.Equal(t, 50, tokenBalance)

	pendingTokenBalance, err := tokenisationStore.GetPendingTokenBalance("invoiceHash", hash)
	if err != nil {
		t.Fatalf("Failed to get pending token balance: %v", err)
	}
	assert.Equal(t, 50, pendingTokenBalance.Quantity)

	message3 := protocol.OnChainPaymentMessage{
		Hash: "invoiceHash",
	}

	encodedMessage3, err := proto.Marshal(&message3)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}

	_, err = tokenisationStore.SaveOnChainTransaction("txHash004", 1, 1, protocol.ACTION_PAYMENT, protocol.DEFAULT_VERSION, encodedMessage3, "ownerAddress", 100)
	if err != nil {
		t.Fatalf("Failed to save on chain transaction: %v", err)
	}

	processor.Process()

	tokenBalance, err = tokenisationStore.GetTokenBalance("buyerAddress", hash)
	if err != nil {
		t.Fatalf("Failed to get token balance: %v", err)
	}

	assert.Equal(t, 50, tokenBalance)
}

func SetupUnconfirmedMint(t *testing.T, tokenisationStore *store.TokenisationStore) string {
	mintWithoutId := store.MintWithoutID{
		Title:         "Test Mint",
		Description:   "Test Description",
		Tags:          []string{"test", "test2"},
		FeedURL:       "https://test.com",
		FractionCount: 100,
		BlockHeight:   20,
		TransactionHash: sql.NullString{
			String: "txHash001",
			Valid:  true,
		},
	}
	hash, err := mintWithoutId.GenerateHash()
	if err != nil {
		t.Fatalf("Failed to generate hash: %v", err)
	}
	mintWithoutId.Hash = hash

	_, err = tokenisationStore.SaveUnconfirmedMint(&mintWithoutId)
	if err != nil {
		t.Fatalf("Failed to save unconfirmed mint: %v", err)
	}

	message := protocol.OnChainMintMessage{
		Hash: hash,
	}

	encodedMessage, err := proto.Marshal(&message)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}

	_, err = tokenisationStore.SaveOnChainTransaction("txHash001", 1, 1, protocol.ACTION_MINT, protocol.DEFAULT_VERSION, encodedMessage, "ownerAddress", 100)
	if err != nil {
		t.Fatalf("Failed to save on chain transaction: %v", err)
	}

	return hash
}
