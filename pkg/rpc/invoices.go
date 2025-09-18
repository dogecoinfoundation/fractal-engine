package rpc

import (
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/dogenet"
	"dogecoin.org/fractal-engine/pkg/protocol"
	"dogecoin.org/fractal-engine/pkg/store"
	"dogecoin.org/fractal-engine/pkg/validation"
)

type InvoiceRoutes struct {
	store        *store.TokenisationStore
	gossipClient dogenet.GossipClient
	cfg          *config.Config
}

func HandleInvoiceRoutes(store *store.TokenisationStore, gossipClient dogenet.GossipClient, mux *http.ServeMux, cfg *config.Config) {
	ir := &InvoiceRoutes{store: store, gossipClient: gossipClient, cfg: cfg}

	mux.HandleFunc("/invoices/{hash}/signatures", ir.handleCreateInvoiceSignature)
	mux.HandleFunc("/invoices", ir.handleInvoices)
	mux.HandleFunc("/invoices/{address}", ir.handleInvoices)

}

func (ir *InvoiceRoutes) handleCreateInvoiceSignature(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		ir.postCreateInvoiceSignature(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (ir *InvoiceRoutes) postCreateInvoiceSignature(w http.ResponseWriter, r *http.Request) {
	var request CreateInvoiceSignatureRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	newInvoiceSignature := &store.InvoiceSignature{
		InvoiceHash: request.Payload.InvoiceHash,
		Signature:   request.Payload.Signature,
		PublicKey:   request.Payload.PublicKey,
		CreatedAt:   time.Now(),
	}

	id, err := ir.store.SaveInvoiceSignature(newInvoiceSignature)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	response := CreateInvoiceSignatureResponse{
		Id: id,
	}

	respondJSON(w, http.StatusCreated, response)
}

type CreateInvoiceSignatureRequest struct {
	Payload CreateInvoiceSignatureRequestPayload `json:"payload"`
}

type CreateInvoiceSignatureRequestPayload struct {
	InvoiceHash string `json:"invoice_hash"`
	Signature   string `json:"signature"`
	PublicKey   string `json:"public_key"`
}

type CreateInvoiceSignatureResponse struct {
	Id string `json:"id"`
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

// @Summary		Get invoices
// @Description	Returns a list of invoices with optional filtering by mint_hash and address
// @Tags			invoices
// @Accept			json
// @Produce		json
// @Param			address			path		string	true	"Filter by address of buyer or seller"
// @Param			limit			query		int		false	"Limit number of results (max 100)"
// @Param			page			query		int		false	"Page number (max 1000)"
// @Param			mint_hash		query		string	false	"Filter by mint hash"
// @Success		200				{object}	GetInvoicesResponse
// @Failure		400				{object}	string
// @Router			/invoices/{address} [get]
func (ir *InvoiceRoutes) getInvoices(w http.ResponseWriter, r *http.Request) {
	// Extract address from URL path
	address := r.URL.Path[len("/invoices/"):]
	if address == "" {
		http.Error(w, "Address is required", http.StatusBadRequest)
		return
	}

	if err := validation.ValidateAddress(address); err != nil {
		http.Error(w, "Invalid address format", http.StatusBadRequest)
		return
	}

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

	mintHash := validation.SanitizeQueryParam(r.URL.Query().Get("mint_hash"))

	start := page * limit
	end := start + limit
	var invoices []store.Invoice
	var err error

	if mintHash == "" {
		invoices, err = ir.store.GetInvoicesForMe(start, end, address)
	} else {
		invoices, err = ir.store.GetInvoices(start, end, mintHash, address)
	}

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

	count, err := ir.store.CountUnconfirmedInvoices(request.Payload.MintHash, request.Payload.BuyerAddress)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if count >= ir.cfg.InvoiceLimit {
		http.Error(w, "Invoice limit reached", http.StatusBadRequest)
		return
	}

	mint, err := ir.store.GetMintByHash(request.Payload.MintHash)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var initialStatus string
	if mint.SignatureRequirementType == store.SignatureRequirementType_NONE || mint.SignatureRequirementType == "" {
		initialStatus = "draft"
	} else {
		initialStatus = "pending_signatures"
	}

	newInvoiceWithoutId := &store.UnconfirmedInvoice{
		MintHash:       request.Payload.MintHash,
		Quantity:       request.Payload.Quantity,
		Price:          request.Payload.Price,
		BuyerAddress:   request.Payload.BuyerAddress,
		PaymentAddress: request.Payload.PaymentAddress,
		CreatedAt:      time.Now(),
		SellerAddress:  request.Payload.SellerAddress,
		PublicKey:      request.PublicKey,
		Signature:      request.Signature,
		Status:         initialStatus,
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

	envelope := protocol.NewInvoiceTransactionEnvelope(newInvoiceWithoutId.Hash, newInvoiceWithoutId.SellerAddress, newInvoiceWithoutId.MintHash, int32(newInvoiceWithoutId.Quantity), protocol.ACTION_INVOICE)
	encodedTransactionBody := envelope.Serialize()

	response := CreateInvoiceResponse{
		Hash:                   newInvoiceWithoutId.Hash,
		EncodedTransactionBody: hex.EncodeToString(encodedTransactionBody),
	}

	respondJSON(w, http.StatusCreated, response)
}
