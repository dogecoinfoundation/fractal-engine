package rpc

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"dogecoin.org/fractal-engine/pkg/protocol"
	"dogecoin.org/fractal-engine/pkg/store"
)

type MintRoutes struct {
	store *store.TokenisationStore
}

func HandleMintRoutes(store *store.TokenisationStore, mux *http.ServeMux) {
	mr := &MintRoutes{store: store}

	mux.HandleFunc("/mints", mr.handleMints)
}

func (mr *MintRoutes) handleMints(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		mr.getMints(w, r)
	case http.MethodPost:
		mr.postMint(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (mr *MintRoutes) getMints(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 100

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l < limit {
			limit = l
		}
	}

	pageStr := r.URL.Query().Get("page")
	page := 1

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	verifiedStr := r.URL.Query().Get("verified")
	verified := false

	if verifiedStr != "" {
		if v, err := strconv.ParseBool(verifiedStr); err == nil {
			verified = v
		}
	}

	start := (page - 1) * limit
	end := start + limit

	mints, err := mr.store.GetMints(start, end, verified)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Clamp the slice range
	if start >= len(mints) {
		respondJSON(w, http.StatusOK, GetMintsResponse{})
		return
	}

	if end > len(mints) {
		end = len(mints)
	}

	response := GetMintsResponse{
		Mints: mints[start:end],
		Total: len(mints),
		Page:  page,
		Limit: limit,
	}

	respondJSON(w, http.StatusOK, response)
}

func (mr *MintRoutes) postMint(w http.ResponseWriter, r *http.Request) {
	var request CreateMintRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err := request.Validate()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hash, err := request.GenerateHash()
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	id, err := mr.store.SaveMint(&store.MintWithoutID{
		Hash:            hash,
		Title:           request.Title,
		FractionCount:   request.FractionCount,
		Description:     request.Description,
		Tags:            request.Tags,
		Metadata:        request.Metadata,
		TransactionHash: request.TransactionHash,
		Verified:        request.Verified,
		CreatedAt:       time.Now(),
		Requirements:    request.Requirements,
		Resellable:      request.Resellable,
		LockupOptions:   request.LockupOptions,
	})

	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	envelope := protocol.NewMintTransactionEnvelope(hash)
	encodedTransactionBody := envelope.Serialize()

	response := CreateMintResponse{
		EncodedTransactionBody: encodedTransactionBody,
		Id:                     id,
		TransactionHash:        hash,
	}

	respondJSON(w, http.StatusCreated, response)
}
