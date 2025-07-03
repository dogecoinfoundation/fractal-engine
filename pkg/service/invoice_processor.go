package service

import (
	"log"

	"dogecoin.org/fractal-engine/pkg/protocol"
	"dogecoin.org/fractal-engine/pkg/store"
	"google.golang.org/protobuf/proto"
)

type InvoiceProcessor struct {
	store *store.TokenisationStore
}

func NewInvoiceProcessor(store *store.TokenisationStore) *InvoiceProcessor {
	return &InvoiceProcessor{store: store}
}

/*
* Check if invoice already taken pending availability
* If so, then attempt to match to existing invoice or unconfirmed invoice

* If not, check if theres enough availability on token balances
* If so, create a pending token balance
* If not, remove onchain transaction (discard invoice)
 */
func (p *InvoiceProcessor) Process(tx store.OnChainTransaction) error {
	hasPendingTokenBalance, err := p.EnsurePendingTokenBalance(tx)
	if err != nil {
		return err
	}

	if !hasPendingTokenBalance {
		log.Println("Invoice discarded, not enough availability")

		return nil
	}

	if p.store.MatchInvoice(tx) {
		return nil
	}

	err = p.store.MatchUnconfirmedInvoice(tx)
	if err == nil {
		log.Println("Matched invoice:", tx.TxHash)
	} else {
		log.Println("Error matching unconfirmed invoice:", err)
	}

	return nil
}

func (p *InvoiceProcessor) EnsurePendingTokenBalance(tx store.OnChainTransaction) (bool, error) {
	invoice := protocol.OnChainInvoiceMessage{}
	err := proto.Unmarshal(tx.ActionData, &invoice)
	if err != nil {
		log.Println("Error unmarshalling invoice:", err)
	}

	pendingTokenBalance, _ := p.store.GetPendingTokenBalance(invoice.InvoiceHash, invoice.MintHash)
	if pendingTokenBalance.InvoiceHash != "" {
		log.Println("Pending token balance already exists")
		return true, nil
	}

	tokenBalance, err := p.store.GetTokenBalance(invoice.SellOfferAddress, invoice.MintHash)
	if err != nil {
		log.Println("Error getting token balance:", err)
		return false, err
	}

	pendingTokenBalanceTotal, err := p.store.GetPendingTokenBalanceTotalForMintAndOwner(invoice.MintHash, invoice.SellOfferAddress)
	if err != nil {
		log.Println("Error getting pending token balance total:", err)
		return false, err
	}

	tokenBalanceAvailable := tokenBalance - pendingTokenBalanceTotal

	if tokenBalanceAvailable >= int(invoice.Quantity) {
		log.Println("Token balance is enough")
		err = p.store.UpsertPendingTokenBalance(invoice.InvoiceHash, invoice.MintHash, int(invoice.Quantity), tx.Id, invoice.SellOfferAddress)
		if err != nil {
			log.Println("Error inserting pending token balance:", err)
		}

		// err = p.store.UpsertTokenBalance(invoice.SellOfferAddress, invoice.MintHash, tokenBalance-int(invoice.Quantity))
		// if err != nil {
		// 	log.Println("Error upserting token balance:", err)
		// }

		return true, nil
	} else {
		log.Println("Token balance is not enough")

		err = p.store.RemoveOnChainTransaction(tx.Id)
		if err != nil {
			log.Println("Error removing onchain transaction:", err)
		}

		return false, nil
	}
}
