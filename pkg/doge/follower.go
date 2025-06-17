package doge

import (
	"encoding/hex"
	"errors"
	"fmt"
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
	f.Running = true

	blockHeight, blockHash, _, err := f.store.GetChainPosition()
	if err != nil {
		return err
	}

	msgChan := f.chainfollower.Start(&state.ChainPos{
		BlockHash:   blockHash,
		BlockHeight: blockHeight,
	})

	for msg := range msgChan {
		switch msg := msg.(type) {
		case messages.BlockMessage:
			for _, tx := range msg.Block.Tx {
				fractalMessage, err := GetFractalMessageFromVout(tx.VOut)
				if err != nil {
					continue
				}

				err = f.store.SaveOnChainTransaction(tx.Hash, msg.Block.Height, fractalMessage.Action, fractalMessage.Version, fractalMessage.Data)
				if err != nil {
					log.Println("Error saving on chain transaction:", err)
				}
			}

			if f.cfg.PersistFollower {
				err := f.store.UpsertChainPosition(msg.ChainPos.BlockHeight, msg.ChainPos.BlockHash, msg.ChainPos.WaitingForNextHash)
				if err != nil {
					log.Println("Error setting chain position:", err)
				}
			}

		case messages.RollbackMessage:
			log.Println("Received rollback message from chainfollower:")
			if f.cfg.PersistFollower {
				err := f.store.UpsertChainPosition(msg.NewChainPos.BlockHeight, msg.NewChainPos.BlockHash, msg.NewChainPos.WaitingForNextHash)
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

func GetFractalMessageFromVout(vout []types.RawTxnVOut) (protocol.MessageEnvelope, error) {
	for _, vout := range vout {

		fmt.Println("vout:", vout)

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
