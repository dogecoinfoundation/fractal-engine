package doge

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"

	"dogecoin.org/chainfollower/pkg/chainfollower"
	"dogecoin.org/chainfollower/pkg/config"
	"dogecoin.org/chainfollower/pkg/messages"
	"dogecoin.org/chainfollower/pkg/rpc"
	"dogecoin.org/chainfollower/pkg/state"
	"dogecoin.org/chainfollower/pkg/types"
	fecfg "dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/protocol"
	"dogecoin.org/fractal-engine/pkg/store"
)

type DogeFollower struct {
	cfg           *fecfg.Config
	store         *store.TokenisationStore
	chainfollower *chainfollower.ChainFollower
}

func NewFollower(cfg *fecfg.Config, store *store.TokenisationStore) *DogeFollower {
	rpcClient := rpc.NewRpcTransport(&config.Config{
		RpcUrl:  cfg.DogeScheme + "://" + cfg.DogeHost + ":" + cfg.DogePort,
		RpcUser: cfg.DogeUser,
		RpcPass: cfg.DogePassword,
	})

	chainfollower := chainfollower.NewChainFollower(rpcClient)

	return &DogeFollower{cfg: cfg, store: store, chainfollower: chainfollower}
}

func (f *DogeFollower) Start() error {
	blockHeight, blockHash, err := f.store.GetChainPosition()
	if err != nil {
		log.Fatal(err)
	}

	msgChan := f.chainfollower.Start(&state.ChainPos{
		BlockHeight: blockHeight,
		BlockHash:   blockHash,
	})

	for message := range msgChan {
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
					err = fractalMessage.Deserialize(string(fractalMessage.Data))
					if err != nil {
						log.Println("Error deserializing mint:", err)
						continue
					}

					fmt.Printf("Mint VOUT: %s, %v\n", address, fractalMessage)

					// fmt.Println("Mint VOUT:", mint.Id, tx.Hash, address)
					// err = s.Store.CreateOnchainMint(mint, tx.Hash, address)
					// if err != nil {
					// 	log.Println("Error creating onchain mint:", err)
					// 	continue
					// }

					// log.Println("Onchain mint created:", mint.Id)
				}
			}

			err := f.store.UpsertChainPosition(msg.Block.Height, msg.Block.Hash)
			if err != nil {
				log.Println("Error setting chain position:", err)
			}

		case messages.RollbackMessage:
			log.Println("Received rollback message from chainfollower:")
			// log.Println(msg.OldChainPos)
			// log.Println(msg.NewChainPos)

			err := f.store.UpsertChainPosition(msg.NewChainPos.BlockHeight, msg.NewChainPos.BlockHash)
			if err != nil {
				log.Println("Error setting chain position:", err)
			}

		default:
			log.Println("Received unknown message from chainfollower:")
		}
	}
	return nil
}

func (f *DogeFollower) Stop() error {
	f.chainfollower.Stop()

	return nil
}

func GetFractalMessageFromVout(vout []types.RawTxnVOut) (protocol.MessageEnvelope, error) {
	for _, vout := range vout {
		bytes := ParseOpReturnData(vout)
		if bytes == nil {
			return protocol.MessageEnvelope{}, errors.New("no op return data")
		}

		message := protocol.MessageEnvelope{}
		err := message.Deserialize(string(bytes))
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

func ParseOpReturnData(vout types.RawTxnVOut) []byte {
	asm := vout.ScriptPubKey.Asm
	parts := strings.Split(asm, " ")
	if len(parts) > 0 {
		op := parts[0]
		if op == "OP_RETURN" {
			bytes, err := hex.DecodeString(parts[1])
			if err != nil {
				return nil
			}

			return bytes
		}
	}
	return nil
}
