package doge

import (
	"context"
	"log"
	"time"

	"dogecoin.org/fractal-engine/pkg/store"
)

type OnChainProcessor struct {
	dbStore *store.Store
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewOnChainProcessor(dbStore *store.Store) *OnChainProcessor {
	ctx, cancel := context.WithCancel(context.Background())
	return &OnChainProcessor{
		dbStore: dbStore,
		ctx:     ctx,
		cancel:  cancel,
	}
}

func (p *OnChainProcessor) Start() error {
	return nil
}

func (p *OnChainProcessor) Process() error {
	ticker := time.NewTicker(5 * time.Second) // Wait 5 seconds between polls
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			// Context cancelled, exit
			return nil
		case <-ticker.C:
			// Do your database work
			err := p.checkForRecords()
			if err != nil {
				log.Println("Error processing onchain mints:", err)
			}
		}
	}
}

func (p *OnChainProcessor) checkForRecords() error {
	records, err := p.dbStore.GetUnverifiedOnchainMints()
	if err != nil {
		return err
	}

	for _, record := range records {
		log.Println("Processing onchain mint:", record.Id)
		mint, _ := p.dbStore.GetMint(record.Id)
		if mint == nil {
			log.Println("Mint not found:", record.Id)
			continue
		}

		err = p.dbStore.VerifyMint(mint.Id, mint.TransactionHash)
		if err != nil {
			log.Println("Error updating mint:", err)
			continue
		}

		err = p.dbStore.RemoveOnchainMint(mint.Id)
		if err != nil {
			log.Println("Error removing onchain mint:", err)
			continue
		}
	}

	return nil
}

func (p *OnChainProcessor) Stop() error {
	p.cancel()
	return nil
}
