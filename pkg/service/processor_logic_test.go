package service_test

import (
	"database/sql"
	"fmt"
	"testing"

	test_support "dogecoin.org/fractal-engine/internal/test/support"
	"dogecoin.org/fractal-engine/pkg/protocol"
	"dogecoin.org/fractal-engine/pkg/service"
	"dogecoin.org/fractal-engine/pkg/store"
	"google.golang.org/protobuf/proto"
)

func TestProcessEmptyDatabase(t *testing.T) {
	tokenStore := test_support.SetupTestDB()
	processor := service.NewFractalEngineProcessor(tokenStore)

	// Process with empty database should complete without error
	err := processor.Process()
	if err != nil {
		t.Fatalf("Expected no error processing empty database, got: %v", err)
	}
}

func TestProcessMintTransactionMatched(t *testing.T) {
	tokenStore := test_support.SetupTestDB()
	processor := service.NewFractalEngineProcessor(tokenStore)

	// Create a mint that will be matched
	mintHash := "mint123"
	mintMsg := &protocol.OnChainMintMessage{
		Hash: mintHash,
	}
	encodedMsg, _ := proto.Marshal(mintMsg)

	// First create an unconfirmed mint
	_, err := tokenStore.SaveUnconfirmedMint(&store.MintWithoutID{
		Hash:          mintHash,
		Title:         "Test Mint",
		Description:   "Test Description",
		FractionCount: 100,
		BlockHeight:   1,
		TransactionHash: sql.NullString{
			String: "txHash001",
			Valid:  true,
		},
	})
	if err != nil {
		t.Fatalf("Failed to save unconfirmed mint: %v", err)
	}

	// Create the on-chain transaction
	_, err = tokenStore.SaveOnChainTransaction("txHash001", 1, 1, protocol.ACTION_MINT, protocol.DEFAULT_VERSION, encodedMsg, "ownerAddress", 100)
	if err != nil {
		t.Fatalf("Failed to save on-chain transaction: %v", err)
	}

	// Process should match the mint
	err = processor.Process()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify mint was matched by checking token balances
	balances, err := tokenStore.GetTokenBalances("ownerAddress", mintHash)
	if err != nil {
		t.Fatalf("Failed to get token balances: %v", err)
	}

	if len(balances) == 0 {
		t.Error("Expected token balance to be created")
	}

	totalBalance := 0
	for _, balance := range balances {
		totalBalance += balance.Quantity
	}
	if totalBalance != 100 {
		t.Errorf("Expected total balance of 100, got %d", totalBalance)
	}
}

func TestProcessMintTransactionNotMatched(t *testing.T) {
	tokenStore := test_support.SetupTestDB()
	processor := service.NewFractalEngineProcessor(tokenStore)

	// Create an on-chain mint transaction without unconfirmed mint
	mintHash := "mint456"
	mintMsg := &protocol.OnChainMintMessage{
		Hash: mintHash,
	}
	encodedMsg, _ := proto.Marshal(mintMsg)

	_, err := tokenStore.SaveOnChainTransaction("txHash002", 1, 1, protocol.ACTION_MINT, protocol.DEFAULT_VERSION, encodedMsg, "ownerAddress", 100)
	if err != nil {
		t.Fatalf("Failed to save on-chain transaction: %v", err)
	}

	// Process should not match (no unconfirmed mint exists)
	err = processor.Process()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify no token balance was created
	balances, err := tokenStore.GetTokenBalances("ownerAddress", mintHash)
	if err != nil {
		t.Fatalf("Failed to get token balances: %v", err)
	}

	if len(balances) != 0 {
		t.Error("Expected no token balance to be created")
	}
}

