package rpc

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"dogecoin.org/fractal-engine/pkg/dogenet"
	"dogecoin.org/fractal-engine/pkg/store"
)

type OfferRoutes struct {
	store        *store.TokenisationStore
	gossipClient dogenet.GossipClient
}

func HandleOfferRoutes(store *store.TokenisationStore, gossipClient dogenet.GossipClient, mux *http.ServeMux) {
	or := &OfferRoutes{store: store, gossipClient: gossipClient}

	mux.HandleFunc("/offers", or.handleOffers)
}

func (or *OfferRoutes) handleOffers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		or.getOffers(w, r)
	case http.MethodPost:
		or.postOffer(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (or *OfferRoutes) getOffers(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 100

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l < limit {
			limit = l
		}
	}

	// max 100 records per page
	limit = int(math.Min(float64(limit), 100))

	pageStr := r.URL.Query().Get("page")
	page := 1

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	mintHash := r.URL.Query().Get("mint_hash")
	typeStr := r.URL.Query().Get("type")
	typeInt := 0
	if typeStr != "" {
		if t, err := strconv.Atoi(typeStr); err == nil && t >= 0 && t <= 1 {
			typeInt = t
		}
	}

	start := (page - 1) * limit
	end := start + limit

	offers, err := or.store.GetOffers(start, end, mintHash, typeInt)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Clamp the slice range
	if start >= len(offers) {
		respondJSON(w, http.StatusOK, GetOffersResponse{})
		return
	}

	if end > len(offers) {
		end = len(offers)
	}

	response := GetOffersResponse{
		Offers: offers[start:end],
		Total:  len(offers),
		Page:   page,
		Limit:  limit,
	}

	respondJSON(w, http.StatusOK, response)
}

func (or *OfferRoutes) postOffer(w http.ResponseWriter, r *http.Request) {
	var request CreateOfferRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err := request.Validate()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newOfferWithoutId := &store.OfferWithoutID{
		Type:           request.Type,
		OffererAddress: request.OffererAddress,
		MintHash:       request.MintHash,
		Quantity:       request.Quantity,
		Price:          request.Price,
		CreatedAt:      time.Now(),
	}
	newOfferWithoutId.Hash, err = newOfferWithoutId.GenerateHash()
	if err != nil {
		http.Error(w, "Failed to generate hash", http.StatusBadRequest)
		return
	}

	id, err := or.store.SaveOffer(newOfferWithoutId)
	if err != nil {
		http.Error(w, "Failed to save offer", http.StatusBadRequest)
		return
	}

	newOffer := &store.Offer{
		OfferWithoutID: *newOfferWithoutId,
		Id:             id,
	}

	err = or.gossipClient.GossipOffer(*newOffer)
	if err != nil {
		http.Error(w, "Unable to gossip", http.StatusInternalServerError)
		return
	}

	response := CreateOfferResponse{
		Id: id,
	}

	respondJSON(w, http.StatusCreated, response)
}
