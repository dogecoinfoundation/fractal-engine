package service_test

import (
	"testing"
	"time"

	"dogecoin.org/fractal-engine/internal/test/support"
	test_support "dogecoin.org/fractal-engine/internal/test/support"
	"dogecoin.org/fractal-engine/pkg/service"
	"dogecoin.org/fractal-engine/pkg/store"
	"gotest.tools/assert"
)

func TestTrimmerServiceForOnChainTransactions(t *testing.T) {
	tokenisationStore := test_support.SetupTestDB()

	rpcClient := support.NewTestDogeClient(t)

	value := map[string]interface{}{
		"0000000000000000000000000000000000000000000000000000000000000000": 100,
	}

	tokenisationStore.SaveOnChainTransaction("0000000000000000000000000000000000000000000000000000000000000000", 45, "blockHash", 1, 1, 1, []byte{}, "0000000000000000000000000000000000000000000000000000000000000000", value)
	tokenisationStore.SaveOnChainTransaction("0000000000000000000000000000000000000000000000000000000000000000", 30, "blockHash", 1, 1, 1, []byte{}, "0000000000000000000000000000000000000000000000000000000000000000", value)
	tokenisationStore.SaveOnChainTransaction("0000000000000000000000000000000000000000000000000000000000000000", 85, "blockHash", 1, 1, 1, []byte{}, "0000000000000000000000000000000000000000000000000000000000000000", value)
	tokenisationStore.SaveOnChainTransaction("0000000000000000000000000000000000000000000000000000000000000000", 86, "blockHash", 1, 1, 1, []byte{}, "0000000000000000000000000000000000000000000000000000000000000000", value)
	tokenisationStore.SaveOnChainTransaction("0000000000000000000000000000000000000000000000000000000000000000", 87, "blockHash", 1, 1, 1, []byte{}, "0000000000000000000000000000000000000000000000000000000000000000", value)
	tokenisationStore.SaveOnChainTransaction("0000000000000000000000000000000000000000000000000000000000000000", 100, "blockHash", 1, 1, 1, []byte{}, "0000000000000000000000000000000000000000000000000000000000000000", value)

	count, err := tokenisationStore.GetOnChainTransactions(0, 100)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 6, len(count))

	tokenisationStore.SaveUnconfirmedMint(&store.MintWithoutID{
		Hash: "0000000000000000000000000000000000000000000000000000000000000000",
	})
	tokenisationStore.SaveUnconfirmedMint(&store.MintWithoutID{
		Hash: "0000000000000000000000000000000000000000000000000000000000000000",
	})
	tokenisationStore.SaveUnconfirmedMint(&store.MintWithoutID{
		Hash: "0000000000000000000000000000000000000000000000000000000000000000",
	})
	tokenisationStore.SaveUnconfirmedMint(&store.MintWithoutID{
		Hash: "0000000000000000000000000000000000000000000000000000000000000000",
	})

	mintCount, err := tokenisationStore.GetUnconfirmedMints(0, 100)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 4, len(mintCount))

	trimmerService := service.NewTrimmerService(14, 2, tokenisationStore, rpcClient)
	go trimmerService.Start()

	time.Sleep(2 * time.Second)

	count, err = tokenisationStore.GetOnChainTransactions(0, 100)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 3, len(count))

	mintCount, err = tokenisationStore.GetUnconfirmedMints(0, 100)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 2, len(mintCount))

	trimmerService.Stop()
}
