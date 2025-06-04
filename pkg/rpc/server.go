package rpc

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/store"
)

type RpcServer struct {
	config  *config.Config
	quit    chan bool
	server  *http.Server
	Running bool
}

func NewRpcServer(cfg *config.Config, store *store.TokenisationStore) *RpcServer {
	mux := http.NewServeMux()

	handler := withCORS(mux)

	HandleMintRoutes(store, mux)

	server := &http.Server{
		Addr:    cfg.RpcServerHost + ":" + cfg.RpcServerPort,
		Handler: handler,
	}

	return &RpcServer{
		config:  cfg,
		server:  server,
		quit:    make(chan bool),
		Running: false,
	}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // or specific origin
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Proceed to actual handler
		next.ServeHTTP(w, r)
	})
}

func (s *RpcServer) Start() {
	go func() {
		log.Println("Server is ready to handle requests at :8080")
		s.Running = true
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on :8080: %v\n", err)
		}
	}()

	<-s.quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		log.Fatalf("Could not gracefully shutdown the server: %v\n", err)
	}

	log.Println("Server stopped")
}

func (s *RpcServer) Stop() {
	s.quit <- true
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
