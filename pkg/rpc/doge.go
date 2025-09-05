package rpc

import (
	"bytes"
	"encoding/json"
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

	respondJSON(w, http.StatusCreated, map[string]string{
		"transaction_id": txid,
	})
}

func (dr *DogeRoutes) postConfirm(w http.ResponseWriter, r *http.Request) {
	_, err := dr.dogeClient.Request("generate", []interface{}{10})
	if err != nil {
		log.Println("error sending raw transaction", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	respondJSON(w, http.StatusCreated, map[string]string{})
}
