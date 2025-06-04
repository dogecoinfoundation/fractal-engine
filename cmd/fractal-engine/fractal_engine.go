package main

import (
	"flag"
	"os"
	"os/signal"

	"dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/service"
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

	flag.StringVar(&rpcServerHost, "rpc-server-host", "0.0.0.0", "RPC Server Host")
	flag.StringVar(&rpcServerPort, "rpc-server-port", "8080", "RPC Server Port")
	flag.StringVar(&dogeNetNetwork, "doge-net-network", "tcp", "DogeNet Network")
	flag.StringVar(&dogeNetAddress, "doge-net-address", "0.0.0.0:8085", "DogeNet Address")
	flag.StringVar(&dogeScheme, "doge-scheme", "http", "Doge Scheme")
	flag.StringVar(&dogeHost, "doge-host", "0.0.0.0", "Doge Host")
	flag.StringVar(&dogePort, "doge-port", "22555", "Doge Port")
	flag.StringVar(&dogeUser, "doge-user", "test", "Doge User")
	flag.StringVar(&dogePassword, "doge-password", "test", "Doge Password")
	flag.StringVar(&databaseURL, "database-url", "sqlite://fractal-engine.db", "Database URL")
	flag.StringVar(&migrationsPath, "migrations-path", "db/migrations", "Migrations Path")
	flag.BoolVar(&persistFollower, "persist-follower", true, "Persist Follower")

	flag.Parse()

	cfg := &config.Config{
		RpcServerHost:   rpcServerHost,
		RpcServerPort:   rpcServerPort,
		DogeNetNetwork:  dogeNetNetwork,
		DogeNetAddress:  dogeNetAddress,
		DogeScheme:      dogeScheme,
		DogeHost:        dogeHost,
		DogePort:        dogePort,
		DogeUser:        dogeUser,
		DogePassword:    dogePassword,
		DatabaseURL:     databaseURL,
		PersistFollower: persistFollower,
		MigrationsPath:  migrationsPath,
	}

	service := service.NewTokenisationService(cfg)
	service.Start()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	<-signalChan

	service.Stop()
}
