package service

import (
	"log"

	"dogecoin.org/fractal-engine/pkg/store"
)

type PaymentProcessor struct {
	store *store.TokenisationStore
}

func NewPaymentProcessor(store *store.TokenisationStore) *PaymentProcessor {
	return &PaymentProcessor{store: store}
}

func (p *PaymentProcessor) Process(tx store.OnChainTransaction) error {
	err := p.store.MatchPayment(tx)
	if err == nil {
		log.Println("Matched payment:", tx.TxHash)
		return nil
	}

	return err
}
