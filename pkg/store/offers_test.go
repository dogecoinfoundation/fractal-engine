package store_test

import (
	"testing"
	"time"

	"dogecoin.org/fractal-engine/internal/test/support"
	"dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/store"

	"gotest.tools/assert"
)

func TestOfferSaveAndGet(t *testing.T) {
	tokenisationStore, err := store.NewTokenisationStore("memory://test.db", config.Config{
		MigrationsPath: "../../db/migrations",
	})
	if err != nil {
		panic(err)
	}

	err = tokenisationStore.Migrate()
	if err != nil && err.Error() != "no change" {
		t.Fatalf("Failed to migrate: %v", err)
	}

	offererAddress := support.GenerateDogecoinAddress(true)
	sellerAddress := support.GenerateDogecoinAddress(true)

	offer := store.BuyOfferWithoutID{
		OffererAddress: offererAddress,
		SellerAddress:  sellerAddress,
		MintHash:       "myminthash",
		Quantity:       10,
		Price:          25,
		CreatedAt:      time.Now(),
	}

	id, err := tokenisationStore.SaveBuyOffer(&offer)
	if err != nil {
		t.Fatalf("Failed to save offer: %v", err)
	}

	offers, err := tokenisationStore.GetBuyOffersByMintAndSellerAddress(0, 10, "myminthashxxxx", sellerAddress)
	if err != nil {
		t.Fatalf("Failed to get offer: %v", err)
	}
	assert.Equal(t, len(offers), 0)

	offers, err = tokenisationStore.GetBuyOffersByMintAndSellerAddress(0, 10, "myminthashxxxx", "")
	if err != nil {
		t.Fatalf("Failed to get offer: %v", err)
	}
	assert.Equal(t, len(offers), 0)

	offers, err = tokenisationStore.GetBuyOffersByMintAndSellerAddress(0, 10, "myminthash", "mySellerAddressxxx")
	if err != nil {
		t.Fatalf("Failed to get offer: %v", err)
	}
	assert.Equal(t, len(offers), 0)

	offers, err = tokenisationStore.GetBuyOffersByMintAndSellerAddress(0, 10, "myminthash", sellerAddress)
	if err != nil {
		t.Fatalf("Failed to get offer: %v", err)
	}

	assert.Equal(t, len(offers), 1)
	assert.Equal(t, offers[0].Id, id)
	assert.Equal(t, offers[0].OffererAddress, offer.OffererAddress)
	assert.Equal(t, offers[0].Hash, offer.Hash)
	assert.Equal(t, offers[0].MintHash, offer.MintHash)
	assert.Equal(t, offers[0].Quantity, offer.Quantity)
	assert.Equal(t, offers[0].Price, offer.Price)
	assert.Equal(t, offers[0].CreatedAt.Unix(), offer.CreatedAt.Unix())

	offers, err = tokenisationStore.GetBuyOffersByMintAndSellerAddress(0, 10, "myminthash", "")
	if err != nil {
		t.Fatalf("Failed to get offer: %v", err)
	}

	assert.Equal(t, len(offers), 1)
	assert.Equal(t, offers[0].Id, id)
	assert.Equal(t, offers[0].OffererAddress, offer.OffererAddress)
	assert.Equal(t, offers[0].Hash, offer.Hash)
	assert.Equal(t, offers[0].MintHash, offer.MintHash)
	assert.Equal(t, offers[0].Quantity, offer.Quantity)
	assert.Equal(t, offers[0].Price, offer.Price)
	assert.Equal(t, offers[0].CreatedAt.Unix(), offer.CreatedAt.Unix())
}
