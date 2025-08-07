package store

import (
	"fmt"
	"log"

	"dogecoin.org/fractal-engine/pkg/protocol"
	"google.golang.org/protobuf/proto"
)

func (s *TokenisationStore) ProcessPayment(onchainTransaction OnChainTransaction, invoice Invoice) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	_, err = tx.Exec("UPDATE invoices SET paid_at = now() WHERE id = $1", invoice.Id)
	if err != nil {
		log.Println("Error updating invoice:", err)
		return err
	}

	_, err = tx.Exec("DELETE FROM onchain_transactions WHERE id = $1", onchainTransaction.Id)
	if err != nil {
		log.Println("Error deleting onchain transaction:", err)
		return err
	}

	pendingTokenBalance, err := s.GetPendingTokenBalanceForQuantity(invoice.Hash, invoice.BuyOfferMintHash, invoice.BuyOfferQuantity, tx)
	if err != nil {
		log.Println("Error getting pending token balance:", err)
		return err
	}

	err = s.MovePendingToTokenBalance(pendingTokenBalance, invoice.BuyOfferOffererAddress, tx)
	if err != nil {
		log.Println("Error moving pending to token balance:", err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (s *TokenisationStore) MatchPayment(onchainTransaction OnChainTransaction) (Invoice, error) {
	if onchainTransaction.ActionType != protocol.ACTION_PAYMENT {
		return Invoice{}, fmt.Errorf("action type is not payment: %d", onchainTransaction.ActionType)
	}

	var onchainMessage protocol.OnChainPaymentMessage
	err := proto.Unmarshal(onchainTransaction.ActionData, &onchainMessage)
	if err != nil {
		return Invoice{}, err
	}

	rows, err := s.DB.Query("SELECT id, hash, payment_address, buy_offer_offerer_address, buy_offer_hash, buy_offer_mint_hash, buy_offer_quantity, buy_offer_price, buy_offer_value, created_at, sell_offer_address FROM invoices WHERE hash = $1", onchainMessage.Hash)
	if err != nil {
		log.Println("Error querying invoices:", err)
		return Invoice{}, err
	}

	var invoice Invoice

	if rows.Next() {
		err := rows.Scan(&invoice.Id, &invoice.Hash, &invoice.PaymentAddress, &invoice.BuyOfferOffererAddress, &invoice.BuyOfferHash, &invoice.BuyOfferMintHash, &invoice.BuyOfferQuantity, &invoice.BuyOfferPrice, &invoice.BuyOfferValue, &invoice.CreatedAt, &invoice.SellOfferAddress)
		if err != nil {
			log.Println("Error scanning invoice:", err)
			return Invoice{}, err
		}
	}

	rows.Close()

	if invoice.Id == "" {
		return Invoice{}, fmt.Errorf("invoice not found")
	}

	if onchainTransaction.Value != invoice.BuyOfferValue {
		return Invoice{}, fmt.Errorf("payment value is not equal to buy offer value: %f != %f", onchainTransaction.Value, invoice.BuyOfferValue)
	}

	return invoice, nil
}
