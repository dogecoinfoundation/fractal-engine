package rpc

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/doge"
	"dogecoin.org/fractal-engine/pkg/dogenet"
	"dogecoin.org/fractal-engine/pkg/protocol"
	"dogecoin.org/fractal-engine/pkg/store"
	"dogecoin.org/fractal-engine/pkg/validation"
	"github.com/gorilla/mux"
)

type MintRoutes struct {
	store        *store.TokenisationStore
	gossipClient dogenet.GossipClient
	cfg          *config.Config
	dogeClient   *doge.RpcClient
}

func HandleMintRoutes(store *store.TokenisationStore, gossipClient dogenet.GossipClient, mux *http.ServeMux, cfg *config.Config, dogeClient *doge.RpcClient) {
	mr := &MintRoutes{store: store, gossipClient: gossipClient, cfg: cfg, dogeClient: dogeClient}

	mux.HandleFunc("/mints/{hash}", mr.handleMint)
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

func (mr *MintRoutes) handleMint(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		mr.getMint(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (mr *MintRoutes) getMint(w http.ResponseWriter, r *http.Request) {
	hash := validation.SanitizeQueryParam(mux.Vars(r)["hash"])

	// Validate hash format
	if err := validation.ValidateHash(hash); err != nil {
		http.Error(w, "Invalid hash format", http.StatusBadRequest)
		return
	}

	mint, err := mr.store.GetMintByHash(hash)
	if err != nil {
		http.Error(w, "Mint not found", http.StatusNotFound)
		return
	}

	response := GetMintResponse{
		Mint: mint,
	}

	respondJSON(w, http.StatusOK, response)
}

// @Summary		Get all mints
// @Description	Returns a list of mints
// @Tags			mints
// @Accept			json
// @Produce		json
// @Param			limit	query		int	false	"Limit"
// @Param			page	query		int	false	"Page"
// @Success		200		{object}	GetMintsResponse
// @Failure		400		{object}	string
// @Failure		500		{object}	string
// @Router			/mints [get]
func (mr *MintRoutes) getMints(w http.ResponseWriter, r *http.Request) {
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

	publicKey := validation.SanitizeQueryParam(r.URL.Query().Get("public_key"))
	address := validation.SanitizeQueryParam(r.URL.Query().Get("address"))
	includeUnconfirmed := validation.SanitizeQueryParam(r.URL.Query().Get("include_unconfirmed")) == "true"

	// Validate public key format if provided
	if publicKey != "" {
		if err := validation.ValidatePublicKey(publicKey); err != nil {
			http.Error(w, "Invalid public key format", http.StatusBadRequest)
			return
		}
	}

	start := page * limit
	end := start + limit

	var mints []store.Mint
	var err error

	if publicKey != "" {
		mints, err = mr.store.GetMintsByPublicKey(start, end, publicKey, includeUnconfirmed)
		if err != nil {
			log.Println(err)
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
	} else if address != "" {
		mints, err = mr.store.GetMintsByAddress(start, end, address, includeUnconfirmed)
		if err != nil {
			log.Println(err)
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
	} else {
		mints, err = mr.store.GetMints(start, end)
		if err != nil {
			log.Println(err)
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
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

// @Summary		Create a mint
// @Description	Creates a new mint
// @Tags			mints
// @Accept			json
// @Produce		json
// @Param			request	body		CreateMintRequest	true	"Mint request"
// @Success		201		{object}	CreateMintResponse
// @Failure		400		{object}	string
// @Failure		500		{object}	string
// @Router			/mints [post]
func (mr *MintRoutes) postMint(w http.ResponseWriter, r *http.Request) {
	var request CreateMintRequest

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

	err = request.Validate()
	if err != nil {
		log.Println("error validating mint", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newMintWithoutId := &store.MintWithoutID{
		Title:         request.Payload.Title,
		FractionCount: request.Payload.FractionCount,
		Description:   request.Payload.Description,
		Tags:          request.Payload.Tags,
		Metadata:      request.Payload.Metadata,
		CreatedAt:     time.Now(),
		Requirements:  request.Payload.Requirements,
		LockupOptions: request.Payload.LockupOptions,
		FeedURL:       request.Payload.FeedURL,
		PublicKey:     request.PublicKey,
		Signature:     request.Signature,
		OwnerAddress:  request.Address,
	}

	newMintWithoutId.Hash, err = newMintWithoutId.GenerateHash()
	if err != nil {
		http.Error(w, "Failed to generate hash", http.StatusBadRequest)
		return
	}

	id, err := mr.store.SaveUnconfirmedMint(newMintWithoutId)
	if err != nil {
		log.Println("error saving unconfirmed mint", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	newMint := &store.Mint{
		MintWithoutID: *newMintWithoutId,
		Id:            id,
	}

	err = mr.gossipClient.GossipMint(*newMint)
	if err != nil {
		http.Error(w, "Unable to gossip", http.StatusInternalServerError)
		return
	}

	envelope := protocol.NewMintTransactionEnvelope(newMintWithoutId.Hash, protocol.ACTION_MINT)
	encodedTransactionBody := envelope.Serialize()

	response := CreateMintResponse{
		Hash:                   newMintWithoutId.Hash,
		EncodedTransactionBody: hex.EncodeToString(encodedTransactionBody),
	}

	respondJSON(w, http.StatusCreated, response)
}
