package doge

import (
	"context"
	"log"
	"time"

	"dogecoin.org/fractal-engine/pkg/store"
)

type OnChainProcessor struct {
	dbStore *store.TokenisationStore
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewOnChainProcessor(dbStore *store.TokenisationStore) *OnChainProcessor {
	ctx, cancel := context.WithCancel(context.Background())
	return &OnChainProcessor{
		dbStore: dbStore,
		ctx:     ctx,
		cancel:  cancel,
	}
}

func (p *OnChainProcessor) Start(notify chan string) error {
	go p.Process(notify)

	return nil
}

func (p *OnChainProcessor) Process(notify chan string) error {
	ticker := time.NewTicker(5 * time.Second) // Wait 5 seconds between polls
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			// Context cancelled, exit
			return nil
		case <-ticker.C:
			// Do your database work
			err := p.checkForRecords(notify)
			if err != nil {
				log.Println("Error processing onchain mints:", err)
			}
		}
	}
}

func (p *OnChainProcessor) checkForRecords(notify chan string) error {
	records, err := p.dbStore.GetOnChainMints()
	if err != nil {
		return err
	}

	for _, record := range records {
		mint, err := p.dbStore.GetUnverifiedMint(record.MintId)
		if err != nil {
			log.Println("can't find mint:", err)
			continue
		}

		err = p.dbStore.VerifyMint(mint.Id, record.TransactionHash)
		if err != nil {
			log.Println("Error verifying mint:", err)
			continue
		}

		err = p.dbStore.DeleteOnChainMint(mint.Id)
		if err != nil {
			log.Println("Error removing onchain mint:", err)
			continue
		}

		notify <- "mint verified"
	}

	return nil
}

func (p *OnChainProcessor) Stop() error {
	p.cancel()
	return nil
}
