package server

import (
	"errors"
	"fmt"
	"log"

	"dogecoin.org/chainfollower/pkg/chainfollower"
	"dogecoin.org/chainfollower/pkg/config"
	"dogecoin.org/chainfollower/pkg/messages"
	cfrpc "dogecoin.org/chainfollower/pkg/rpc"
	"dogecoin.org/chainfollower/pkg/state"
	"dogecoin.org/chainfollower/pkg/types"
	"dogecoin.org/fractal-engine/pkg/api"
	"dogecoin.org/fractal-engine/pkg/doge"
	"dogecoin.org/fractal-engine/pkg/protocol"
	"dogecoin.org/fractal-engine/pkg/store"
)

type FractalServer struct {
	config *config.Config
	Store  *store.Store
}

func NewFractalServer(cfg *config.Config) *FractalServer {
	if cfg == nil {
		localConfig, err := config.LoadConfig("config.toml")
		if err != nil {
			log.Fatal(err)
		}
		cfg = localConfig
	}

	dbStore, err := store.NewStore(cfg.DbUrl)
	if err != nil {
		log.Fatal(err)
	}

	return &FractalServer{
		config: cfg,
		Store:  dbStore,
	}
}

func (s *FractalServer) Start(status chan string) {
	rpcClient := cfrpc.NewRpcTransport(s.config)
	chainfollower := chainfollower.NewChainFollower(rpcClient)

	err := s.Store.Migrate()
	if err != nil {
		if err.Error() != "no change" {
			log.Fatal(err)
		}
	}

	apiServer := api.NewAPIServer(s.Store)
	go apiServer.Start()

	blockHeight, blockHash, err := s.Store.GetChainPosition()
	if err != nil {
		log.Fatal(err)
	}

	// Onchain Doge processor
	onchainProcessor := doge.NewOnChainProcessor(s.Store)
	go onchainProcessor.Start(status)

	fmt.Println("Starting chainfollower from block height:", blockHeight, "and block hash:", blockHash)

	messageChan := chainfollower.Start(&state.ChainPos{
		BlockHeight: blockHeight,
		BlockHash:   blockHash,
	})

	status <- "started"

	for message := range messageChan {
		switch msg := message.(type) {
		case messages.BlockMessage:

			for _, tx := range msg.Block.Tx {

				fractalMessage, err := GetFractalMessageFromVout(tx.VOut)
				if err != nil {
					log.Println("Error getting fractal message from vout:", err)
					continue
				}

				address, err := GetAddressFromVout(tx.VOut)
				if err != nil {
					log.Println("Error getting address from vout:", err)
					continue
				}

				switch fractalMessage.Action {
				case protocol.ACTION_MINT:
					status <- "processing mint"

					mint := protocol.Mint{}
					err = mint.Deserialize(fractalMessage.Data)
					if err != nil {
						log.Println("Error deserializing mint:", err)
						continue
					}

					fmt.Println("Mint VOUT:", mint.Id, tx.Hash, address)
					err = s.Store.CreateOnchainMint(mint, tx.Hash, address)
					if err != nil {
						log.Println("Error creating onchain mint:", err)
						continue
					}

					log.Println("Onchain mint created:", mint.Id)
				}
			}

			err := s.Store.UpsertChainPosition(msg.Block.Height, msg.Block.Hash)
			if err != nil {
				log.Println("Error setting chain position:", err)
			}

		case messages.RollbackMessage:
			log.Println("Received rollback message from chainfollower:")
			// log.Println(msg.OldChainPos)
			// log.Println(msg.NewChainPos)

			err := s.Store.UpsertChainPosition(msg.NewChainPos.BlockHeight, msg.NewChainPos.BlockHash)
			if err != nil {
				log.Println("Error setting chain position:", err)
			}

		default:
			log.Println("Received unknown message from chainfollower:")
		}
	}
}

func GetFractalMessageFromVout(vout []types.RawTxnVOut) (protocol.MessageEnvelope, error) {
	for _, vout := range vout {
		bytes := doge.ParseOpReturnData(vout)
		if bytes == nil {
			return protocol.MessageEnvelope{}, errors.New("no op return data")
		}

		message := protocol.MessageEnvelope{}
		err := message.Deserialize(bytes)
		if err != nil {
			log.Println("Error deserializing message envelope:", err)
			continue
		}

		if message.IsFractalEngineMessage() {
			return message, nil
		}

		return protocol.MessageEnvelope{}, errors.New("no fractal engine message")
	}

	return protocol.MessageEnvelope{}, errors.New("no fractal engine message")
}

func GetAddressFromVout(vout []types.RawTxnVOut) (string, error) {
	for _, vout := range vout {
		if vout.ScriptPubKey.Type == "pubkeyhash" {
			return vout.ScriptPubKey.Addresses[0], nil
		}
	}

	return "", errors.New("no address found")
}
