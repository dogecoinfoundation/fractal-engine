package rpc

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"dogecoin.org/fractal-engine/pkg/doge"
	"dogecoin.org/fractal-engine/pkg/protocol"
	"dogecoin.org/fractal-engine/pkg/store"
)

type DogeRoutes struct {
	store      *store.TokenisationStore
	dogeClient *doge.RpcClient
}

func HandleDogeRoutes(store *store.TokenisationStore, dogeClient *doge.RpcClient, mux *http.ServeMux) {
	dr := &DogeRoutes{store: store, dogeClient: dogeClient}

	mux.HandleFunc("/doge/mint", dr.handleRawTransactionsMint)
	mux.HandleFunc("/doge/send", dr.handleSend)
}

func (dr *DogeRoutes) handleRawTransactionsMint(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		dr.postRawMintTransactions(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (dr *DogeRoutes) handleSend(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		dr.postSend(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

type CreateRawMintTransactionRequest struct {
	MintHash string `json:"mint_hash"`
	Address  string `json:"address"`
	TxID     string `json:"tx_id"`
	VOut     int    `json:"vout"`
	Value    int    `json:"value"`
}

type SendRequest struct {
	EncodedTrxn string `json:"encoded_transaction_hex"`
}

func (dr *DogeRoutes) postRawMintTransactions(w http.ResponseWriter, r *http.Request) {
	var request CreateRawMintTransactionRequest

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

	envelope := protocol.NewMintTransactionEnvelope(request.MintHash, protocol.ACTION_MINT)
	encodedTransactionBody := envelope.Serialize()

	inputs := []interface{}{
		map[string]interface{}{
			"txid": request.TxID,
			"vout": request.VOut,
		},
	}

	address := request.Address

	outputs := map[string]interface{}{
		"data":  hex.EncodeToString(encodedTransactionBody),
		address: request.Value - 1,
	}

	fmt.Println(inputs)

	rawTx, err := dr.dogeClient.Request("createrawtransaction", []interface{}{inputs, outputs})
	if err != nil {
		log.Fatal(err)
	}

	var rawTxResponse string
	if err := json.Unmarshal(*rawTx, &rawTxResponse); err != nil {
		log.Fatal(err)
	}

	fmt.Println("RAW TRX HEX", rawTxResponse)

	respondJSON(w, http.StatusCreated, map[string]string{
		"raw_transaction_hex": rawTxResponse,
	})
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

	fmt.Println("SIGNED TRX HEX", request.EncodedTrxn)

	res, err := dr.dogeClient.Request("sendrawtransaction", []interface{}{request.EncodedTrxn})
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
