package rpc

import (
	"encoding/hex"
	"encoding/json"
	"net/http"

	"dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/dogenet"
	"dogecoin.org/fractal-engine/pkg/protocol"
	"dogecoin.org/fractal-engine/pkg/store"
)

type PaymentRoutes struct {
	store        *store.TokenisationStore
	gossipClient dogenet.GossipClient
	cfg          *config.Config
}

func HandlePaymentRoutes(store *store.TokenisationStore, gossipClient dogenet.GossipClient, mux *http.ServeMux, cfg *config.Config) {
	ir := &PaymentRoutes{store: store, gossipClient: gossipClient, cfg: cfg}

	mux.HandleFunc("/payments/new", ir.handleNewPayment)
}

func (ir *PaymentRoutes) handleNewPayment(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		ir.postNewPayment(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// @Summary		Prepares an encoded transaction body for a new payment
// @Description	Generates an encoded transaction body for paying an invoice
// @Tags			payments
// @Accept			json
// @Produce		json
// @Param			request	body		CreateNewPaymentRequest	true	"Payment request"
// @Success		201		{object}	map[string]string
// @Failure		400		{object}	string
// @Router			/payments/new [post]
func (ir *PaymentRoutes) postNewPayment(w http.ResponseWriter, r *http.Request) {
	var request CreateNewPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	envelope := protocol.NewPaymentTransactionEnvelope(request.InvoiceHash, protocol.ACTION_PAYMENT)
	encodedTransactionBody := envelope.Serialize()

	respondJSON(w, http.StatusCreated, map[string]string{
		"encoded_transaction_body": hex.EncodeToString(encodedTransactionBody),
	})
}
