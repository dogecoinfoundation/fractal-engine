package dogenet

import (
	"context"
	"log"
	"time"

	"dogecoin.org/fractal-engine/pkg/store"
)

type DogenetService struct {
	dbStore       *store.TokenisationStore
	dogenetClient *DogeNetClient
	ctx           context.Context
	cancel        context.CancelFunc
}

func NewDogenetService(dbStore *store.TokenisationStore, dogenetClient *DogeNetClient) *DogenetService {
	ctx, cancel := context.WithCancel(context.Background())
	return &DogenetService{
		dbStore:       dbStore,
		dogenetClient: dogenetClient,
		ctx:           ctx,
		cancel:        cancel,
	}
}

func (p *DogenetService) Start() error {
	go p.Process()

	return nil
}

func (p *DogenetService) Process() error {
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
				log.Println("Error processing mints:", err)
			}
		}
	}
}

func (p *DogenetService) checkForRecords() error {
	records, err := p.dbStore.GetMintsForGossip(10)
	if err != nil {
		return err
	}

	for _, record := range records {
		err := p.dogenetClient.GossipMint(record)
		if err != nil {
			log.Println("Error gossiping mint:", err)
		}

		err = p.dbStore.UpdateMintGossiped(record.Id)
		if err != nil {
			log.Println("Error updating mint gossiped:", err)
		}
	}

	return nil
}

func (p *DogenetService) Stop() error {
	p.cancel()
	return nil
}
