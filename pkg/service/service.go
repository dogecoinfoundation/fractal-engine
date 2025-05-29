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

type tokenisationService struct {
	RpcServer      *rpc.RpcServer
	store          *store.TokenisationStore
	DogeNetClient  *dogenet.DogeNetClient
	DogeClient     *doge.RpcClient
	StatusChan     chan string
	Follower       *doge.DogeFollower
	ChainProcessor *doge.OnChainProcessor
}

func NewTokenisationService(cfg *config.Config) *tokenisationService {
	store, err := store.NewTokenisationStore(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to create tokenisation store: %v", err)
	}

	follower := doge.NewFollower(cfg, store)
	chainProcessor := doge.NewOnChainProcessor(store)

	return &tokenisationService{
		RpcServer:      rpc.NewRpcServer(cfg, store),
		store:          store,
		DogeNetClient:  dogenet.NewDogeNetClient(cfg),
		DogeClient:     doge.NewRpcClient(cfg),
		StatusChan:     make(chan string),
		Follower:       follower,
		ChainProcessor: chainProcessor,
	}
}

func (s *tokenisationService) Start() {
	err := s.store.Migrate()
	if err != nil && err.Error() != migrate.ErrNoChange.Error() {
		log.Fatalf("Failed to migrate tokenisation store: %v", err)
	}

	go s.RpcServer.Start()
	go s.Follower.Start()
	go s.ChainProcessor.Start(s.StatusChan)
}

func (s *tokenisationService) waitForFollower() {
	for {
		if !s.Follower.Running {
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}
}

func (s *tokenisationService) waitForRpc() {
	for {
		if !s.RpcServer.Running {
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}
}

func (s *tokenisationService) WaitForRunning() {
	s.waitForFollower()
	s.waitForRpc()
}

func (s *tokenisationService) Stop() {
	err := s.store.Close()
	if err != nil {
		log.Fatalf("Failed to close tokenisation store: %v", err)
	}
	s.RpcServer.Stop()
}
