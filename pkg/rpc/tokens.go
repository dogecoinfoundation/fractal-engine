package rpc

import (
	"encoding/json"
	"net/http"

	"dogecoin.org/fractal-engine/pkg/store"
)

type TokenRoutes struct {
	store *store.TokenisationStore
}

func HandleTokenRoutes(store *store.TokenisationStore, mux *http.ServeMux) {
	tr := &TokenRoutes{store: store}

	mux.HandleFunc("/token-balances", tr.handleTokenBalances)
}

func (tr *TokenRoutes) handleTokenBalances(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		tr.getTokenBalances(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (tr *TokenRoutes) getTokenBalances(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	mintHash := r.URL.Query().Get("mint_hash")

	tokenBalances, err := tr.store.GetTokenBalances(address, mintHash)
	if err != nil {
		http.Error(w, "Failed to get tokens", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(tokenBalances)
}
