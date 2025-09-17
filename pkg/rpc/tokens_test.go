package rpc_test

import (
	"testing"

	"dogecoin.org/fractal-engine/pkg/rpc"
	"dogecoin.org/fractal-engine/pkg/store"
	"gotest.tools/assert"
)

func TestGetTokenBalance(t *testing.T) {
	tokenisationStore, _, mux, feClient := SetupRpcTest(t)
	rpc.HandleTokenRoutes(tokenisationStore, mux)

	err := tokenisationStore.UpsertTokenBalance("address1", "mint1", 10)
	if err != nil {
		t.Fatalf("Failed to upsert token balance: %v", err)
	}

	tokens, err := feClient.GetTokenBalance("address1", "mint1")
	if err != nil {
		t.Fatalf("Failed to get token balances: %v", err)
	}

	assert.Equal(t, len(tokens), 1)
	assert.Equal(t, tokens[0].Address, "address1")
	assert.Equal(t, tokens[0].MintHash, "mint1")
	assert.Equal(t, tokens[0].Quantity, 10)
}

func TestGetTokenBalanceWithMintDetails(t *testing.T) {
	tokenisationStore, _, mux, feClient := SetupRpcTest(t)
	rpc.HandleTokenRoutes(tokenisationStore, mux)

	_, err := tokenisationStore.SaveMint(&store.MintWithoutID{
		Title:         "mint1",
		Description:   "description1",
		FractionCount: 10,
		Hash:          "mint1",
	}, "address1")
	if err != nil {
		t.Fatalf("Failed to save mint: %v", err)
	}

	err = tokenisationStore.UpsertTokenBalance("address1", "mint1", 10)
	if err != nil {
		t.Fatalf("Failed to upsert token balance: %v", err)
	}

	response, err := feClient.GetTokenBalanceWithMintDetails("address1")
	if err != nil {
		t.Fatalf("Failed to get token balances: %v", err)
	}

	tokens := response.Mints

	assert.Equal(t, len(tokens), 1)
	assert.Equal(t, tokens[0].Address, "address1")
	assert.Equal(t, tokens[0].Quantity, 10)
	assert.Equal(t, tokens[0].Mint.Hash, "mint1")
	assert.Equal(t, tokens[0].Mint.Title, "mint1")
	assert.Equal(t, tokens[0].Mint.Description, "description1")
	assert.Equal(t, tokens[0].Mint.FractionCount, 10)

}
