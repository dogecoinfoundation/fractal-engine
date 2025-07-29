package health

import (
	"log"
	"time"

	"dogecoin.org/fractal-engine/pkg/doge"
	"dogecoin.org/fractal-engine/pkg/store"
)

type HealthService struct {
	dogeClient *doge.RpcClient
	tokenStore *store.TokenisationStore
	running    bool
}

func NewHealthService(dogeClient *doge.RpcClient, tokenStore *store.TokenisationStore) *HealthService {
	return &HealthService{dogeClient: dogeClient, tokenStore: tokenStore, running: false}
}

func (h *HealthService) Start() {
	h.running = true

	for {
		bestBlockHash, err := h.dogeClient.GetBestBlockHash()
		if err != nil {
			log.Println("Error getting best block hash:", err)
			time.Sleep(10 * time.Second)
			continue
		}

		blockHeader, err := h.dogeClient.GetBlockHeader(bestBlockHash)
		if err != nil {
			log.Println("Error getting block header:", err)
			time.Sleep(10 * time.Second)
			continue
		}

		blockchainInfo, err := h.dogeClient.GetBlockchainInfo()
		if err != nil {
			log.Println("Error getting blockchain info:", err)
			time.Sleep(10 * time.Second)
			continue
		}
		chain := blockchainInfo.Chain

		latestBlockHeight := int(blockHeader.Height)
		currentBlockHeight, _, _, err := h.tokenStore.GetChainPosition()
		if err != nil {
			log.Println("Error getting chain position:", err)
			time.Sleep(10 * time.Second)
			continue
		}

		_, err = h.dogeClient.GetWalletInfo()

		err = h.tokenStore.UpsertHealth(int64(currentBlockHeight), int64(latestBlockHeight), chain, err == nil)
		if err != nil {
			log.Println("Error upserting health:", err)
		}

		time.Sleep(10 * time.Second)

		if !h.running {
			break
		}
	}
}

func (h *HealthService) Stop() {
	h.running = false
}
