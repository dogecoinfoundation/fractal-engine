package rpc

import (
	"database/sql"
	"net/http"

	"dogecoin.org/fractal-engine/pkg/store"
)

type HealthRoutes struct {
	store *store.TokenisationStore
}

func HandleHealthRoutes(store *store.TokenisationStore, mux *http.ServeMux) {
	hr := &HealthRoutes{store: store}

	mux.HandleFunc("/health", hr.handleHealth)
}

func (hr *HealthRoutes) handleHealth(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		hr.getHealth(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// @Summary		Get health
// @Description	Returns the current and latest block height
// @Tags			health
// @Accept			json
// @Produce		json
// @Success		200		{object}	GetHealthResponse
// @Failure		400		{object}	string
// @Router			/health [get]
func (hr *HealthRoutes) getHealth(w http.ResponseWriter, _ *http.Request) {
	currentBlockHeight, latestBlockHeight, chain, walletsEnabled, updatedAt, err := hr.store.GetHealth()
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No health data found", http.StatusNotFound)
			return
		}
		http.Error(w, "Error getting health", http.StatusInternalServerError)
		return
	}

	response := GetHealthResponse{
		CurrentBlockHeight: currentBlockHeight,
		LatestBlockHeight:  latestBlockHeight,
		UpdatedAt:          updatedAt,
		Chain:              chain,
		WalletsEnabled:     walletsEnabled,
	}

	respondJSON(w, http.StatusOK, response)
}
