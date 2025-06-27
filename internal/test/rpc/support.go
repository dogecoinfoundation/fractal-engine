package test_rpc

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"dogecoin.org/fractal-engine/pkg/client"
	"dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/dogenet"
	"dogecoin.org/fractal-engine/pkg/store"
)

type FakeGossipClient struct {
	dogenet.GossipClient
	offers   []store.Offer
	mints    []store.Mint
	invoices []store.UnconfirmedInvoice
}

func (g *FakeGossipClient) GossipOffer(offer store.Offer) error {
	g.offers = append(g.offers, offer)
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

	testDir := os.TempDir()
	dbPath := filepath.Join(testDir, fmt.Sprintf("test_rpc_%d.db", rand.Intn(1000000)))

	tokenisationStore, err := store.NewTokenisationStore("sqlite:///"+dbPath, config.Config{
		MigrationsPath: "../../../db/migrations",
	})
	if err != nil {
		t.Fatalf("Failed to create tokenisation store: %v", err)
	}

	err = tokenisationStore.Migrate()
	if err != nil {
		t.Fatalf("Failed to migrate tokenisation store: %v", err)
	}

	dogenetClient := &FakeGossipClient{
		offers: []store.Offer{},
		mints:  []store.Mint{},
	}

	feClient := client.NewTokenisationClient(server.URL)

	return tokenisationStore, dogenetClient, mux, feClient
}
