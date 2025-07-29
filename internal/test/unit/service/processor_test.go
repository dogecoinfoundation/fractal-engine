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
	tokenisationStore := test_support.SetupTestDB()

	hash := CreateUnconfirmedMint(t, "txHash001", tokenisationStore)

	processor := service.NewFractalEngineProcessor(tokenisationStore)
	processor.Process()

	AssertUnconfirmedMintCreation(t, hash, tokenisationStore)
}

func TestInvoiceMatch(t *testing.T) {
	tokenisationStore := test_support.SetupTestDB()

	hash := CreateUnconfirmedMint(t, "txHash001", tokenisationStore)

	processor := service.NewFractalEngineProcessor(tokenisationStore)
	processor.Process()

	AssertUnconfirmedMintCreation(t, hash, tokenisationStore)

	CreateOnChainInvoiceMessage(t, "txHash002", 1, 1, "ownerAddress", "ownerAddress", "invoiceHash", hash, 50, tokenisationStore)
	processor.Process()

	CreateOnChainInvoiceMessage(t, "txHash003", 1, 1, "ownerAddress", "ownerAddress", "invoiceHash2", hash, 90, tokenisationStore)
	processor.Process()

	SaveUnconfirmedInvoice(t, "ownerAddress", "buyerAddress", "invoiceHash", hash, 50, 100, tokenisationStore)
	processor.Process()

	AssertTokenBalance(t, "ownerAddress", hash, 100, tokenisationStore)
	AssertPendingTokenBalance(t, "invoiceHash", hash, 50, tokenisationStore)

	CreateOnChainPaymentMessage(t, "txHash004", "invoiceHash", "ownerAddress", 1, 1, 100, tokenisationStore)
	processor.Process()

	AssertTokenBalance(t, "buyerAddress", hash, 50, tokenisationStore)
	AssertTokenBalance(t, "ownerAddress", hash, 50, tokenisationStore)
}

func TestInvoiceMatchEarlierBlockHeightAndTransactionNumber(t *testing.T) {
	tokenisationStore := test_support.SetupTestDB()

	hash := CreateUnconfirmedMint(t, "txHash001", tokenisationStore)

	processor := service.NewFractalEngineProcessor(tokenisationStore)
	processor.Process()

	AssertUnconfirmedMintCreation(t, hash, tokenisationStore)

	CreateOnChainInvoiceMessage(t, "txHash002", 3, 1, "ownerAddress", "ownerAddress", "invoiceHash", hash, 66, tokenisationStore)
	CreateOnChainInvoiceMessage(t, "txHash003", 1, 2, "ownerAddress", "ownerAddress", "invoiceHash2", hash, 77, tokenisationStore)
	CreateOnChainInvoiceMessage(t, "txHash004", 1, 1, "ownerAddress", "ownerAddress", "invoiceHash3", hash, 88, tokenisationStore)

	processor.Process()

	AssertPendingTokenBalance(t, "invoiceHash3", hash, 88, tokenisationStore)
}

func TestPaymentIsLessThanExpected(t *testing.T) {
	tokenisationStore := test_support.SetupTestDB()

	hash := CreateUnconfirmedMint(t, "txHash001", tokenisationStore)

	processor := service.NewFractalEngineProcessor(tokenisationStore)
	processor.Process()

	AssertUnconfirmedMintCreation(t, hash, tokenisationStore)

	CreateOnChainInvoiceMessage(t, "txHash002", 3, 1, "ownerAddress", "ownerAddress", "invoiceHash", hash, 50, tokenisationStore)
	SaveUnconfirmedInvoice(t, "ownerAddress", "buyerAddress", "invoiceHash", hash, 50, 75, tokenisationStore)

	processor.Process()

	AssertPendingTokenBalance(t, "invoiceHash", hash, 50, tokenisationStore)

	CreateOnChainPaymentMessage(t, "txHash003", "invoiceHash", "ownerAddress", 1, 1, 49, tokenisationStore)
	processor.Process()

	AssertTokenBalance(t, "buyerAddress", hash, 0, tokenisationStore)
	AssertTokenBalance(t, "ownerAddress", hash, 100, tokenisationStore)
}

func TestInvoiceTimesOutAfter14BlockDays(t *testing.T) {
	tokenisationStore := test_support.SetupTestDB()

	hash := CreateUnconfirmedMint(t, "txHash001", tokenisationStore)

	invoiceTimeoutProcessor := service.NewInvoiceTimeoutProcessor(tokenisationStore)
	processor := service.NewFractalEngineProcessor(tokenisationStore)
	processor.Process()

	AssertUnconfirmedMintCreation(t, hash, tokenisationStore)

	CreateOnChainInvoiceMessage(t, "txHash002", 3, 1, "ownerAddress", "ownerAddress", "invoiceHash", hash, 50, tokenisationStore)
	processor.Process()

	AssertPendingTokenBalance(t, "invoiceHash", hash, 50, tokenisationStore)

	invoiceTimeoutProcessor.Process(4)

	AssertNoPendingTokenBalance(t, "invoiceHash", hash, tokenisationStore)
}

func TestInvoiceCreationNotFromSeller(t *testing.T) {
	tokenisationStore := test_support.SetupTestDB()

	hash := CreateUnconfirmedMint(t, "txHash001", tokenisationStore)

	processor := service.NewFractalEngineProcessor(tokenisationStore)
	processor.Process()

	AssertUnconfirmedMintCreation(t, hash, tokenisationStore)

	CreateOnChainInvoiceMessage(t, "txHash002", 3, 1, "ownerAddress", "NotOwnerAddress", "invoiceHash", hash, 50, tokenisationStore)
	processor.Process()

	AssertNoPendingTokenBalance(t, "invoiceHash", hash, tokenisationStore)
}

