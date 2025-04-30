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
	records, err := p.dbStore.GetUnverifiedOnchainMints()
	if err != nil {
		return err
	}

	for _, record := range records {
		log.Println("Processing onchain mint:", record.Id)
		mint, _ := p.dbStore.GetMintForOutputAddress(record.Id, record.OutputAddress)
		if mint == nil {
			log.Println("Mint not found:", record.Id)
			continue
		}

		err = p.dbStore.VerifyMint(mint.Id, record.Hash)
		if err != nil {
			log.Println("Error updating mint:", err)
			continue
		}

		// Insert Account Record

		// id text primary key,
		// address VARCHAR(255) NOT NULL,
		// balance BIGINT NOT NULL DEFAULT 0,
		// mint_id text NOT NULL,
		// created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP

		err = p.dbStore.CreateAccountFromMint(mint)
		if err != nil {
			log.Println("Error creating account:", err)
			continue
		}

		err = p.dbStore.RemoveOnchainMint(mint.Id)
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
