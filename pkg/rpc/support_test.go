package rpc_test

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	test_support "dogecoin.org/fractal-engine/internal/test/support"
	"dogecoin.org/fractal-engine/pkg/client"
	"dogecoin.org/fractal-engine/pkg/doge"
	"dogecoin.org/fractal-engine/pkg/dogenet"
	"dogecoin.org/fractal-engine/pkg/store"
)

type FakeGossipClient struct {
	dogenet.GossipClient
	buyOffers  []store.BuyOffer
	sellOffers []store.SellOffer
	mints      []store.Mint
	invoices   []store.UnconfirmedInvoice
}

func (g *FakeGossipClient) GossipBuyOffer(offer store.BuyOffer) error {
	g.buyOffers = append(g.buyOffers, offer)
	return nil
}

func (g *FakeGossipClient) GossipSellOffer(offer store.SellOffer) error {
	g.sellOffers = append(g.sellOffers, offer)
	return nil
}

func (g *FakeGossipClient) GossipMint(mint store.Mint) error {
	g.mints = append(g.mints, mint)
	return nil
}

func (g *FakeGossipClient) GossipUnconfirmedInvoice(invoice store.UnconfirmedInvoice) error {
	g.invoices = append(g.invoices, invoice)
	return nil
}

func (g *FakeGossipClient) GossipDeleteBuyOffer(hash string, publicKey string, signature string) error {
	for i, offer := range g.buyOffers {
		if offer.Hash == hash {
			log.Printf("Deleted buy offer: %v", offer)
			g.buyOffers = append(g.buyOffers[:i], g.buyOffers[i+1:]...)
			return nil
		}
	}

	return nil
}

func (g *FakeGossipClient) GossipDeleteSellOffer(hash string, publicKey string, signature string) error {
	for i, offer := range g.sellOffers {
		if offer.Hash == hash {
			log.Printf("Deleted sell offer: %v", offer)
			g.sellOffers = append(g.sellOffers[:i], g.sellOffers[i+1:]...)
			return nil
		}
	}

	return nil
}

func SetupRpcTest(t *testing.T) (*store.TokenisationStore, *FakeGossipClient, *http.ServeMux, *client.TokenisationClient) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	dogenetClient := &FakeGossipClient{
		buyOffers:  []store.BuyOffer{},
		sellOffers: []store.SellOffer{},
		mints:      []store.Mint{},
		invoices:   []store.UnconfirmedInvoice{},
	}

	privHex, pubHex, _, err := doge.GenerateDogecoinKeypair(doge.PrefixTestnet)
	if err != nil {
		t.Fatalf("Failed to generate dogecoin keypair: %v", err)
	}

	feClient := client.NewTokenisationClient(server.URL, privHex, pubHex)

	tokenisationStore := test_support.SetupTestDB()

	return tokenisationStore, dogenetClient, mux, feClient
}
