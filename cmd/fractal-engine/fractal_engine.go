package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"code.dogecoin.org/gossip/dnet"
	"dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/dogenet"
	"dogecoin.org/fractal-engine/pkg/service"
	"dogecoin.org/fractal-engine/pkg/store"
)

func main() {
	var rpcServerHost string
	var rpcServerPort string
	var dogeNetNetwork string
	var dogeNetAddress string
	var dogeScheme string
	var dogeHost string
	var dogePort string
	var dogeUser string
	var dogePassword string
	var databaseURL string
	var persistFollower bool
	var migrationsPath string
	var rateLimitPerSecond int
	var invoiceLimit int
	var buyOfferLimit int
	var sellOfferLimit int

	flag.StringVar(&rpcServerHost, "rpc-server-host", "0.0.0.0", "RPC Server Host")
	flag.StringVar(&rpcServerPort, "rpc-server-port", "8080", "RPC Server Port")
	flag.StringVar(&dogeNetNetwork, "doge-net-network", "unix", "DogeNet Network")
	flag.StringVar(&dogeNetAddress, "doge-net-address", "/tmp/dogenet.sock", "DogeNet Address")
	flag.StringVar(&dogeScheme, "doge-scheme", "http", "Doge Scheme")
	flag.StringVar(&dogeHost, "doge-host", "0.0.0.0", "Doge Host")
	flag.StringVar(&dogePort, "doge-port", "22556", "Doge Port")
	flag.StringVar(&dogeUser, "doge-user", "test", "Doge User")
	flag.StringVar(&dogePassword, "doge-password", "test", "Doge Password")
	flag.StringVar(&databaseURL, "database-url", "sqlite://fractal-engine.db", "Database URL")
	flag.StringVar(&migrationsPath, "migrations-path", "db/migrations", "Migrations Path")
	flag.BoolVar(&persistFollower, "persist-follower", true, "Persist Follower")
	flag.IntVar(&rateLimitPerSecond, "api-rate-limit-per-second", 10, "API Rate Limit Per Second")
	flag.IntVar(&invoiceLimit, "invoice-limit", 100, "Invoice Limit (per mint)")
	flag.IntVar(&buyOfferLimit, "buy-offer-limit", 3, "Buy Offer Limit (per buyer per mint)")
	flag.IntVar(&sellOfferLimit, "sell-offer-limit", 3, "Sell Offer Limit (per seller per mint)")

	flag.Parse()

	cfg := &config.Config{
		RpcServerHost:      rpcServerHost,
		RpcServerPort:      rpcServerPort,
		DogeNetNetwork:     dogeNetNetwork,
		DogeNetAddress:     dogeNetAddress,
		DogeScheme:         dogeScheme,
		DogeHost:           dogeHost,
		DogePort:           dogePort,
		DogeUser:           dogeUser,
		DogePassword:       dogePassword,
		DatabaseURL:        databaseURL,
		PersistFollower:    persistFollower,
		MigrationsPath:     migrationsPath,
		RateLimitPerSecond: rateLimitPerSecond,
		InvoiceLimit:       invoiceLimit,
		BuyOfferLimit:      buyOfferLimit,
		SellOfferLimit:     sellOfferLimit,
	}

	tokenStore, err := store.NewTokenisationStore(cfg.DatabaseURL, *cfg)
	if err != nil {
		log.Fatalf("Failed to create tokenisation store: %v", err)
	}

	kp, err := dnet.GenerateKeyPair()
	if err != nil {
		log.Fatalf("Failed to generate key pair: %v", err)
	}

	cfg.DogeNetKeyPair = kp

	dogenetClient := dogenet.NewDogeNetClient(cfg, tokenStore)

	service := service.NewTokenisationService(cfg, dogenetClient, tokenStore)
	service.Start()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	<-signalChan

	service.Stop()
}
