package rpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"dogecoin.org/fractal-engine/pkg/doge"
	"dogecoin.org/fractal-engine/pkg/store"
)

type DogeRoutes struct {
	store      *store.TokenisationStore
	dogeClient *doge.RpcClient
}

func HandleDogeRoutes(store *store.TokenisationStore, dogeClient *doge.RpcClient, mux *http.ServeMux) {
	dr := &DogeRoutes{store: store, dogeClient: dogeClient}

	mux.HandleFunc("/doge/send", dr.handleSend)
	mux.HandleFunc("/doge/confirm", dr.handleConfirm)
	mux.HandleFunc("/doge/top-up", dr.handleTopUp)
}

func (dr *DogeRoutes) handleSend(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		dr.postSend(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (dr *DogeRoutes) handleConfirm(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		dr.postConfirm(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

type SendRequest struct {
	EncodedTrxn string `json:"encoded_transaction_hex"`
}

type SendResponse struct {
	TransactionID string `json:"transaction_id"`
}

// @Summary		Send a raw transaction
// @Description	Sends a raw transaction to the Dogecoin network
// @Tags			doge
// @Accept			json
// @Produce		json
// @Param			request	body		SendRequest	true	"Send transaction request"
// @Success		201		{object}	SendResponse
// @Failure		400		{object}	string
// @Failure		500		{object}	string
// @Router			/doge/send [post]
func (dr *DogeRoutes) postSend(w http.ResponseWriter, r *http.Request) {
	var request SendRequest

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("error reading request body", err)
		http.Error(w, "Error reading request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&request); err != nil {
		log.Println("error decoding request", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	res, err := dr.dogeClient.Request("sendrawtransaction", []interface{}{request.EncodedTrxn, true})
	if err != nil {
		log.Println("error sending raw transaction", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var txid string

	if err := json.Unmarshal(*res, &txid); err != nil {
		log.Println("error parsing send raw transaction response", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	respondJSON(w, http.StatusCreated, SendResponse{
		TransactionID: txid,
	})
}

// @Summary		Confirm transactions by generating blocks (Regtest only)
// @Description	Generates 10 blocks for transaction confirmation
// @Tags			doge
// @Accept			json
// @Produce		json
// @Success		201		{object}	map[string]string
// @Failure		400		{object}	string
// @Failure		500		{object}	string
// @Router			/doge/confirm [post]
func (dr *DogeRoutes) postConfirm(w http.ResponseWriter, r *http.Request) {
	_, err := dr.dogeClient.Request("generate", []interface{}{10})
	if err != nil {
		log.Println("error sending raw transaction", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	respondJSON(w, http.StatusCreated, map[string]string{})
}

func (dr *DogeRoutes) handleTopUp(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		dr.postTopUp(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// @Summary		Top up an address with test DOGE (Regtest only)
// @Description	Sends 1000 DOGE to the specified address for testing/development
// @Tags			doge
// @Accept			json
// @Produce		json
// @Param			address	query		string	true	"Dogecoin address to send funds to"
// @Success		200		{object}	string
// @Failure		400		{object}	string
// @Failure		500		{object}	string
// @Router			/doge/top-up [post]
func (dr *DogeRoutes) postTopUp(w http.ResponseWriter, r *http.Request) {
	_, err := dr.dogeClient.Generate(101)
	if err != nil {
		fmt.Println("error generating blocks", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	address := r.URL.Query().Get("address")
	if address == "" {
		fmt.Println("address is required")
		http.Error(w, "Address is required", http.StatusBadRequest)
		return
	}

	_, err = dr.dogeClient.SendToAddress(address, 1000)
	if err != nil {
		fmt.Println("error sending to address", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = dr.dogeClient.Generate(1)
	if err != nil {
		fmt.Println("error generating blocks", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, address)
}
