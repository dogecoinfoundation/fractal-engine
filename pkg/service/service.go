package service

import (
	"fmt"
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
	TrimmerService *TrimmerService
	Processor      *FractalEngineProcessor
}

func NewTokenisationService(cfg *config.Config, dogenetClient *dogenet.DogeNetClient, tokenStore *store.TokenisationStore) *TokenisationService {
	dogeClient := doge.NewRpcClient(cfg)
	follower := doge.NewFollower(cfg, tokenStore)

	trimmerService := NewTrimmerService(100, tokenStore, dogeClient)
	processor := NewFractalEngineProcessor(tokenStore)

	return &TokenisationService{
		RpcServer:      rpc.NewRpcServer(cfg, tokenStore, dogenetClient),
		Store:          tokenStore,
		DogeNetClient:  dogenetClient,
		DogeClient:     dogeClient,
		Follower:       follower,
		TrimmerService: trimmerService,
		Processor:      processor,
	}
}

func (s *TokenisationService) Start() {
	err := s.Store.Migrate()

	if err != nil && err.Error() != migrate.ErrNoChange.Error() {
		log.Fatalf("Failed to migrate tokenisation store: %v", err)
	}

	statusChan := make(chan string)

	go s.DogeNetClient.Start(statusChan)

	<-statusChan

	go s.RpcServer.Start()
	go s.Follower.Start()
	go s.TrimmerService.Start()
	go s.Processor.Start()
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
	fmt.Println("Waiting for follower")
	s.waitForFollower()
	fmt.Println("Waiting for rpc")
	s.waitForRpc()
}

func (s *TokenisationService) Stop() {
	s.Processor.Stop()
	s.Follower.Stop()
	s.Store.Close()
	s.RpcServer.Stop()
	s.TrimmerService.Stop()
	s.DogeNetClient.Stop()
}
