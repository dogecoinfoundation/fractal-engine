package doge

import (
	"encoding/hex"
	"errors"
	"log"
	"strings"

	fecfg "dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/protocol"
	"dogecoin.org/fractal-engine/pkg/store"
	"github.com/dogecoinfoundation/chainfollower/pkg/chainfollower"
	"github.com/dogecoinfoundation/chainfollower/pkg/config"
	"github.com/dogecoinfoundation/chainfollower/pkg/messages"
	"github.com/dogecoinfoundation/chainfollower/pkg/rpc"
	"github.com/dogecoinfoundation/chainfollower/pkg/state"
	"github.com/dogecoinfoundation/chainfollower/pkg/types"
)

type DogeFollower struct {
	cfg           *fecfg.Config
	store         *store.TokenisationStore
	chainfollower *chainfollower.ChainFollower
	Running       bool
}

func NewFollower(cfg *fecfg.Config, store *store.TokenisationStore) *DogeFollower {
	rpcClient := rpc.NewRpcTransport(&config.Config{
		RpcUrl:  cfg.DogeScheme + "://" + cfg.DogeHost + ":" + cfg.DogePort,
		RpcUser: cfg.DogeUser,
		RpcPass: cfg.DogePassword,
	})

	chainfollower := chainfollower.NewChainFollower(rpcClient)

	return &DogeFollower{cfg: cfg, store: store, chainfollower: chainfollower, Running: false}
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

	f.Running = true

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
					mint := protocol.MintTransaction{}
					err = mint.Deserialize(fractalMessage.Data)
					if err != nil {
						log.Println("Error deserializing mint:", err)
						continue
					}

					err = f.store.SaveOnChainMint(mint.MintID, address, tx.Hash)
					if err != nil {
						log.Println("Error saving on chain mint:", err)
						continue
					}

					log.Println("Onchain mint created:", mint.MintID)
				}
			}

			if f.cfg.PersistFollower {
				err := f.store.UpsertChainPosition(msg.Block.Height, msg.Block.Hash)
				if err != nil {
					log.Println("Error setting chain position:", err)
				}
			}

		case messages.RollbackMessage:
			log.Println("Received rollback message from chainfollower:")
			// log.Println(msg.OldChainPos)
			// log.Println(msg.NewChainPos)

			if f.cfg.PersistFollower {
				err := f.store.UpsertChainPosition(msg.NewChainPos.BlockHeight, msg.NewChainPos.BlockHash)
				if err != nil {
					log.Println("Error setting chain position:", err)
				}
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
