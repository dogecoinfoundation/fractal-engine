package service

import (
	"fmt"
	"log"

	"dogecoin.org/fractal-engine/pkg/doge"
	"dogecoin.org/fractal-engine/pkg/store"
)

const MIN_CONFIRMATIONS_REQUIRED = 6

type PaymentProcessor struct {
	store      *store.TokenisationStore
	dogeClient *doge.RpcClient
}

func NewPaymentProcessor(store *store.TokenisationStore, dogeClient *doge.RpcClient) *PaymentProcessor {
	return &PaymentProcessor{store: store, dogeClient: dogeClient}
}

func (p *PaymentProcessor) Process(tx store.OnChainTransaction) error {
	invoice, err := p.store.MatchPayment(tx)
	if err != nil {
		log.Println("Match Payment", err)
		return err
	}

	blockHeader, err := p.dogeClient.GetBlockHeader(tx.BlockHash)
	if err != nil {
		return err
	}

	if blockHeader.Confirmations < MIN_CONFIRMATIONS_REQUIRED {
		log.Println("Minimum confirmations not met:", err)
		return fmt.Errorf("Minimum confirmations not met: %d < %d", blockHeader.Confirmations, MIN_CONFIRMATIONS_REQUIRED)
	}

	err = p.store.ProcessPayment(tx, invoice)
	if err != nil {
		log.Println("ProcessPayment:", err)
		return err
	}

	log.Println("Matched payment:", tx.TxHash)
	return nil
}
