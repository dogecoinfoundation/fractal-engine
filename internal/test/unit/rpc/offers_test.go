package test_rpc

import (
	"testing"

	"dogecoin.org/fractal-engine/pkg/rpc"
	"dogecoin.org/fractal-engine/pkg/store"
	"gotest.tools/assert"
)

func TestOffers(t *testing.T) {
	tokenisationStore, dogenetClient, mux, feClient := SetupRpcTest(t)
	rpc.HandleOfferRoutes(tokenisationStore, dogenetClient, mux)

	offer := rpc.CreateOfferRequest{
		OffererAddress: "0x122122121212121",
		Type:           store.OfferTypeSell,
		MintHash:       "myminthash",
		Quantity:       10,
		Price:          100,
	}

	offerResponse, err := feClient.Offer(&offer)
	if err != nil {
		t.Fatalf("Failed to create offer: %v", err)
	}

	offers, err := tokenisationStore.GetOffers(0, 10, "myminthash", int(store.OfferTypeSell))
	if err != nil {
		t.Fatalf("Failed to get offers: %v", err)
	}

	assert.Equal(t, len(offers), 1)
	assert.Equal(t, offers[0].Id, offerResponse.Id)
	assert.Equal(t, offers[0].Type, offer.Type)
	assert.Equal(t, offers[0].OffererAddress, offer.OffererAddress)
	assert.Equal(t, offers[0].MintHash, offer.MintHash)
	assert.Equal(t, offers[0].Quantity, offer.Quantity)
	assert.Equal(t, offers[0].Price, offer.Price)

	assert.Equal(t, len(dogenetClient.offers), 1)
	assert.Equal(t, dogenetClient.offers[0].Id, offerResponse.Id)
	assert.Equal(t, dogenetClient.offers[0].Type, offer.Type)
	assert.Equal(t, dogenetClient.offers[0].OffererAddress, offer.OffererAddress)
	assert.Equal(t, dogenetClient.offers[0].MintHash, offer.MintHash)
	assert.Equal(t, dogenetClient.offers[0].Quantity, offer.Quantity)
	assert.Equal(t, dogenetClient.offers[0].Price, offer.Price)
}
