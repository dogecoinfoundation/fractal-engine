package store_test

import (
	"testing"

	"dogecoin.org/fractal-engine/internal/test/support"
	"gotest.tools/assert"
)

func TestSaveOnChainTransaction(t *testing.T) {
	db := support.SetupTestDB()

	txHash := "abc123def456"
	height := int64(12345)
	transactionNumber := 1
	actionType := uint8(1)
	actionVersion := uint8(1)
	actionData := []byte("test action data")
	address := "DTestAddress123"
	value := 100.5

	id, err := db.SaveOnChainTransaction(txHash, height, transactionNumber, actionType, actionVersion, actionData, address, value)
	assert.NilError(t, err)
	assert.Assert(t, id != "")

	// Verify the transaction was saved by querying it
	transactions, err := db.GetOnChainTransactions(0, 10)
	assert.NilError(t, err)
	assert.Equal(t, len(transactions), 1)
	assert.Equal(t, transactions[0].TxHash, txHash)
	assert.Equal(t, transactions[0].Height, height)
	assert.Equal(t, transactions[0].TransactionNumber, transactionNumber)
	assert.Equal(t, transactions[0].ActionType, actionType)
	assert.Equal(t, transactions[0].ActionVersion, actionVersion)
	assert.DeepEqual(t, transactions[0].ActionData, actionData)
	assert.Equal(t, transactions[0].Address, address)
	assert.Equal(t, transactions[0].Value, value)
}

func TestGetOldOnchainTransactions(t *testing.T) {
	db := support.SetupTestDB()

	// Save transactions at different block heights
	_, err := db.SaveOnChainTransaction("tx1", 100, 1, 1, 1, []byte("data1"), "addr1", 10.0)
	assert.NilError(t, err)
	_, err = db.SaveOnChainTransaction("tx2", 200, 1, 1, 1, []byte("data2"), "addr2", 20.0)
	assert.NilError(t, err)
	_, err = db.SaveOnChainTransaction("tx3", 300, 1, 1, 1, []byte("data3"), "addr3", 30.0)
	assert.NilError(t, err)

	// Get transactions older than block height 250
	oldTransactions, err := db.GetOldOnchainTransactions(250)
	assert.NilError(t, err)
	assert.Equal(t, len(oldTransactions), 2)
	assert.Equal(t, oldTransactions[0].TxHash, "tx1")
	assert.Equal(t, oldTransactions[1].TxHash, "tx2")
}

func TestTrimOldOnChainTransactions(t *testing.T) {
	db := support.SetupTestDB()

	// Save transactions at different block heights
	_, err := db.SaveOnChainTransaction("tx1", 100, 1, 1, 1, []byte("data1"), "addr1", 10.0)
	assert.NilError(t, err)
	_, err = db.SaveOnChainTransaction("tx2", 200, 1, 1, 1, []byte("data2"), "addr2", 20.0)
	assert.NilError(t, err)
	_, err = db.SaveOnChainTransaction("tx3", 300, 1, 1, 1, []byte("data3"), "addr3", 30.0)
	assert.NilError(t, err)

	// Trim transactions older than block height 250
	err = db.TrimOldOnChainTransactions(250)
	assert.NilError(t, err)

	// Verify only tx3 remains
	transactions, err := db.GetOnChainTransactions(0, 10)
	assert.NilError(t, err)
	assert.Equal(t, len(transactions), 1)
	assert.Equal(t, transactions[0].TxHash, "tx3")
}

func TestRemoveOnChainTransaction(t *testing.T) {
	db := support.SetupTestDB()

	// Save a transaction
	id, err := db.SaveOnChainTransaction("tx1", 100, 1, 1, 1, []byte("data1"), "addr1", 10.0)
	assert.NilError(t, err)

	// Remove the transaction
	err = db.RemoveOnChainTransaction(id)
	assert.NilError(t, err)

	// Verify it's gone
	transactions, err := db.GetOnChainTransactions(0, 10)
	assert.NilError(t, err)
	assert.Equal(t, len(transactions), 0)
}

func TestCountOnChainTransactions(t *testing.T) {
	db := support.SetupTestDB()

	// Save transactions at different block heights
	_, err := db.SaveOnChainTransaction("tx1", 100, 1, 1, 1, []byte("data1"), "addr1", 10.0)
	assert.NilError(t, err)
	_, err = db.SaveOnChainTransaction("tx2", 100, 2, 1, 1, []byte("data2"), "addr2", 20.0)
	assert.NilError(t, err)
	_, err = db.SaveOnChainTransaction("tx3", 200, 1, 1, 1, []byte("data3"), "addr3", 30.0)
	assert.NilError(t, err)

	// Count transactions at block height 100
	count, err := db.CountOnChainTransactions(100)
	assert.NilError(t, err)
	assert.Equal(t, count, 2)

	// Count transactions at block height 200
	count, err = db.CountOnChainTransactions(200)
	assert.NilError(t, err)
	assert.Equal(t, count, 1)

	// Count transactions at block height with no transactions
	count, err = db.CountOnChainTransactions(300)
	assert.NilError(t, err)
	assert.Equal(t, count, 0)
}

