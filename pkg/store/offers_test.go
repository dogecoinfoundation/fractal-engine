package store

import (
	"testing"
	"time"

	"dogecoin.org/fractal-engine/pkg/config"

	"gotest.tools/assert"
)

func TestOfferSaveAndGet(t *testing.T) {
	tokenisationStore, err := NewTokenisationStore("memory://test.db", config.Config{
		MigrationsPath: "../../db/migrations",
	})
	if err != nil {
		panic(err)
	}

	err = tokenisationStore.Migrate()
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	offer := OfferWithoutID{
		OffererAddress: "test",
		Type:           OfferTypeBuy,
		Hash:           "test",
		MintHash:       "myminthash",
		Quantity:       10,
		Price:          25,
		CreatedAt:      time.Now(),
	}

	id, err := tokenisationStore.SaveOffer(&offer)
	if err != nil {
		t.Fatalf("Failed to save offer: %v", err)
	}

	offers, err := tokenisationStore.GetOffers(0, 10, "myminthashxxxx", int(OfferTypeBuy))
	if err != nil {
		t.Fatalf("Failed to get offer: %v", err)
	}
	assert.Equal(t, len(offers), 0)

	offers, err = tokenisationStore.GetOffers(0, 10, "myminthash", int(OfferTypeSell))
	if err != nil {
		t.Fatalf("Failed to get offer: %v", err)
	}
	assert.Equal(t, len(offers), 0)

	offers, err = tokenisationStore.GetOffers(0, 10, "myminthash", int(OfferTypeBuy))
	if err != nil {
		t.Fatalf("Failed to get offer: %v", err)
	}

	assert.Equal(t, len(offers), 1)
	assert.Equal(t, offers[0].Id, id)
	assert.Equal(t, offers[0].OffererAddress, offer.OffererAddress)
	assert.Equal(t, offers[0].Type, offer.Type)
	assert.Equal(t, offers[0].Hash, offer.Hash)
	assert.Equal(t, offers[0].MintHash, offer.MintHash)
	assert.Equal(t, offers[0].Quantity, offer.Quantity)
	assert.Equal(t, offers[0].Price, offer.Price)
	assert.Equal(t, offers[0].CreatedAt.Unix(), offer.CreatedAt.Unix())
}
