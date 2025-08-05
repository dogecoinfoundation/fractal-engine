package store_test

import (
	"database/sql"
	"testing"
	"time"

	test_support "dogecoin.org/fractal-engine/internal/test/support"
	"dogecoin.org/fractal-engine/pkg/protocol"
	"dogecoin.org/fractal-engine/pkg/store"
	"google.golang.org/protobuf/proto"
	"gotest.tools/assert"
)

// Helper function to find a transaction by ID
func findTransactionById(txs []store.OnChainTransaction, id string) *store.OnChainTransaction {
	for i := range txs {
		if txs[i].Id == id {
			return &txs[i]
		}
	}
	return nil
}

func TestMatchPaymentSuccess(t *testing.T) {
	tokenStore := test_support.SetupTestDB()

	// This test demonstrates the full payment flow:
	// 1. Create mint and match it to establish token balance
	// 2. Create invoice and match it to establish pending token balance  
	// 3. Create payment and match it to transfer tokens

	mintHash := "testMint123"
	invoiceHash := "testInvoice123"
	sellerAddress := test_support.GenerateDogecoinAddress(true)
	buyerAddress := test_support.GenerateDogecoinAddress(true)
	quantity := 50
	value := 50.0

	// Step 1: Create and match mint
	_, err := tokenStore.SaveUnconfirmedMint(&store.MintWithoutID{
		Hash:          mintHash,
		Title:         "Test Mint",
		Description:   "Test Description",
		FractionCount: 100,
		BlockHeight:   1,
		TransactionHash: sql.NullString{
			String: "mintTx",
			Valid:  true,
		},
	})
	assert.NilError(t, err)

	mintMsg := &protocol.OnChainMintMessage{Hash: mintHash}
	encodedMintMsg, _ := proto.Marshal(mintMsg)
	mintTxId, err := tokenStore.SaveOnChainTransaction("mintTx", 1, 1, protocol.ACTION_MINT, protocol.DEFAULT_VERSION, encodedMintMsg, sellerAddress, 100)
	assert.NilError(t, err)

	txs, err := tokenStore.GetOnChainTransactions(0, 10)
	assert.NilError(t, err)
	mintTx := findTransactionById(txs, mintTxId)
	assert.Assert(t, mintTx != nil)

	err = tokenStore.MatchUnconfirmedMint(*mintTx)
	assert.NilError(t, err)

	// Step 2: Create and match invoice
	_, err = tokenStore.SaveUnconfirmedInvoice(&store.UnconfirmedInvoice{
		Hash:                   invoiceHash,
		PaymentAddress:         sellerAddress,
		BuyOfferOffererAddress: buyerAddress,
		BuyOfferHash:           "buyOffer123",
		BuyOfferMintHash:       mintHash,
		BuyOfferQuantity:       quantity,
		BuyOfferPrice:          100,
		BuyOfferValue:          value,
		CreatedAt:              time.Now(),
		SellOfferAddress:       sellerAddress,
	})
	assert.NilError(t, err)

	invoiceMsg := &protocol.OnChainInvoiceMessage{
		SellOfferAddress: sellerAddress,
		InvoiceHash:      invoiceHash,
		MintHash:         mintHash,
		Quantity:         int32(quantity),
	}
	encodedInvoiceMsg, _ := proto.Marshal(invoiceMsg)
	invoiceTxId, err := tokenStore.SaveOnChainTransaction("invoiceTx", 2, 1, protocol.ACTION_INVOICE, protocol.DEFAULT_VERSION, encodedInvoiceMsg, sellerAddress, float64(quantity))
	assert.NilError(t, err)

	txs, err = tokenStore.GetOnChainTransactions(0, 10)
	assert.NilError(t, err)
	invoiceTx := findTransactionById(txs, invoiceTxId)
	assert.Assert(t, invoiceTx != nil)

	// The invoice processing normally creates pending token balance
	// For this test, we need to create it manually
	err = tokenStore.UpsertPendingTokenBalance(invoiceHash, mintHash, quantity, invoiceTx.Id, sellerAddress)
	assert.NilError(t, err)

	err = tokenStore.MatchUnconfirmedInvoice(*invoiceTx)
	assert.NilError(t, err)

	// Step 3: Create and match payment
	paymentMsg := &protocol.OnChainPaymentMessage{
		Hash: invoiceHash,
	}
	encodedPaymentMsg, _ := proto.Marshal(paymentMsg)
	paymentTxId, err := tokenStore.SaveOnChainTransaction("paymentTx", 3, 1, protocol.ACTION_PAYMENT, protocol.DEFAULT_VERSION, encodedPaymentMsg, buyerAddress, value)
	assert.NilError(t, err)

	txs, err = tokenStore.GetOnChainTransactions(0, 10)
	assert.NilError(t, err)
	paymentTx := findTransactionById(txs, paymentTxId)
	assert.Assert(t, paymentTx != nil)

	// Test MatchPayment
	err = tokenStore.MatchPayment(*paymentTx)
	assert.NilError(t, err)

	// Verify results
	// Check invoice is paid
	tx, _ := tokenStore.DB.Begin()
	defer tx.Rollback()

	var paidAt sql.NullTime
	err = tx.QueryRow("SELECT paid_at FROM invoices WHERE hash = $1", invoiceHash).Scan(&paidAt)
	assert.NilError(t, err)
	assert.Assert(t, paidAt.Valid, "Invoice should be marked as paid")

	// Check payment transaction was deleted
	var count int
	err = tx.QueryRow("SELECT COUNT(*) FROM onchain_transactions WHERE id = $1", paymentTx.Id).Scan(&count)
	assert.NilError(t, err)
	assert.Equal(t, 0, count, "Payment transaction should be deleted")

	// Check pending balance was removed
	_, err = tokenStore.GetPendingTokenBalance(invoiceHash, mintHash, tx)
	assert.Assert(t, err != nil, "Pending balance should be removed")

	// Check buyer received tokens
	buyerBalances, err := tokenStore.GetTokenBalances(buyerAddress, mintHash)
	assert.NilError(t, err)
	assert.Equal(t, 1, len(buyerBalances))
	assert.Equal(t, quantity, buyerBalances[0].Quantity)

	// Check seller's balance was reduced
	sellerBalances, err := tokenStore.GetTokenBalances(sellerAddress, mintHash)
	assert.NilError(t, err)
	totalSellerBalance := 0
	for _, balance := range sellerBalances {
		totalSellerBalance += balance.Quantity
	}
	assert.Equal(t, 50, totalSellerBalance) // 100 - 50
}