func TestGetOnChainTransactionsPagination(t *testing.T) {
	db := support.SetupTestDB()

	// Save 5 transactions with different heights and transaction numbers
	_, err := db.SaveOnChainTransaction("tx1", 100, 1, 1, 1, []byte("data1"), "addr1", 10.0)
	assert.NilError(t, err)
	_, err = db.SaveOnChainTransaction("tx2", 100, 2, 1, 1, []byte("data2"), "addr2", 20.0)
	assert.NilError(t, err)
	_, err = db.SaveOnChainTransaction("tx3", 200, 1, 1, 1, []byte("data3"), "addr3", 30.0)
	assert.NilError(t, err)
	_, err = db.SaveOnChainTransaction("tx4", 200, 2, 1, 1, []byte("data4"), "addr4", 40.0)
	assert.NilError(t, err)
	_, err = db.SaveOnChainTransaction("tx5", 300, 1, 1, 1, []byte("data5"), "addr5", 50.0)
	assert.NilError(t, err)

	// Test pagination - first page
	transactions, err := db.GetOnChainTransactions(0, 2)
	assert.NilError(t, err)
	assert.Equal(t, len(transactions), 2)
	assert.Equal(t, transactions[0].TxHash, "tx1")
	assert.Equal(t, transactions[1].TxHash, "tx2")

	// Test pagination - second page
	transactions, err = db.GetOnChainTransactions(2, 2)
	assert.NilError(t, err)
	assert.Equal(t, len(transactions), 2)
	assert.Equal(t, transactions[0].TxHash, "tx3")
	assert.Equal(t, transactions[1].TxHash, "tx4")

	// Test pagination - third page
	transactions, err = db.GetOnChainTransactions(4, 2)
	assert.NilError(t, err)
	assert.Equal(t, len(transactions), 1)
	assert.Equal(t, transactions[0].TxHash, "tx5")
}

func TestGetOnChainTransactionsOrdering(t *testing.T) {
	db := support.SetupTestDB()

	// Save transactions out of order
	_, err := db.SaveOnChainTransaction("tx1", 200, 2, 1, 1, []byte("data1"), "addr1", 10.0)
	assert.NilError(t, err)
	_, err = db.SaveOnChainTransaction("tx2", 100, 2, 1, 1, []byte("data2"), "addr2", 20.0)
	assert.NilError(t, err)
	_, err = db.SaveOnChainTransaction("tx3", 200, 1, 1, 1, []byte("data3"), "addr3", 30.0)
	assert.NilError(t, err)
	_, err = db.SaveOnChainTransaction("tx4", 100, 1, 1, 1, []byte("data4"), "addr4", 40.0)
	assert.NilError(t, err)

	// Get all transactions and verify ordering
	transactions, err := db.GetOnChainTransactions(0, 10)
	assert.NilError(t, err)
	assert.Equal(t, len(transactions), 4)
	
	// Should be ordered by block height first, then transaction number
	assert.Equal(t, transactions[0].TxHash, "tx4") // height 100, tx_num 1
	assert.Equal(t, transactions[1].TxHash, "tx2") // height 100, tx_num 2
	assert.Equal(t, transactions[2].TxHash, "tx3") // height 200, tx_num 1
	assert.Equal(t, transactions[3].TxHash, "tx1") // height 200, tx_num 2
}

func TestOnChainTransactionEdgeCases(t *testing.T) {
	db := support.SetupTestDB()

	// Test with empty action data
	id, err := db.SaveOnChainTransaction("tx1", 100, 1, 1, 1, []byte{}, "addr1", 0.0)
	assert.NilError(t, err)
	assert.Assert(t, id != "")

	// Test with zero value
	id2, err := db.SaveOnChainTransaction("tx2", 100, 2, 0, 0, []byte("data"), "addr2", 0.0)
	assert.NilError(t, err)
	assert.Assert(t, id2 != "")

	// Verify both were saved correctly
	transactions, err := db.GetOnChainTransactions(0, 10)
	assert.NilError(t, err)
	assert.Equal(t, len(transactions), 2)
	assert.Equal(t, len(transactions[0].ActionData), 0)
	assert.Equal(t, transactions[0].Value, 0.0)
	assert.Equal(t, transactions[1].ActionType, uint8(0))
	assert.Equal(t, transactions[1].ActionVersion, uint8(0))
}

func TestRemoveNonExistentTransaction(t *testing.T) {
	db := support.SetupTestDB()

	// Try to remove a non-existent transaction
	err := db.RemoveOnChainTransaction("non-existent-id")
	assert.NilError(t, err) // Should not error, just no rows affected
}

func TestGetOnChainTransactionsEmptyResult(t *testing.T) {
	db := support.SetupTestDB()

	// Get transactions when database is empty
	transactions, err := db.GetOnChainTransactions(0, 10)
	assert.NilError(t, err)
	assert.Equal(t, len(transactions), 0)
}