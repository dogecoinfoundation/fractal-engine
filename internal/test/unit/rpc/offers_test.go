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
		Payload: rpc.CreateSellOfferRequestPayload{
			OffererAddress: "0x122122121212121",
			MintHash:       "myminthash",
			Quantity:       10,
			Price:          100,
		},
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
	assert.Equal(t, offers[0].OffererAddress, offer.Payload.OffererAddress)
	assert.Equal(t, offers[0].MintHash, offer.Payload.MintHash)
	assert.Equal(t, offers[0].Quantity, offer.Payload.Quantity)
	assert.Equal(t, offers[0].Price, offer.Payload.Price)

	assert.Equal(t, len(dogenetClient.sellOffers), 1)
	assert.Equal(t, dogenetClient.sellOffers[0].Id, offerResponse.Id)
	assert.Equal(t, dogenetClient.sellOffers[0].OffererAddress, offer.Payload.OffererAddress)
	assert.Equal(t, dogenetClient.sellOffers[0].MintHash, offer.Payload.MintHash)
	assert.Equal(t, dogenetClient.sellOffers[0].Quantity, offer.Payload.Quantity)
	assert.Equal(t, dogenetClient.sellOffers[0].Price, offer.Payload.Price)

	deleteOffer := rpc.DeleteSellOfferRequest{
		Payload: rpc.DeleteSellOfferRequestPayload{
			OfferHash: offerResponse.Hash,
		},
	}

	_, err = feClient.DeleteSellOffer(&deleteOffer)
	if err != nil {
		t.Fatalf("Failed to delete offer: %v", err)
	}

	offers, err = tokenisationStore.GetSellOffers(0, 10, "myminthash", "0x122122121212121")
	if err != nil {
		t.Fatalf("Failed to get offers: %v", err)
	}

	assert.Equal(t, len(offers), 0)
	assert.Equal(t, len(dogenetClient.sellOffers), 0)
}

func TestBuyOffers(t *testing.T) {
	tokenisationStore, dogenetClient, mux, feClient := SetupRpcTest(t)
	rpc.HandleOfferRoutes(tokenisationStore, dogenetClient, mux, config.NewConfig())

	offer := rpc.CreateBuyOfferRequest{
		Payload: rpc.CreateBuyOfferRequestPayload{
			OffererAddress: "0x122122121212121",
			SellerAddress:  "0x122122121212121",
			MintHash:       "myminthash",
			Quantity:       10,
			Price:          100,
		},
	}

	offerResponse, err := feClient.CreateBuyOffer(&offer)
	if err != nil {
		t.Fatalf("Failed to create offer: %v", err)
	}

	offers, err := tokenisationStore.GetBuyOffersByMintAndSellerAddress(0, 10, "myminthash", "0x122122121212121")
	if err != nil {
		t.Fatalf("Failed to get offers: %v", err)
	}

	assert.Equal(t, len(offers), 1)
	assert.Equal(t, offers[0].Id, offerResponse.Id)
	assert.Equal(t, offers[0].OffererAddress, offer.Payload.OffererAddress)
	assert.Equal(t, offers[0].SellerAddress, offer.Payload.SellerAddress)
	assert.Equal(t, offers[0].MintHash, offer.Payload.MintHash)
	assert.Equal(t, offers[0].Quantity, offer.Payload.Quantity)

	assert.Equal(t, len(dogenetClient.buyOffers), 1)
	assert.Equal(t, dogenetClient.buyOffers[0].Id, offerResponse.Id)
	assert.Equal(t, dogenetClient.buyOffers[0].OffererAddress, offer.Payload.OffererAddress)
	assert.Equal(t, dogenetClient.buyOffers[0].SellerAddress, offer.Payload.SellerAddress)
	assert.Equal(t, dogenetClient.buyOffers[0].MintHash, offer.Payload.MintHash)
	assert.Equal(t, dogenetClient.buyOffers[0].Quantity, offer.Payload.Quantity)

	deleteOffer := rpc.DeleteBuyOfferRequest{
		Payload: rpc.DeleteBuyOfferRequestPayload{
			OfferHash: offerResponse.Hash,
		},
	}

	_, err = feClient.DeleteBuyOffer(&deleteOffer)
	if err != nil {
		t.Fatalf("Failed to delete offer: %v", err)
	}

	offers, err = tokenisationStore.GetBuyOffersByMintAndSellerAddress(0, 10, "myminthash", "0x122122121212121")
	if err != nil {
		t.Fatalf("Failed to get offers: %v", err)
	}

	assert.Equal(t, len(offers), 0)
	assert.Equal(t, len(dogenetClient.buyOffers), 0)
}
