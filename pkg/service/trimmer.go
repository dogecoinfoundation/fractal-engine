package service

import (
	"fmt"
	"log"
	"time"

	"dogecoin.org/fractal-engine/pkg/doge"
	"dogecoin.org/fractal-engine/pkg/store"
)

type TrimmerService struct {
	blocksToKeep int
	store        *store.TokenisationStore
	dogeClient   *doge.RpcClient
	running      bool
}

func NewTrimmerService(blocksToKeep int, store *store.TokenisationStore, dogeClient *doge.RpcClient) *TrimmerService {
	return &TrimmerService{blocksToKeep: blocksToKeep, store: store, dogeClient: dogeClient, running: false}
}

func (t *TrimmerService) Start() {
	t.running = true

	for {
		bestBlockHash, err := t.dogeClient.GetBestBlockHash()
		if err != nil {
			log.Println("Error getting best block hash:", err)
			time.Sleep(10 * time.Second)
			continue
		}

		blockHeader, err := t.dogeClient.GetBlockHeader(bestBlockHash)
		if err != nil {
			log.Println("Error getting block header:", err)
			time.Sleep(10 * time.Second)
			continue
		}

		latestBlockHeight := int(blockHeader.Height)
		oldestBlockHeight := latestBlockHeight - t.blocksToKeep

		err = t.store.TrimOldUnconfirmedMints(oldestBlockHeight)
		if err != nil {
			log.Println("Error trimming unconfirmed mints:", err)
		}

		err = t.store.TrimOldOnChainTransactions(oldestBlockHeight)
		if err != nil {
			log.Println("Error trimming on chain transactions:", err)
		}

		time.Sleep(10 * time.Second)

		if !t.running {
			break
		}
	}
}

func (t *TrimmerService) Stop() {
	fmt.Println("Stopping trimmer service")
	t.running = false
}