func TestProcessPaymentTransaction(t *testing.T) {
	tokenStore := test_support.SetupTestDB()
	processor := service.NewFractalEngineProcessor(tokenStore)

	// Setup: Create a mint and invoice first
	mintHash := "mint789"
	invoiceHash := "invoice123"

	// Create unconfirmed mint
	_, err := tokenStore.SaveUnconfirmedMint(&store.MintWithoutID{
		Hash:          mintHash,
		Title:         "Test Mint",
		Description:   "Test Description",
		FractionCount: 100,
		BlockHeight:   1,
		TransactionHash: sql.NullString{
			String: "txMint",
			Valid:  true,
		},
	})
	if err != nil {
		t.Fatalf("Failed to save unconfirmed mint: %v", err)
	}

	// Create and process mint transaction
	mintMsg := &protocol.OnChainMintMessage{Hash: mintHash}
	encodedMintMsg, _ := proto.Marshal(mintMsg)
	_, err = tokenStore.SaveOnChainTransaction("txMint", 1, 1, protocol.ACTION_MINT, protocol.DEFAULT_VERSION, encodedMintMsg, "sellerAddress", 100)
	if err != nil {
		t.Fatalf("Failed to save mint transaction: %v", err)
	}
	processor.Process()

	// Create unconfirmed invoice
	_, err = tokenStore.SaveUnconfirmedInvoice(&store.UnconfirmedInvoice{
		Hash:                   invoiceHash,
		PaymentAddress:         "sellerAddress",
		BuyOfferOffererAddress: "buyerAddress",
		BuyOfferHash:           "buyOffer123",
		BuyOfferMintHash:       mintHash,
		BuyOfferQuantity:       50,
		BuyOfferPrice:          100,
		BuyOfferValue:          50,
		SellOfferAddress:       "sellerAddress",
	})
	if err != nil {
		t.Fatalf("Failed to save unconfirmed invoice: %v", err)
	}

	// Create and process invoice transaction
	invoiceMsg := &protocol.OnChainInvoiceMessage{
		SellOfferAddress: "sellerAddress",
		InvoiceHash:      invoiceHash,
		MintHash:         mintHash,
		Quantity:         50,
	}
	encodedInvoiceMsg, _ := proto.Marshal(invoiceMsg)
	_, err = tokenStore.SaveOnChainTransaction("txInvoice", 2, 1, protocol.ACTION_INVOICE, protocol.DEFAULT_VERSION, encodedInvoiceMsg, "sellerAddress", 50)
	if err != nil {
		t.Fatalf("Failed to save invoice transaction: %v", err)
	}
	processor.Process()

	// Create payment transaction
	paymentMsg := &protocol.OnChainPaymentMessage{
		Hash: invoiceHash,
	}
	encodedPaymentMsg, _ := proto.Marshal(paymentMsg)
	_, err = tokenStore.SaveOnChainTransaction("txPayment", 3, 1, protocol.ACTION_PAYMENT, protocol.DEFAULT_VERSION, encodedPaymentMsg, "buyerAddress", 50)
	if err != nil {
		t.Fatalf("Failed to save payment transaction: %v", err)
	}

	// Process payment
	err = processor.Process()
	if err != nil {
		t.Fatalf("Expected no error processing payment, got: %v", err)
	}

	// Verify buyer received tokens
	buyerBalances, err := tokenStore.GetTokenBalances("buyerAddress", mintHash)
	if err != nil {
		t.Fatalf("Failed to get buyer token balances: %v", err)
	}

	buyerTotal := 0
	for _, balance := range buyerBalances {
		buyerTotal += balance.Quantity
	}
	if buyerTotal != 50 {
		t.Errorf("Expected buyer to have 50 tokens, got %d", buyerTotal)
	}
}

