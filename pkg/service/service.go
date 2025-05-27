package service

import (
	"log"
	"os"
	"os/signal"

	"dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/doge"
	"dogecoin.org/fractal-engine/pkg/dogenet"
	"dogecoin.org/fractal-engine/pkg/rpc"
	"dogecoin.org/fractal-engine/pkg/store"
	"github.com/golang-migrate/migrate"
)

type tokenisationService struct {
	signalChan    chan os.Signal
	RpcServer     *rpc.RpcServer
	store         *store.TokenisationStore
	DogeNetClient *dogenet.DogeNetClient
	DogeClient    *doge.RpcClient
	StatusChan    chan string
	Follower      *doge.DogeFollower
}

func NewTokenisationService() *tokenisationService {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	cfg := config.NewConfig()

	store, err := store.NewTokenisationStore(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to create tokenisation store: %v", err)
	}

	follower := doge.NewFollower(cfg, store)

	return &tokenisationService{
		signalChan:    signalChan,
		RpcServer:     rpc.NewRpcServer(cfg, store),
		store:         store,
		DogeNetClient: dogenet.NewDogeNetClient(cfg),
		DogeClient:    doge.NewRpcClient(cfg),
		StatusChan:    make(chan string),
		Follower:      follower,
	}
}

func (s *tokenisationService) Start() {
	err := s.store.Migrate()
	if err != nil && err.Error() != migrate.ErrNoChange.Error() {
		log.Fatalf("Failed to migrate tokenisation store: %v", err)
	}

	go s.RpcServer.Start()
	go s.Follower.Start()

	s.StatusChan <- "Started"

	<-s.signalChan

	log.Println("Received interrupt signal, shutting down...")

	s.Stop()
}

func (s *tokenisationService) Stop() {
	err := s.store.Close()
	if err != nil {
		log.Fatalf("Failed to close tokenisation store: %v", err)
	}
	s.RpcServer.Stop()
}
