package rpc

import (
	"encoding/json"
	"net/http"
	"strconv"

	"dogecoin.org/fractal-engine/pkg/store"
	"dogecoin.org/fractal-engine/pkg/validation"
)

type TokenRoutes struct {
	store *store.TokenisationStore
}

func HandleTokenRoutes(store *store.TokenisationStore, mux *http.ServeMux) {
	tr := &TokenRoutes{store: store}

	mux.HandleFunc("/token-balances", tr.handleTokenBalances)
	mux.HandleFunc("/mint-token-balances", tr.getTokenBalancesWithMints)
	mux.HandleFunc("/pending-token-balances", tr.handlePendingTokenBalances)
}

func (tr *TokenRoutes) handlePendingTokenBalances(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		tr.getPendingTokenBalances(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (tr *TokenRoutes) handleTokenBalances(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		tr.getTokenBalances(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (tr *TokenRoutes) getTokenBalancesWithMints(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	limitStr := validation.SanitizeQueryParam(r.URL.Query().Get("limit"))
	limit := 100

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= limit {
			limit = l
		}
	}

	pageStr := validation.SanitizeQueryParam(r.URL.Query().Get("page"))
	page := 0

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 && p <= 1000 { // Reasonable page limit
			page = p
		}
	}

	start := page * limit
	end := start + limit

	tokenBalances, err := tr.store.GetMyMintTokenBalances(address, start, end)
	if err != nil {
		http.Error(w, "Failed to get tokens", http.StatusInternalServerError)
		return
	}

	// Clamp the slice range
	if start >= len(tokenBalances) {
		respondJSON(w, http.StatusOK, GetTokenBalanceWithMintsResponse{})
		return
	}

	if end > len(tokenBalances) {
		end = len(tokenBalances)
	}

	response := GetTokenBalanceWithMintsResponse{
		Mints: tokenBalances[start:end],
		Total: len(tokenBalances),
		Page:  page,
		Limit: limit,
	}

	respondJSON(w, http.StatusOK, response)
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

func (tr *TokenRoutes) getPendingTokenBalances(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	mintHash := r.URL.Query().Get("mint_hash")

	tokenBalances, err := tr.store.GetPendingTokenBalances(address, mintHash)
	if err != nil {
		http.Error(w, "Failed to get tokens", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(tokenBalances)
}
