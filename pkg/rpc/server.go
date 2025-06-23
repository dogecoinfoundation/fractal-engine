package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/dogenet"
	"dogecoin.org/fractal-engine/pkg/store"
)

type RpcServer struct {
	config  *config.Config
	quit    chan bool
	server  *http.Server
	Running bool
	dogenet *dogenet.DogeNetClient
}

func NewRpcServer(cfg *config.Config, store *store.TokenisationStore, dogenet *dogenet.DogeNetClient) *RpcServer {
	mux := http.NewServeMux()

	handler := withCORS(mux)

	HandleMintRoutes(store, dogenet, mux)
	HandleStatRoutes(store, mux)

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
		log.Println("Server is ready to handle requests at " + s.server.Addr)
		s.Running = true
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on %s: %v\n", s.server.Addr, err)
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
	fmt.Println("Stopping rpc server")
	s.quit <- true
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
