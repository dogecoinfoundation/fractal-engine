package service

import (
	"log"
	"time"

	"dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/doge"
	"dogecoin.org/fractal-engine/pkg/dogenet"
	"dogecoin.org/fractal-engine/pkg/rpc"
	"dogecoin.org/fractal-engine/pkg/store"
	"github.com/golang-migrate/migrate"
)

type TokenisationService struct {
	RpcServer      *rpc.RpcServer
	Store          *store.TokenisationStore
	DogeNetClient  *dogenet.DogeNetClient
	DogeClient     *doge.RpcClient
	Follower       *doge.DogeFollower
	ChainProcessor *doge.OnChainProcessor
	DogenetService *dogenet.DogenetService
}

func NewTokenisationService(cfg *config.Config) *TokenisationService {
	store, err := store.NewTokenisationStore(cfg.DatabaseURL, *cfg)
	if err != nil {
		log.Fatalf("Failed to create tokenisation store: %v", err)
	}

	dogenetClient := dogenet.NewDogeNetClient(cfg)
	follower := doge.NewFollower(cfg, store)
	chainProcessor := doge.NewOnChainProcessor(store)
	dogenetService := dogenet.NewDogenetService(store, dogenetClient)

	return &TokenisationService{
		RpcServer:      rpc.NewRpcServer(cfg, store),
		Store:          store,
		DogeNetClient:  dogenetClient,
		DogeClient:     doge.NewRpcClient(cfg),
		Follower:       follower,
		ChainProcessor: chainProcessor,
		DogenetService: dogenetService,
	}
}

func (s *TokenisationService) Start() {
	err := s.Store.Migrate()
	if err != nil && err.Error() != migrate.ErrNoChange.Error() {
		log.Fatalf("Failed to migrate tokenisation store: %v", err)
	}

	go s.RpcServer.Start()
	go s.Follower.Start()
	go s.ChainProcessor.Start()
	go s.DogenetService.Start()
}

func (s *TokenisationService) waitForFollower() {
	for {
		if !s.Follower.Running {
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}
}

func (s *TokenisationService) waitForRpc() {
	for {
		if !s.RpcServer.Running {
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}
}

func (s *TokenisationService) WaitForRunning() {
	s.waitForFollower()
	s.waitForRpc()
}

func (s *TokenisationService) Stop() {
	err := s.Store.Close()
	if err != nil {
		log.Fatalf("Failed to close tokenisation store: %v", err)
	}
	s.RpcServer.Stop()
}