func AssertNoPendingTokenBalance(t *testing.T, invoiceHash string, mintHash string, tokenisationStore *store.TokenisationStore) {
	tx, err := tokenisationStore.DB.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	_, err = tokenisationStore.GetPendingTokenBalance(invoiceHash, mintHash, tx)
	if err == nil {
		t.Fatalf("Expected error getting pending token balance")
	}
}

func AssertPendingTokenBalance(t *testing.T, invoiceHash string, mintHash string, quantity int, tokenisationStore *store.TokenisationStore) {
	tx, err := tokenisationStore.DB.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	pendingTokenBalance, err := tokenisationStore.GetPendingTokenBalance(invoiceHash, mintHash, tx)
	if err != nil {
		t.Fatalf("Failed to get pending token balance: %v", err)
	}

	assert.Equal(t, quantity, pendingTokenBalance.Quantity)
}

func AssertTokenBalance(t *testing.T, s, hash string, i int, tokenisationStore *store.TokenisationStore) {
	tokenBalance, err := tokenisationStore.GetTokenBalances(s, hash)
	if err != nil {
		t.Fatalf("Failed to get token balance: %v", err)
	}

	totalQuantity := 0
	for _, balance := range tokenBalance {
		totalQuantity += balance.Quantity
	}

	assert.Equal(t, i, totalQuantity)
}

func CreateOnChainPaymentMessage(t *testing.T, trxnHash string, invoiceHash string, ownerAddress string, blockHeight int64, trxnNo int, value float64, tokenisationStore *store.TokenisationStore) {
	message3 := protocol.OnChainPaymentMessage{
		Hash: invoiceHash,
	}

	encodedMessage3, err := proto.Marshal(&message3)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}

	_, err = tokenisationStore.SaveOnChainTransaction(trxnHash, blockHeight, trxnNo, protocol.ACTION_PAYMENT, protocol.DEFAULT_VERSION, encodedMessage3, ownerAddress, value)
	if err != nil {
		t.Fatalf("Failed to save on chain transaction: %v", err)
	}
}

func SaveUnconfirmedInvoice(t *testing.T, ownerAddress string, buyerAddress string, invoiceHash string, mintHash string, quantity int, totalValue int, tokenisationStore *store.TokenisationStore) {
	_, err := tokenisationStore.SaveUnconfirmedInvoice(&store.UnconfirmedInvoice{
		Hash:                   invoiceHash,
		PaymentAddress:         ownerAddress,
		BuyOfferOffererAddress: buyerAddress,
		BuyOfferHash:           "buyOfferHash",
		BuyOfferMintHash:       mintHash,
		BuyOfferQuantity:       quantity,
		BuyOfferPrice:          100,
		BuyOfferValue:          float64(totalValue),
		CreatedAt:              time.Now(),
		SellOfferAddress:       "ownerAddress",
	})

	if err != nil {
		t.Fatalf("Failed to save invoice: %v", err)
	}

}

func CreateOnChainInvoiceMessage(t *testing.T, trxnHash string, blockHeight int64, trxnNo int, ownerAddress string, sellOfferAddress string, invoiceHash string, mintHash string, quantity int, tokenisationStore *store.TokenisationStore) {
	message := protocol.OnChainInvoiceMessage{
		SellOfferAddress: sellOfferAddress,
		InvoiceHash:      invoiceHash,
		MintHash:         mintHash,
		Quantity:         int32(quantity),
	}

	encodedMessage, err := proto.Marshal(&message)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}

	_, err = tokenisationStore.SaveOnChainTransaction(trxnHash, blockHeight, trxnNo, protocol.ACTION_INVOICE, protocol.DEFAULT_VERSION, encodedMessage, ownerAddress, float64(quantity))
	if err != nil {
		t.Fatalf("Failed to save on chain transaction: %v", err)
	}
}

func AssertUnconfirmedMintCreation(t *testing.T, hash string, tokenisationStore *store.TokenisationStore) {
	mints, err := tokenisationStore.GetMints(0, 100)
	if err != nil {
		t.Fatalf("Failed to get mints: %v", err)
	}

	assert.Equal(t, 1, len(mints))
	assert.Equal(t, "Test Mint", mints[0].Title)

	tokenBalance, err := tokenisationStore.GetTokenBalances("ownerAddress", hash)
	if err != nil {
		t.Fatalf("Failed to get token balance: %v", err)
	}

	assert.Equal(t, 100, tokenBalance[0].Quantity)
}

func CreateUnconfirmedMint(t *testing.T, trxnHash string, tokenisationStore *store.TokenisationStore) string {
	mintWithoutId := store.MintWithoutID{
		Title:         "Test Mint",
		Description:   "Test Description",
		Tags:          []string{"test", "test2"},
		FeedURL:       "https://test.com",
		FractionCount: 100,
		BlockHeight:   20,
		TransactionHash: sql.NullString{
			String: trxnHash,
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

	_, err = tokenisationStore.SaveOnChainTransaction(trxnHash, 1, 1, protocol.ACTION_MINT, protocol.DEFAULT_VERSION, encodedMessage, "ownerAddress", 100)
	if err != nil {
		t.Fatalf("Failed to save on chain transaction: %v", err)
	}

	return hash
}
