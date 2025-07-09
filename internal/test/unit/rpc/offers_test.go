package test_rpc

import (
	"testing"

	"dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/rpc"
	"gotest.tools/assert"
)

func TestSellOffers(t *testing.T) {
	tokenisationStore, dogenetClient, mux, feClient := SetupRpcTest(t)
	rpc.HandleOfferRoutes(tokenisationStore, dogenetClient, mux, config.NewConfig())

	offer := rpc.CreateSellOfferRequest{
		OffererAddress: "0x122122121212121",
		MintHash:       "myminthash",
		Quantity:       10,
		Price:          100,
	}

	offerResponse, err := feClient.CreateSellOffer(&offer)
	if err != nil {
		t.Fatalf("Failed to create offer: %v", err)
	}

	offers, err := tokenisationStore.GetSellOffers(0, 10, "myminthash", "0x122122121212121")
	if err != nil {
		t.Fatalf("Failed to get offers: %v", err)
	}

	assert.Equal(t, len(offers), 1)
	assert.Equal(t, offers[0].Id, offerResponse.Id)
	assert.Equal(t, offers[0].OffererAddress, offer.OffererAddress)
	assert.Equal(t, offers[0].MintHash, offer.MintHash)
	assert.Equal(t, offers[0].Quantity, offer.Quantity)
	assert.Equal(t, offers[0].Price, offer.Price)

	assert.Equal(t, len(dogenetClient.sellOffers), 1)
	assert.Equal(t, dogenetClient.sellOffers[0].Id, offerResponse.Id)
	assert.Equal(t, dogenetClient.sellOffers[0].OffererAddress, offer.OffererAddress)
	assert.Equal(t, dogenetClient.sellOffers[0].MintHash, offer.MintHash)
	assert.Equal(t, dogenetClient.sellOffers[0].Quantity, offer.Quantity)
	assert.Equal(t, dogenetClient.sellOffers[0].Price, offer.Price)
}
