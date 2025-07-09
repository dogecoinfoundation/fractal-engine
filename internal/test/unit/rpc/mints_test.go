package test_rpc

import (
	"testing"

	"dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/rpc"
	"gotest.tools/assert"
)

func TestMints(t *testing.T) {
	tokenisationStore, dogenetClient, mux, feClient := SetupRpcTest(t)

	rpc.HandleMintRoutes(tokenisationStore, dogenetClient, mux, &config.Config{})

	offer := rpc.CreateMintRequest{
		Title:         "Test Mint",
		FractionCount: 100,
		Description:   "Test Description",
		Tags:          []string{"test", "mint"},
		Metadata: map[string]interface{}{
			"test": "test",
		},
		Requirements:  map[string]interface{}{},
		LockupOptions: map[string]interface{}{},
		FeedURL:       "https://test.com",
	}

	offerResponse, err := feClient.Mint(&offer)
	if err != nil {
		t.Fatalf("Failed to create mint: %v", err)
	}

	mints, err := tokenisationStore.GetUnconfirmedMints(0, 10)
	if err != nil {
		t.Fatalf("Failed to get mints: %v", err)
	}

	assert.Equal(t, len(mints), 1)
	assert.Equal(t, mints[0].Id, offerResponse.Id)
	assert.Equal(t, mints[0].Title, offer.Title)
	assert.Equal(t, mints[0].FractionCount, offer.FractionCount)
	assert.Equal(t, mints[0].Description, offer.Description)
	assert.DeepEqual(t, mints[0].Tags, offer.Tags)
	assert.DeepEqual(t, mints[0].Metadata, offer.Metadata)
	assert.DeepEqual(t, mints[0].Requirements, offer.Requirements)
	assert.DeepEqual(t, mints[0].LockupOptions, offer.LockupOptions)
	assert.Equal(t, mints[0].FeedURL, offer.FeedURL)

	assert.Equal(t, len(dogenetClient.mints), 1)
	assert.Equal(t, dogenetClient.mints[0].Id, offerResponse.Id)
	assert.Equal(t, dogenetClient.mints[0].Title, offer.Title)
	assert.Equal(t, dogenetClient.mints[0].FractionCount, offer.FractionCount)
	assert.Equal(t, dogenetClient.mints[0].Description, offer.Description)
	assert.DeepEqual(t, dogenetClient.mints[0].Tags, offer.Tags)
	assert.DeepEqual(t, dogenetClient.mints[0].Metadata, offer.Metadata)
	assert.DeepEqual(t, dogenetClient.mints[0].Requirements, offer.Requirements)
	assert.DeepEqual(t, dogenetClient.mints[0].LockupOptions, offer.LockupOptions)
	assert.Equal(t, dogenetClient.mints[0].FeedURL, offer.FeedURL)
}
