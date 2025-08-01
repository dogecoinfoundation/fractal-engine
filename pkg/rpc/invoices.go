package rpc

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/dogenet"

	"dogecoin.org/fractal-engine/pkg/store"
)

type InvoiceRoutes struct {
	store        *store.TokenisationStore
	gossipClient dogenet.GossipClient
	cfg          *config.Config
}

func HandleInvoiceRoutes(store *store.TokenisationStore, gossipClient dogenet.GossipClient, mux *http.ServeMux, cfg *config.Config) {
	ir := &InvoiceRoutes{store: store, gossipClient: gossipClient, cfg: cfg}

	mux.HandleFunc("/invoices", ir.handleInvoices)
}

func (ir *InvoiceRoutes) handleInvoices(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		ir.getInvoices(w, r)
	case http.MethodPost:
		ir.postInvoice(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// @Summary		Get all invoices
// @Description	Returns a list of invoices
// @Tags			invoices
// @Accept			json
// @Produce		json
// @Param			limit	query		int	false	"Limit"
// @Param			page	query		int	false	"Page"
// @Param			mint_hash	query		string	false	"Mint hash"
// @Param			offerer_address	query		string	false	"Offerer address"
// @Success		200		{object}	GetInvoicesResponse
// @Failure		400		{object}	string
// @Failure		500		{object}	string
// @Router			/invoices [get]

func (ir *InvoiceRoutes) getInvoices(w http.ResponseWriter, r *http.Request) {
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

	mintHash := r.URL.Query().Get("mint_hash")
	offererAddress := r.URL.Query().Get("offerer_address")

	start := (page - 1) * limit
	end := start + limit

	invoices, err := ir.store.GetInvoices(start, end, mintHash, offererAddress)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Clamp the slice range
	if start >= len(invoices) {
		respondJSON(w, http.StatusOK, GetInvoicesResponse{})
		return
	}

	if end > len(invoices) {
		end = len(invoices)
	}

	response := GetInvoicesResponse{
		Invoices: invoices[start:end],
		Total:    len(invoices),
		Page:     page,
		Limit:    limit,
	}

	respondJSON(w, http.StatusOK, response)
}

// @Summary		Create an invoice
// @Description	Creates a new invoice
// @Tags			invoices
// @Accept			json
// @Produce		json
// @Param			request	body		CreateInvoiceRequest	true	"Invoice request"
// @Success		201		{object}	CreateInvoiceResponse
// @Failure		400		{object}	string
// @Failure		500		{object}	string
// @Router			/invoices [post]

func (ir *InvoiceRoutes) postInvoice(w http.ResponseWriter, r *http.Request) {
	var request CreateInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err := request.Validate()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	count, err := ir.store.CountUnconfirmedInvoices(request.Payload.BuyOfferMintHash, request.Payload.BuyOfferOffererAddress)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if count >= ir.cfg.InvoiceLimit {
		http.Error(w, "Invoice limit reached", http.StatusBadRequest)
		return
	}

	newInvoiceWithoutId := &store.UnconfirmedInvoice{
		BuyOfferHash:           request.Payload.BuyOfferHash,
		BuyOfferMintHash:       request.Payload.BuyOfferMintHash,
		BuyOfferQuantity:       request.Payload.BuyOfferQuantity,
		BuyOfferPrice:          request.Payload.BuyOfferPrice,
		BuyOfferOffererAddress: request.Payload.BuyOfferOffererAddress,
		PaymentAddress:         request.Payload.PaymentAddress,
		CreatedAt:              time.Now(),
		SellOfferAddress:       request.Payload.SellOfferAddress,
		PublicKey:              request.PublicKey,
	}

	newInvoiceWithoutId.Hash, err = newInvoiceWithoutId.GenerateHash()
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	id, err := ir.store.SaveUnconfirmedInvoice(newInvoiceWithoutId)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	newInvoiceWithoutId.Id = id

	err = ir.gossipClient.GossipUnconfirmedInvoice(*newInvoiceWithoutId)
	if err != nil {
		http.Error(w, "Unable to gossip", http.StatusInternalServerError)
		return
	}

	response := CreateInvoiceResponse{
		Hash: newInvoiceWithoutId.Hash,
	}

	respondJSON(w, http.StatusCreated, response)
}
