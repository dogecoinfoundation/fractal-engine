package main

import (
	"fmt"
	"log"

	"dogecoin.org/chainfollower/pkg/config"
	"dogecoin.org/chainfollower/pkg/rpc"
	"dogecoin.org/fractal-engine/pkg/doge"
	"dogecoin.org/fractal-engine/pkg/protocol"
	"dogecoin.org/fractal-engine/pkg/store"
)

func main() {
	appConfig, err := config.LoadConfig("config.toml")
	if err != nil {
		log.Fatal(err)
	}

	rpc := rpc.NewRpcTransport(&config.Config{
		RpcUrl: "http://your_username:your_password@localhost:18332",
	})

	ownerAddress := "mzd1LknowaPDaQWu9e9DBUF2GgDpi2tgnE"
	privKey := "cUCJ7pQkNsz41phQfYJ5XhqPBp8JZbLat5AdvZ9TGGuFo1egXAq7"

	dogeClient := doge.NewDogeClient(rpc)

	unspent, err := dogeClient.GetUnspent(ownerAddress)
	if err != nil {
		log.Fatal(err)
	}

	if len(unspent) == 0 {
		log.Fatal("No unspent UTXOs found")
	}

	dbStore, err := store.NewStore(appConfig.DbUrl)
	if err != nil {
		log.Fatal(err)
	}

	err = dbStore.Migrate()
	if err != nil {
		if err.Error() != "no change" {
			log.Fatal(err)
		}
	}

	newMint := &protocol.MintWithoutID{
		Title:         "TEST",
		FractionCount: 8,
		Description:   "TEST",
		Tags:          []string{"TEST"},
		Metadata:      nil,
	}

	id, err := dbStore.SaveMint(newMint)
	if err != nil {
		log.Fatal(err)
	}

	newMintRequest := &protocol.Mint{
		Id:            id,
		MintWithoutID: *newMint,
	}

	txId, err := dogeClient.CreateMint(newMintRequest, privKey, unspent, ownerAddress)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Mint created with txId:", txId)
}

/***

	Presign Trxn + Commitment Hash

	Buyer signs Presigned Trxn


***/
