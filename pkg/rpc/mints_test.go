package rpc_test

import (
	"testing"

	"dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/doge"
	"dogecoin.org/fractal-engine/pkg/rpc"
	"dogecoin.org/fractal-engine/pkg/util"
	"gotest.tools/assert"
)

func TestMints(t *testing.T) {
	tokenisationStore, dogenetClient, mux, feClient := SetupRpcTest(t)

	rpc.HandleMintRoutes(tokenisationStore, dogenetClient, mux, &config.Config{}, doge.NewRpcClient(&config.Config{}))

	privHex, pubHex, address, err := doge.GenerateDogecoinKeypair(doge.PrefixRegtest)
	if err != nil {
		t.Fatalf("GenerateDogecoinKeypair: %v", err)
	}

	payload := rpc.CreateMintRequestPayload{
		Title:         "Test Mint",
		FractionCount: 100,
		Description:   "Test Description",
		Tags:          []string{"test", "mint"},
		Metadata: map[string]interface{}{
			"test": "test",
		},
		Requirements:  map[string]interface{}{},
		LockupOptions: map[string]interface{}{},
		FeedURL:       util.StrPtr("https://test.com"),
	}

	mintRequest := rpc.CreateMintRequest{
		Payload:   payload,
		Address:   address,
		PublicKey: pubHex,
	}

	mintRequest.Signature, err = doge.SignPayload(payload, privHex, pubHex)
	if err != nil {
		t.Fatalf("Failed to sign payload: %v", err)
	}

	mintResponse, err := feClient.Mint(&mintRequest)
	if err != nil {
		t.Fatalf("Failed to create mint: %v", err)
	}

	mints, err := tokenisationStore.GetUnconfirmedMints(0, 10)
	if err != nil {
		t.Fatalf("Failed to get mints: %v", err)
	}

	assert.Equal(t, len(mints), 1)
	assert.Equal(t, mints[0].Hash, mintResponse.Hash)
	assert.Equal(t, mints[0].Title, mintRequest.Payload.Title)
	assert.Equal(t, mints[0].FractionCount, mintRequest.Payload.FractionCount)
	assert.Equal(t, mints[0].Description, mintRequest.Payload.Description)
	assert.DeepEqual(t, mints[0].Tags, mintRequest.Payload.Tags)
	assert.DeepEqual(t, mints[0].Metadata, mintRequest.Payload.Metadata)
	assert.DeepEqual(t, mints[0].Requirements, mintRequest.Payload.Requirements)
	assert.DeepEqual(t, mints[0].LockupOptions, mintRequest.Payload.LockupOptions)
	assert.Equal(t, *mints[0].FeedURL, *mintRequest.Payload.FeedURL)

	assert.Equal(t, len(dogenetClient.mints), 1)
	assert.Equal(t, dogenetClient.mints[0].Hash, mintResponse.Hash)
	assert.Equal(t, dogenetClient.mints[0].Title, mintRequest.Payload.Title)
	assert.Equal(t, dogenetClient.mints[0].FractionCount, mintRequest.Payload.FractionCount)
	assert.Equal(t, dogenetClient.mints[0].Description, mintRequest.Payload.Description)
	assert.DeepEqual(t, dogenetClient.mints[0].Tags, mintRequest.Payload.Tags)
	assert.DeepEqual(t, dogenetClient.mints[0].Metadata, mintRequest.Payload.Metadata)
	assert.DeepEqual(t, dogenetClient.mints[0].Requirements, mintRequest.Payload.Requirements)
	assert.DeepEqual(t, dogenetClient.mints[0].LockupOptions, mintRequest.Payload.LockupOptions)
	assert.Equal(t, *dogenetClient.mints[0].FeedURL, *mintRequest.Payload.FeedURL)
}
