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

// @Summary		Get token balances
// @Description	Returns token balances for an address, optionally filtered by mint hash and with optional mint details
// @Tags			Token Balances
// @Accept			json
// @Produce		json
// @Param			address				query		string	false	"Address to get token balances for"
// @Param			mint_hash			query		string	false	"Filter by mint hash"
// @Param			include_mint_details	query		boolean	false	"Include mint details in response"
// @Param			limit				query		int		false	"Limit number of results (max 100, only used with include_mint_details)"
// @Param			page				query		int		false	"Page number (max 1000, only used with include_mint_details)"
// @Success		200					{object}	interface{}
// @Failure		500					{object}	string
// @Router			/token-balances [get]
func (tr *TokenRoutes) getTokenBalances(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	mintHash := r.URL.Query().Get("mint_hash")
	includeMintDetails := r.URL.Query().Get("include_mint_details") == "true"

	var response interface{}

	if includeMintDetails {
		// Handle pagination for mint details
		limitStr := validation.SanitizeQueryParam(r.URL.Query().Get("limit"))
		limit := 100

		if limitStr != "" {
			if l, parseErr := strconv.Atoi(limitStr); parseErr == nil && l > 0 && l <= limit {
				limit = l
			}
		}

		pageStr := validation.SanitizeQueryParam(r.URL.Query().Get("page"))
		page := 0

		if pageStr != "" {
			if p, parseErr := strconv.Atoi(pageStr); parseErr == nil && p > 0 && p <= 1000 {
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
			response = GetTokenBalanceWithMintsResponse{}
		} else {
			if end > len(tokenBalances) {
				end = len(tokenBalances)
			}
			response = GetTokenBalanceWithMintsResponse{
				Mints: tokenBalances[start:end],
				Total: len(tokenBalances),
				Page:  page,
				Limit: limit,
			}
		}
	} else {
		// Simple token balances without mint details
		tokenBalances, err := tr.store.GetTokenBalances(address, mintHash)
		if err != nil {
			http.Error(w, "Failed to get tokens", http.StatusInternalServerError)
			return
		}
		response = tokenBalances
	}

	respondJSON(w, http.StatusOK, response)
}

// @Summary		Get pending token balances
// @Description	Returns pending token balances for an address, optionally filtered by mint hash
// @Tags			Token Balances
// @Accept			json
// @Produce		json
// @Param			address		query		string	false	"Address to get pending token balances for"
// @Param			mint_hash	query		string	false	"Filter by mint hash"
// @Success		200			{object}	[]store.TokenBalance
// @Failure		500			{object}	string
// @Router			/pending-token-balances [get]
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