func TestMatchPaymentWrongActionType(t *testing.T) {
	tokenStore := test_support.SetupTestDB()

	paymentTx := store.OnChainTransaction{
		Id:         "payment123",
		TxHash:     "paymentTx",
		ActionType: protocol.ACTION_MINT, // Wrong type
	}

	err := tokenStore.MatchPayment(paymentTx)
	assert.ErrorContains(t, err, "action type is not payment")
}

func TestMatchPaymentInvalidProtobuf(t *testing.T) {
	tokenStore := test_support.SetupTestDB()

	paymentTx := store.OnChainTransaction{
		Id:         "payment123",
		TxHash:     "paymentTx",
		ActionType: protocol.ACTION_PAYMENT,
		ActionData: []byte("invalid protobuf data"),
	}

	err := tokenStore.MatchPayment(paymentTx)
	assert.Assert(t, err != nil, "Should fail with invalid protobuf")
}

func TestMatchPaymentInvoiceNotFound(t *testing.T) {
	tokenStore := test_support.SetupTestDB()

	buyerAddress := test_support.GenerateDogecoinAddress(true)

	paymentMsg := &protocol.OnChainPaymentMessage{
		Hash: "nonexistentInvoice",
	}
	encodedPaymentMsg, _ := proto.Marshal(paymentMsg)

	paymentTxId, err := tokenStore.SaveOnChainTransaction("paymentTx", 1, 1, protocol.ACTION_PAYMENT, protocol.DEFAULT_VERSION, encodedPaymentMsg, buyerAddress, 50.0)
	assert.NilError(t, err)

	txs, err := tokenStore.GetOnChainTransactions(0, 10)
	assert.NilError(t, err)
	paymentTx := findTransactionById(txs, paymentTxId)
	assert.Assert(t, paymentTx != nil)

	err = tokenStore.MatchPayment(*paymentTx)
	assert.ErrorContains(t, err, "invoice not found")
}

func TestMatchPaymentValueMismatch(t *testing.T) {
	tokenStore := test_support.SetupTestDB()

	invoiceHash := "testInvoice123"
	buyerAddress := test_support.GenerateDogecoinAddress(true)
	sellerAddress := test_support.GenerateDogecoinAddress(true)
	
	// Create invoice with specific value
	actualInvoice := &store.Invoice{
		Hash:                   invoiceHash,
		PaymentAddress:         sellerAddress,
		BuyOfferOffererAddress: buyerAddress,
		BuyOfferHash:           "buyOffer123",
		BuyOfferMintHash:       "mint123",
		BuyOfferQuantity:       50,
		BuyOfferPrice:          100,
		BuyOfferValue:          50.0, // Expected value
		CreatedAt:              time.Now(),
		SellOfferAddress:       sellerAddress,
	}
	_, err := tokenStore.SaveInvoice(actualInvoice)
	assert.NilError(t, err)

	// Create payment with wrong value
	paymentMsg := &protocol.OnChainPaymentMessage{
		Hash: invoiceHash,
	}
	encodedPaymentMsg, _ := proto.Marshal(paymentMsg)

	paymentTxId, err := tokenStore.SaveOnChainTransaction("paymentTx", 1, 1, protocol.ACTION_PAYMENT, protocol.DEFAULT_VERSION, encodedPaymentMsg, buyerAddress, 25.0) // Wrong value
	assert.NilError(t, err)

	txs, err := tokenStore.GetOnChainTransactions(0, 10)
	assert.NilError(t, err)
	paymentTx := findTransactionById(txs, paymentTxId)
	assert.Assert(t, paymentTx != nil)

	err = tokenStore.MatchPayment(*paymentTx)
	assert.ErrorContains(t, err, "payment value is not equal to buy offer value")
}

