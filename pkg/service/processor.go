package service

import (
	"fmt"
	"log"
	"time"

	"dogecoin.org/fractal-engine/pkg/store"
)

type FractalEngineProcessor struct {
	store   *store.TokenisationStore
	Running bool
}

func NewFractalEngineProcessor(store *store.TokenisationStore) *FractalEngineProcessor {
	return &FractalEngineProcessor{store: store}
}

func (p *FractalEngineProcessor) Process() error {
	txs, err := p.store.GetOnChainTransactions(10)
	if err != nil {
		return err
	}

	for _, tx := range txs {
		fmt.Println("Processing transaction:", tx.TxHash)

		if p.store.MatchInvoice(tx) {
			continue
		}

		if p.store.MatchMint(tx) {
			continue
		}

		err = p.store.MatchUnconfirmedInvoice(tx)
		if err == nil {
			log.Println("Matched invoice:", tx.TxHash)
		}

		err = p.store.MatchUnconfirmedMint(tx)
		if err == nil {
			log.Println("Matched mint:", tx.TxHash)
		}
	}

	return nil
}

func (p *FractalEngineProcessor) Start() {
	p.Running = true

	for {
		if !p.Running {
			break
		}

		err := p.Process()
		if err != nil {
			log.Println("Error processing:", err)
		}

		time.Sleep(3 * time.Second)
	}
}

func (p *FractalEngineProcessor) Stop() {
	fmt.Println("Stopping processor")
	p.Running = false
}
