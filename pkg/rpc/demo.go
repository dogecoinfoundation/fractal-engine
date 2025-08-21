package rpc

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/doge"
	"dogecoin.org/fractal-engine/pkg/store"
	dogelib "github.com/dogeorg/doge"
)

// setupDemoBalance

type DemoRoutes struct {
	store      *store.TokenisationStore
	cfg        *config.Config
	dogeClient *doge.RpcClient
}

func HandleDemoRoutes(store *store.TokenisationStore, mux *http.ServeMux, cfg *config.Config, dogeClient *doge.RpcClient) {
	dr := &DemoRoutes{store: store, cfg: cfg, dogeClient: dogeClient}

	mux.HandleFunc("/setup-demo-balance", dr.handleSetupDemoBalance)
	mux.HandleFunc("/list-unspent", dr.handleListUnspent)
	mux.HandleFunc("/sign-tx", dr.handleSignTx)
}

func (dr *DemoRoutes) handleSetupDemoBalance(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		dr.postSetupDemoBalance(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (dr *DemoRoutes) handleListUnspent(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		dr.getListUnspent(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (dr *DemoRoutes) handleSignTx(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		dr.postSignTx(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (dr *DemoRoutes) postSignTx(w http.ResponseWriter, r *http.Request) {
	var request SignTxRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

}

func (dr *DemoRoutes) getListUnspent(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "Address is required", http.StatusBadRequest)
		return
	}

	unspent, err := dr.dogeClient.ListUnspent(address)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, unspent)
}

func (dr *DemoRoutes) postSetupDemoBalance(w http.ResponseWriter, r *http.Request) {
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

func (dr *DemoRoutes) SetupAddress(label string, initialBalance int64) (Address, error) {
	_, err := dr.dogeClient.Generate(100)
	if err != nil {
		return Address{}, err
	}

	address, err := dr.dogeClient.GetNewAddress()
	if err != nil {
		return Address{}, err
	}

	privKey, err := dr.dogeClient.DumpPrivKey(address)
	if err != nil {
		return Address{}, err
	}

	// chain := dogelib.ChainFromWIFString(privKey)

	privKeyBytes, _, err := dogelib.DecodeECPrivKeyWIF(privKey, &dogelib.DogeRegTestChain)
	if err != nil {
		log.Println("error decoding priv key", err)
		return Address{}, err
	}

	pubKey := dogelib.ECPubKeyFromECPrivKey(privKeyBytes)
	pubKeyHex := hex.EncodeToString(pubKey[:])

	_, err = dr.dogeClient.SendToAddress(address, initialBalance)
	if err != nil {
		return Address{}, err
	}

	newAddress := Address{
		Label:      label,
		Address:    address,
		PrivateKey: privKey,
		PublicKey:  pubKeyHex,
	}

	_, err = dr.dogeClient.Generate(1)
	if err != nil {
		return Address{}, err
	}

	return newAddress, nil
}