func TestMatchPaymentPendingBalanceMismatch(t *testing.T) {
	tokenStore := test_support.SetupTestDB()

	invoiceHash := "testInvoice123"
	mintHash := "mint123"
	buyerAddress := test_support.GenerateDogecoinAddress(true)
	sellerAddress := test_support.GenerateDogecoinAddress(true)

	// Create invoice
	actualInvoice := &store.Invoice{
		Hash:                   invoiceHash,
		PaymentAddress:         sellerAddress,
		BuyOfferOffererAddress: buyerAddress,
		BuyOfferHash:           "buyOffer123",
		BuyOfferMintHash:       mintHash,
		BuyOfferQuantity:       50, // Expected quantity
		BuyOfferPrice:          100,
		BuyOfferValue:          50.0,
		CreatedAt:              time.Now(),
		SellOfferAddress:       sellerAddress,
	}
	_, err := tokenStore.SaveInvoice(actualInvoice)
	assert.NilError(t, err)

	// Create pending balance with wrong quantity
	err = tokenStore.UpsertPendingTokenBalance(invoiceHash, mintHash, 30, "invoiceTxId", sellerAddress) // Wrong quantity
	assert.NilError(t, err)

	// Create payment
	paymentMsg := &protocol.OnChainPaymentMessage{
		Hash: invoiceHash,
	}
	encodedPaymentMsg, _ := proto.Marshal(paymentMsg)

	paymentTxId, err := tokenStore.SaveOnChainTransaction("paymentTx", 1, 1, protocol.ACTION_PAYMENT, protocol.DEFAULT_VERSION, encodedPaymentMsg, buyerAddress, 50.0)
	assert.NilError(t, err)

	txs, err := tokenStore.GetOnChainTransactions(0, 10)
	assert.NilError(t, err)
	paymentTx := findTransactionById(txs, paymentTxId)
	assert.Assert(t, paymentTx != nil)

	err = tokenStore.MatchPayment(*paymentTx)
	assert.ErrorContains(t, err, "pending token balance quantity is not equal to buy offer quantity")
}

func TestMatchPaymentNoPendingBalance(t *testing.T) {
	tokenStore := test_support.SetupTestDB()

	invoiceHash := "testInvoice123"
	buyerAddress := test_support.GenerateDogecoinAddress(true)
	sellerAddress := test_support.GenerateDogecoinAddress(true)

	// Create invoice without pending balance
	actualInvoice := &store.Invoice{
		Hash:                   invoiceHash,
		PaymentAddress:         sellerAddress,
		BuyOfferOffererAddress: buyerAddress,
		BuyOfferHash:           "buyOffer123",
		BuyOfferMintHash:       "mint123",
		BuyOfferQuantity:       50,
		BuyOfferPrice:          100,
		BuyOfferValue:          50.0,
		CreatedAt:              time.Now(),
		SellOfferAddress:       sellerAddress,
	}
	_, err := tokenStore.SaveInvoice(actualInvoice)
	assert.NilError(t, err)

	// Create payment
	paymentMsg := &protocol.OnChainPaymentMessage{
		Hash: invoiceHash,
	}
	encodedPaymentMsg, _ := proto.Marshal(paymentMsg)

	paymentTxId, err := tokenStore.SaveOnChainTransaction("paymentTx", 1, 1, protocol.ACTION_PAYMENT, protocol.DEFAULT_VERSION, encodedPaymentMsg, buyerAddress, 50.0)
	assert.NilError(t, err)

	txs, err := tokenStore.GetOnChainTransactions(0, 10)
	assert.NilError(t, err)
	paymentTx := findTransactionById(txs, paymentTxId)
	assert.Assert(t, paymentTx != nil)

	// Should fail due to missing pending balance
	err = tokenStore.MatchPayment(*paymentTx)
	assert.Assert(t, err != nil, "Should fail without pending balance")

	// Verify transaction was rolled back - invoice should NOT be paid
	tx, _ := tokenStore.DB.Begin()
	defer tx.Rollback()

	var paidAt sql.NullTime
	err = tx.QueryRow("SELECT paid_at FROM invoices WHERE hash = $1", invoiceHash).Scan(&paidAt)
	assert.NilError(t, err)
	assert.Assert(t, !paidAt.Valid, "Invoice should NOT be paid due to rollback")
}