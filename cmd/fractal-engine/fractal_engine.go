package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"strconv"
	"time"

	dn "code.dogecoin.org/dogenet/pkg/dogenet"
	"code.dogecoin.org/dogenet/pkg/spec"
	"code.dogecoin.org/gossip/dnet"
	"code.dogecoin.org/governor"
	"dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/dogenet"
	"dogecoin.org/fractal-engine/pkg/service"
	"dogecoin.org/fractal-engine/pkg/store"
	"dogecoin.org/fractal-engine/pkg/version"
)

func main() {
	var rpcServerHost string
	var rpcServerPort string
	var dogeNetNetwork string
	var dogeNetAddress string
	var dogeNetWebAddress string
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
	var embedDogenet bool
	var corsAllowedOrigins string
	var showVersion bool

	flag.StringVar(&rpcServerHost, "rpc-server-host", "0.0.0.0", "RPC Server Host")
	flag.StringVar(&rpcServerPort, "rpc-server-port", "8891", "RPC Server Port")
	flag.StringVar(&dogeNetNetwork, "doge-net-network", "tcp", "DogeNet Network")
	flag.StringVar(&dogeNetAddress, "doge-net-address", "0.0.0.0:8086", "DogeNet Address")
	flag.StringVar(&dogeNetWebAddress, "doge-net-web-address", "0.0.0.0:8085", "DogeNet Web Address")
	flag.BoolVar(&embedDogenet, "embed-dogenet", false, "Embed the DogeNet service")
	flag.StringVar(&dogeScheme, "doge-scheme", "http", "Doge Scheme")
	flag.StringVar(&dogeHost, "doge-host", "0.0.0.0", "Doge Host")
	flag.StringVar(&dogePort, "doge-port", "22556", "Doge Port")
	flag.StringVar(&dogeUser, "doge-user", "test", "Doge User")
	flag.StringVar(&dogePassword, "doge-password", "test", "Doge Password")
	flag.StringVar(&databaseURL, "database-url", "postgres://fractalstore:fractalstore@localhost:5432/fractalstore?sslmode=disable", "Database URL")
	flag.StringVar(&migrationsPath, "migrations-path", "db/migrations", "Migrations Path")
	flag.BoolVar(&persistFollower, "persist-follower", true, "Persist Follower")
	flag.IntVar(&rateLimitPerSecond, "api-rate-limit-per-second", 10, "API Rate Limit Per Second")
	flag.IntVar(&invoiceLimit, "invoice-limit", 100, "Invoice Limit (per mint)")
	flag.IntVar(&buyOfferLimit, "buy-offer-limit", 3, "Buy Offer Limit (per buyer per mint)")
	flag.IntVar(&sellOfferLimit, "sell-offer-limit", 3, "Sell Offer Limit (per seller per mint)")
	flag.StringVar(&corsAllowedOrigins, "cors-allowed-origins", "*", "Comma-separated list of allowed CORS origins or *")
	flag.BoolVar(&showVersion, "version", false, "Print version and exit")

	flag.Parse()

	if showVersion {
		log.Printf("fractal-engine %s\n", version.String())
		return
	}

	cfg := &config.Config{
		RpcServerHost:      rpcServerHost,
		RpcServerPort:      rpcServerPort,
		DogeNetNetwork:     dogeNetNetwork,
		DogeNetAddress:     dogeNetAddress,
		DogeNetWebAddress:  dogeNetWebAddress,
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
		CORSAllowedOrigins: corsAllowedOrigins,
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

	gov := governor.New().CatchSignals()

	dogenetClient := dogenet.NewDogeNetClient(cfg, tokenStore)

	if embedDogenet {
		const WebAPIDefaultPort = 8085
		const DBFile = "dogenet.db"
		const DefaultStorage = "./storage"

		dogeNetServerKp, err := dnet.GenerateKeyPair()
		if err != nil {
			log.Fatalf("Failed to generate key pair: %v", err)
		}

		fmt.Printf("DogeNet PubKey is: %s\n", hex.EncodeToString(dogeNetServerKp.Pub[:]))

		var HandlerDefaultBind = spec.BindTo{Network: dogeNetNetwork, Address: dogeNetAddress} // const

		webAddy, err := dnet.ParseAddress(dogeNetWebAddress)
		if err != nil {
			log.Fatalf("Failed to parse web address: %v", err)
		}

		rawAddy := "0.0.0.0:" + strconv.Itoa(int(webAddy.Port)+33)

		addy, err := dnet.ParseAddress(rawAddy)
		if err != nil {
			log.Fatalf("Failed to parse web address: %v", err)
		}

		err = dn.DogeNet(gov, dn.DogeNetConfig{
			Dir:          DefaultStorage,
			DBFile:       DBFile,
			Binds:        []dnet.Address{addy},
			BindWeb:      []dnet.Address{webAddy},
			HandlerBind:  HandlerDefaultBind,
			NodeKey:      dogeNetServerKp,
			AllowLocal:   true,
			Public:       addy,
			UseReflector: false,
		})

		if err != nil {
			log.Fatal("Failed to setup DogeNet")
		}
	}

	gov.Start()

	if embedDogenet {
		for {
			active, err := dogenetClient.UnixSockActive()
			if active || err != nil {
				break
			}

			time.Sleep(200 * time.Millisecond)
		}
	}

	go dogenetClient.Run()

	service := service.NewTokenisationService(cfg, dogenetClient, tokenStore)
	service.Run()

	gov.WaitForShutdown()
}
