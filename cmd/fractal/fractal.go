package main

import (
	"log"

	"dogecoin.org/chainfollower/pkg/chainfollower"
	"dogecoin.org/chainfollower/pkg/config"
	"dogecoin.org/chainfollower/pkg/messages"
	cfrpc "dogecoin.org/chainfollower/pkg/rpc"
	"dogecoin.org/chainfollower/pkg/state"
	"dogecoin.org/fractal-engine/pkg/api"
	"dogecoin.org/fractal-engine/pkg/doge"
	"dogecoin.org/fractal-engine/pkg/protocol"
	"dogecoin.org/fractal-engine/pkg/store"
)

func main() {
	config, err := config.LoadConfig("config.toml")
	if err != nil {
		log.Fatal(err)
	}

	rpcClient := cfrpc.NewRpcTransport(config)
	chainfollower := chainfollower.NewChainFollower(rpcClient)

	dbStore, err := store.NewStore(config.DbUrl)
	if err != nil {
		log.Fatal(err)
	}

	err = dbStore.Migrate()
	if err != nil {
		if err.Error() != "no change" {
			log.Fatal(err)
		}
	}

	dogeClient := doge.NewDogeClient(rpcClient)
	apiServer := api.NewAPIServer(dbStore, dogeClient)
	go apiServer.Start()

	blockHeight, blockHash, err := dbStore.GetChainPosition()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Starting chainfollower from block height:", blockHeight, "and block hash:", blockHash)

	messageChan := chainfollower.Start(&state.ChainPos{
		BlockHeight: blockHeight,
		BlockHash:   blockHash,
	})

	for message := range messageChan {
		switch msg := message.(type) {
		case messages.BlockMessage:

			for _, tx := range msg.Block.Tx {
				for _, vout := range tx.VOut {
					bytes := doge.ParseOpReturnData(vout)
					if bytes == nil {
						continue
					}

					message := protocol.MessageEnvelope{}
					err := message.Deserialize(bytes)
					if err != nil {
						log.Println("Error deserializing message envelope:", err)
						continue
					}

					if message.IsFractalEngineMessage() {
						switch message.Action {
						case protocol.ACTION_MINT:
							mint := protocol.Mint{}
							err = mint.Deserialize(message.Data)
							if err != nil {
								log.Println("Error deserializing mint:", err)
								continue
							}

							existingMint, err := dbStore.GetMint(mint.Id)
							if err != nil {
								log.Println("Error getting mint:", err)
								continue
							}

							if existingMint != nil {
								// TODO VALIDATE

								err = dbStore.VerifyMint(mint.Id)
								if err != nil {
									log.Println("Error verifying mint:", err)
									continue
								}

								log.Println("Mint verified:", mint.Id)
							}
						}
					}
				}
			}

			err := dbStore.UpsertChainPosition(msg.Block.Height, msg.Block.Hash)
			if err != nil {
				log.Println("Error setting chain position:", err)
			}

		case messages.RollbackMessage:
			log.Println("Received rollback message from chainfollower:")
			// log.Println(msg.OldChainPos)
			// log.Println(msg.NewChainPos)

			err := dbStore.UpsertChainPosition(msg.NewChainPos.BlockHeight, msg.NewChainPos.BlockHash)
			if err != nil {
				log.Println("Error setting chain position:", err)
			}

		default:
			log.Println("Received unknown message from chainfollower:")
		}
	}
}
