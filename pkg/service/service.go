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
)

type tokenisationService struct {
	signalChan    chan os.Signal
	RpcServer     *rpc.RpcServer
	store         *store.TokenisationStore
	DogeNetClient *dogenet.DogeNetClient
	DogeClient    *doge.DogeClient
}

func NewTokenisationService() *tokenisationService {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	cfg := config.NewConfig()

	store := store.NewTokenisationStore(cfg)

	return &tokenisationService{
		signalChan:    signalChan,
		RpcServer:     rpc.NewRpcServer(cfg, store),
		store:         store,
		DogeNetClient: dogenet.NewDogeNetClient(cfg),
		DogeClient:    doge.NewDogeClient(cfg),
	}
}

func (s *tokenisationService) Start() {
	go s.RpcServer.Start()

	<-s.signalChan

	log.Println("Received interrupt signal, shutting down...")

	s.Stop()
}

func (s *tokenisationService) Stop() {
	s.store.Close()
	s.RpcServer.Stop()
}
