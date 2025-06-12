package service

import (
	"log"
	"time"

	"dogecoin.org/fractal-engine/pkg/protocol"
	"dogecoin.org/fractal-engine/pkg/store"
)

type MatcherService struct {
	store   *store.TokenisationStore
	Running bool
}

func NewMatcherService(store *store.TokenisationStore) *MatcherService {
	return &MatcherService{store: store, Running: false}
}

func (s *MatcherService) Start() {
	s.Running = true

	for {
		if !s.Running {
			break
		}

		s.Process()

		time.Sleep(5 * time.Second)
	}
}

func (s *MatcherService) Process() {
	onChainTransactions, err := s.store.GetOnChainTransactions(10)
	if err != nil {
		log.Println("Error getting on chain transactions:", err)
		return
	}

	for _, transaction := range onChainTransactions {
		if transaction.ActionType == protocol.ACTION_MINT {
			err := s.store.MatchUnconfirmedMint(transaction)
			if err != nil {
				log.Println("Error matching unconfirmed mint:", err)
			}
		}
	}
}

func (s *MatcherService) Stop() {
	s.Running = false
}
