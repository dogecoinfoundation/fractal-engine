package test_rpc

import (
	"net/http"
	"net/http/httptest"
	"testing"

	test_support "dogecoin.org/fractal-engine/internal/test/support"
	"dogecoin.org/fractal-engine/pkg/client"
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

func SetupRpcTest(t *testing.T) (*store.TokenisationStore, *FakeGossipClient, *http.ServeMux, *client.TokenisationClient) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	dogenetClient := &FakeGossipClient{
		buyOffers:  []store.BuyOffer{},
		sellOffers: []store.SellOffer{},
		mints:      []store.Mint{},
		invoices:   []store.UnconfirmedInvoice{},
	}

	feClient := client.NewTokenisationClient(server.URL)

	tokenisationStore := test_support.SetupTestDB(t)

	return tokenisationStore, dogenetClient, mux, feClient
}
