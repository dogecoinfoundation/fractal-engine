package rpc

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/dogenet"
	"dogecoin.org/fractal-engine/pkg/store"
)

type OfferRoutes struct {
	store        *store.TokenisationStore
	gossipClient dogenet.GossipClient
	cfg          *config.Config
}

func HandleOfferRoutes(store *store.TokenisationStore, gossipClient dogenet.GossipClient, mux *http.ServeMux, cfg *config.Config) {
	or := &OfferRoutes{store: store, gossipClient: gossipClient, cfg: cfg}

	mux.HandleFunc("/buy-offers/delete", or.handleDeleteBuyOffer)
	mux.HandleFunc("/buy-offers", or.handleBuyOffers)
	mux.HandleFunc("/sell-offers/delete", or.handleDeleteSellOffer)
	mux.HandleFunc("/sell-offers", or.handleSellOffers)
}

func (or *OfferRoutes) handleBuyOffers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		or.getBuyOffers(w, r)
	case http.MethodPost:
		or.postBuyOffer(w, r)
	case http.MethodDelete:
		or.deleteBuyOffer(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (or *OfferRoutes) handleSellOffers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		or.getSellOffers(w, r)
	case http.MethodPost:
		or.postSellOffer(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (or *OfferRoutes) handleDeleteBuyOffer(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		or.deleteBuyOffer(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (or *OfferRoutes) handleDeleteSellOffer(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		or.deleteSellOffer(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (or *OfferRoutes) deleteBuyOffer(w http.ResponseWriter, r *http.Request) {
	var request DeleteBuyOfferRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err := request.Validate()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = or.store.DeleteBuyOffer(request.Payload.OfferHash, request.PublicKey)
	if err != nil {
		http.Error(w, "Failed to delete buy offer", http.StatusBadRequest)
		return
	}

	err = or.gossipClient.GossipDeleteBuyOffer(request.Payload.OfferHash, request.PublicKey, request.Signature)
	if err != nil {
		http.Error(w, "Unable to gossip", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, "Buy offer deleted")
}

func (or *OfferRoutes) deleteSellOffer(w http.ResponseWriter, r *http.Request) {
	var request DeleteSellOfferRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err := request.Validate()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = or.store.DeleteSellOffer(request.Payload.OfferHash, request.PublicKey)
	if err != nil {
		http.Error(w, "Failed to delete sell offer", http.StatusBadRequest)
		return
	}

	err = or.gossipClient.GossipDeleteSellOffer(request.Payload.OfferHash, request.PublicKey, request.Signature)
	if err != nil {
		http.Error(w, "Unable to gossip", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, "Sell offer deleted")
}

func (or *OfferRoutes) getSellOffers(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 100

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l < limit {
			limit = l
		}
	}

	pageStr := r.URL.Query().Get("page")
	page := 0

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	// max 100 records per page
	limit = int(math.Min(float64(limit), 100))

	start := page * limit
	end := start + limit

	mintHash := r.URL.Query().Get("mint_hash")
	offererAddress := r.URL.Query().Get("offerer_address")

	offers, err := or.store.GetSellOffers(start, end, mintHash, offererAddress)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if start >= len(offers) {
		respondJSON(w, http.StatusOK, GetSellOffersResponse{})
		return
	}

	if end > len(offers) {
		end = len(offers)
	}

	offersWithMints := []SellOfferWithMint{}
	for _, offer := range offers {
		mint, err := or.store.GetMintByHash(offer.MintHash)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		offersWithMints = append(offersWithMints, SellOfferWithMint{
			Offer: offer,
			Mint:  mint,
		})
	}

	response := GetSellOffersResponse{
		Offers: offersWithMints[start:end],
		Total:  len(offers),
		Page:   page,
		Limit:  limit,
	}

	respondJSON(w, http.StatusOK, response)
}

func (or *OfferRoutes) postSellOffer(w http.ResponseWriter, r *http.Request) {
	var request CreateSellOfferRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err := request.Validate()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	count, err := or.store.CountSellOffers(request.Payload.MintHash, request.Payload.OffererAddress)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if count >= or.cfg.SellOfferLimit {
		http.Error(w, "Sell offer limit reached", http.StatusBadRequest)
		return
	}

	newOfferWithoutId := &store.SellOfferWithoutID{
		OffererAddress: request.Payload.OffererAddress,
		MintHash:       request.Payload.MintHash,
		Quantity:       request.Payload.Quantity,
		Price:          request.Payload.Price,
		CreatedAt:      time.Now(),
		PublicKey:      request.PublicKey,
		Signature:      request.Signature,
	}
	newOfferWithoutId.Hash, err = newOfferWithoutId.GenerateHash()
	if err != nil {
		http.Error(w, "Failed to generate hash", http.StatusBadRequest)
		return
	}

	id, err := or.store.SaveSellOffer(newOfferWithoutId)
	if err != nil {
		http.Error(w, "Failed to save sell offer", http.StatusBadRequest)
		return
	}

	newOffer := &store.SellOffer{
		SellOfferWithoutID: *newOfferWithoutId,
		Id:                 id,
	}

	err = or.gossipClient.GossipSellOffer(*newOffer)
	if err != nil {
		http.Error(w, "Unable to gossip", http.StatusInternalServerError)
		return
	}

	response := CreateOfferResponse{
		Id:   id,
		Hash: newOfferWithoutId.Hash,
	}

	respondJSON(w, http.StatusCreated, response)
}

// @Summary		Get all buy offers
// @Description	Returns a list of buy offers
// @Tags			buy-offers
// @Accept			json
// @Produce		json
// @Param			limit	query		int	false	"Limit"
// @Param			page	query		int	false	"Page"
// @Param			mint_hash	query		string	false	"Mint hash"
// @Param			type	query		int	false	"Type"
// @Success		200		{object}	GetOffersResponse
// @Failure		400		{object}	string
// @Failure		500		{object}	string
// @Router			/buy-offers [get]

func (or *OfferRoutes) getBuyOffers(w http.ResponseWriter, r *http.Request) {
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
	page := 0

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	mintHash := r.URL.Query().Get("mint_hash")
	sellerAddress := r.URL.Query().Get("seller_address")

	start := page * limit
	end := start + limit

	offers, err := or.store.GetBuyOffersByMintAndSellerAddress(start, end, mintHash, sellerAddress)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Clamp the slice range
	if start >= len(offers) {
		respondJSON(w, http.StatusOK, GetBuyOffersResponse{})
		return
	}

	if end > len(offers) {
		end = len(offers)
	}

	offersWithMints := []BuyOfferWithMint{}
	for _, offer := range offers {
		mint, err := or.store.GetMintByHash(offer.MintHash)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		offersWithMints = append(offersWithMints, BuyOfferWithMint{
			Offer: offer,
			Mint:  mint,
		})
	}

	response := GetBuyOffersResponse{
		Offers: offersWithMints[start:end],
		Total:  len(offers),
		Page:   page,
		Limit:  limit,
	}

	respondJSON(w, http.StatusOK, response)
}

// @Summary		Create a buy offer
// @Description	Creates a new buy offer
// @Tags			buy-offers
// @Accept			json
// @Produce		json
// @Param			request	body		CreateBuyOfferRequest	true	"Buy offer request"
// @Success		201		{object}	CreateOfferResponse
// @Failure		400		{object}	string
// @Failure		500		{object}	string
// @Router			/buy-offers [post]
func (or *OfferRoutes) postBuyOffer(w http.ResponseWriter, r *http.Request) {
	var request CreateBuyOfferRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err := request.Validate()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	count, err := or.store.CountBuyOffers(request.Payload.MintHash, request.Payload.OffererAddress, request.Payload.SellerAddress)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if count >= or.cfg.BuyOfferLimit {
		http.Error(w, "Buy offer limit reached", http.StatusBadRequest)
		return
	}

	newOfferWithoutId := &store.BuyOfferWithoutID{
		OffererAddress: request.Payload.OffererAddress,
		MintHash:       request.Payload.MintHash,
		SellerAddress:  request.Payload.SellerAddress,
		Quantity:       request.Payload.Quantity,
		Price:          request.Payload.Price,
		CreatedAt:      time.Now(),
		PublicKey:      request.PublicKey,
	}
	newOfferWithoutId.Hash, err = newOfferWithoutId.GenerateHash()
	if err != nil {
		http.Error(w, "Failed to generate hash", http.StatusBadRequest)
		return
	}

	id, err := or.store.SaveBuyOffer(newOfferWithoutId)
	if err != nil {
		http.Error(w, "Failed to save buy offer", http.StatusBadRequest)
		return
	}

	newOffer := &store.BuyOffer{
		BuyOfferWithoutID: *newOfferWithoutId,
		Id:                id,
	}

	err = or.gossipClient.GossipBuyOffer(*newOffer)
	if err != nil {
		http.Error(w, "Unable to gossip", http.StatusInternalServerError)
		return
	}

	response := CreateOfferResponse{
		Id:   id,
		Hash: newOfferWithoutId.Hash,
	}

	respondJSON(w, http.StatusCreated, response)
}
