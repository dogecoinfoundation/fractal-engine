package store

import (
	"log"
	"time"
)

type TrimmerService struct {
	store   *TokenisationStore
	running bool
}

func NewTrimmerService(store *TokenisationStore) *TrimmerService {
	return &TrimmerService{store: store, running: false}
}

func (t *TrimmerService) Start() {
	t.running = true

	for {
		err := t.store.TrimOldUnconfirmedMints(100)
		if err != nil {
			log.Println("Error trimming unconfirmed mints:", err)
		}

		time.Sleep(1 * time.Minute)

		if !t.running {
			break
		}
	}
}

func (t *TrimmerService) Stop() {
	t.running = false
}
