package main

import (
	"log"

	"dogecoin.org/chainfollower/pkg/chainfollower"
	"dogecoin.org/chainfollower/pkg/config"
	"dogecoin.org/chainfollower/pkg/messages"
	"dogecoin.org/chainfollower/pkg/rpc"
	"dogecoin.org/chainfollower/pkg/state"
	"dogecoin.org/fractal-engine/pkg/doge"
	"dogecoin.org/fractal-engine/pkg/protocol"
	"dogecoin.org/fractal-engine/pkg/store"
)

func main() {
	config, err := config.LoadConfig("config.toml")
	if err != nil {
		log.Fatal(err)
	}

	rpcClient := rpc.NewRpcTransport(config)
	chainfollower := chainfollower.NewChainFollower(rpcClient)

	dbStore, err := store.NewStore(config.DbUrl)
	if err != nil {
		log.Fatal(err)
	}

	err = dbStore.Migrate()
	if err != nil {
		log.Fatal(err)
	}

	messageChan := chainfollower.Start(&state.ChainPos{})

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

							log.Println("Received mint message from chainfollower:")
							log.Println("Title:", mint.Title)
							log.Println("Fraction Count:", mint.FractionCount)
							log.Println("Description:", mint.Description)
						}
					}
				}
			}
			// log.Println(msg.ChainPos)

		case messages.RollbackMessage:
			log.Println("Received rollback message from chainfollower:")
			// log.Println(msg.OldChainPos)
			// log.Println(msg.NewChainPos)

		default:
			log.Println("Received unknown message from chainfollower:")
		}
	}
}
