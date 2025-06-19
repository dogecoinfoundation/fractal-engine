package rpc

import (
	"log"
	"net/http"

	"dogecoin.org/fractal-engine/pkg/store"
)

type StatRoutes struct {
	store *store.TokenisationStore
}

func HandleStatRoutes(store *store.TokenisationStore, mux *http.ServeMux) {
	sr := &StatRoutes{store: store}

	mux.HandleFunc("/stats", sr.handleStats)
}

func (sr *StatRoutes) handleStats(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		sr.getStats(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (sr *StatRoutes) getStats(w http.ResponseWriter, _ *http.Request) {
	stats, err := sr.store.GetStats()
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	response := GetStatsResponse{
		Stats: stats,
	}

	respondJSON(w, http.StatusOK, response)
}