func TestProcessInvoiceTransaction(t *testing.T) {
	tokenStore := test_support.SetupTestDB()
	processor := service.NewFractalEngineProcessor(tokenStore)

	// Setup: Create a mint first
	mintHash := "mintABC"

	// Create unconfirmed mint
	_, err := tokenStore.SaveUnconfirmedMint(&store.MintWithoutID{
		Hash:          mintHash,
		Title:         "Test Mint",
		Description:   "Test Description",
		FractionCount: 100,
		BlockHeight:   1,
		TransactionHash: sql.NullString{
			String: "txMint",
			Valid:  true,
		},
	})
	if err != nil {
		t.Fatalf("Failed to save unconfirmed mint: %v", err)
	}

	// Create and process mint transaction
	mintMsg := &protocol.OnChainMintMessage{Hash: mintHash}
	encodedMintMsg, _ := proto.Marshal(mintMsg)
	_, err = tokenStore.SaveOnChainTransaction("txMint", 1, 1, protocol.ACTION_MINT, protocol.DEFAULT_VERSION, encodedMintMsg, "sellerAddress", 100)
	if err != nil {
		t.Fatalf("Failed to save mint transaction: %v", err)
	}
	processor.Process()

	// Create invoice transaction
	invoiceHash := "invoiceXYZ"
	invoiceMsg := &protocol.OnChainInvoiceMessage{
		SellOfferAddress: "sellerAddress",
		InvoiceHash:      invoiceHash,
		MintHash:         mintHash,
		Quantity:         30,
	}
	encodedInvoiceMsg, _ := proto.Marshal(invoiceMsg)
	_, err = tokenStore.SaveOnChainTransaction("txInvoice", 2, 1, protocol.ACTION_INVOICE, protocol.DEFAULT_VERSION, encodedInvoiceMsg, "sellerAddress", 30)
	if err != nil {
		t.Fatalf("Failed to save invoice transaction: %v", err)
	}

	// Process invoice
	err = processor.Process()
	if err != nil {
		t.Fatalf("Expected no error processing invoice, got: %v", err)
	}

	// Verify pending token balance was created
	tx, _ := tokenStore.DB.Begin()
	defer tx.Rollback()

	pendingBalance, err := tokenStore.GetPendingTokenBalance(invoiceHash, mintHash, tx)
	if err != nil {
		t.Fatalf("Failed to get pending token balance: %v", err)
	}

	if pendingBalance.Quantity != 30 {
		t.Errorf("Expected pending balance of 30, got %d", pendingBalance.Quantity)
	}
}

func TestProcessPagination(t *testing.T) {
	tokenStore := test_support.SetupTestDB()
	processor := service.NewFractalEngineProcessor(tokenStore)

	// Create 150 transactions to test pagination (limit is 100)
	for i := 0; i < 150; i++ {
		mintHash := fmt.Sprintf("mint%d", i)
		mintMsg := &protocol.OnChainMintMessage{Hash: mintHash}
		encodedMsg, _ := proto.Marshal(mintMsg)

		txHash := fmt.Sprintf("tx%d", i)
		_, err := tokenStore.SaveOnChainTransaction(txHash, int64(i+1), 1, protocol.ACTION_MINT, protocol.DEFAULT_VERSION, encodedMsg, "ownerAddress", 1)
		if err != nil {
			t.Fatalf("Failed to save transaction %d: %v", i, err)
		}
	}

	// Process should handle pagination correctly
	err := processor.Process()
	if err != nil {
		t.Fatalf("Expected no error with pagination, got: %v", err)
	}

	// Verify all transactions were processed
	txs, err := tokenStore.GetOnChainTransactions(0, 200)
	if err != nil {
		t.Fatalf("Failed to get transactions: %v", err)
	}

	if len(txs) != 150 {
		t.Errorf("Expected 150 transactions, got %d", len(txs))
	}
}

func TestProcessUnknownActionType(t *testing.T) {
	tokenStore := test_support.SetupTestDB()
	processor := service.NewFractalEngineProcessor(tokenStore)

	// Create transaction with unknown action type (use a valid uint8 value)
	_, err := tokenStore.SaveOnChainTransaction("txUnknown", 1, 1, 99, protocol.DEFAULT_VERSION, []byte{}, "ownerAddress", 0)
	if err != nil {
		t.Fatalf("Failed to save transaction: %v", err)
	}

	// Process should skip unknown action types without error
	err = processor.Process()
	if err != nil {
		t.Fatalf("Expected no error with unknown action type, got: %v", err)
	}
}
