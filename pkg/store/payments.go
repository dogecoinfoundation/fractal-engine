package store

import (
	"fmt"
	"log"

	"dogecoin.org/fractal-engine/pkg/protocol"
	"google.golang.org/protobuf/proto"
)

func (s *TokenisationStore) MatchPayment(onchainTransaction OnChainTransaction) error {
	if onchainTransaction.ActionType != protocol.ACTION_PAYMENT {
		return fmt.Errorf("action type is not payment: %d", onchainTransaction.ActionType)
	}

	var onchainMessage protocol.OnChainPaymentMessage
	err := proto.Unmarshal(onchainTransaction.ActionData, &onchainMessage)
	if err != nil {
		return err
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	rows, err := tx.Query("SELECT id, hash, payment_address, buy_offer_offerer_address, buy_offer_hash, buy_offer_mint_hash, buy_offer_quantity, buy_offer_price, buy_offer_value, created_at, sell_offer_address FROM invoices WHERE hash = $1", onchainMessage.Hash)
	if err != nil {
		log.Println("Error querying invoices:", err)
		return err
	}

	var invoice Invoice

	if rows.Next() {
		err := rows.Scan(&invoice.Id, &invoice.Hash, &invoice.PaymentAddress, &invoice.BuyOfferOffererAddress, &invoice.BuyOfferHash, &invoice.BuyOfferMintHash, &invoice.BuyOfferQuantity, &invoice.BuyOfferPrice, &invoice.BuyOfferValue, &invoice.CreatedAt, &invoice.SellOfferAddress)
		if err != nil {
			log.Println("Error scanning invoice:", err)
			return err
		}
	}

	rows.Close()

	if invoice.Id == "" {
		return fmt.Errorf("invoice not found")
	}

	if onchainTransaction.Value != invoice.BuyOfferValue {
		return fmt.Errorf("payment value is not equal to buy offer value: %f != %f", onchainTransaction.Value, invoice.BuyOfferValue)
	}

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

	pendingTokenBalance, err := s.GetPendingTokenBalance(invoice.Hash, invoice.BuyOfferMintHash, tx)
	if err != nil {
		log.Println("Error getting pending token balance:", err)
		return err
	}

	if pendingTokenBalance.Quantity != invoice.BuyOfferQuantity {
		return fmt.Errorf("pending token balance quantity is not equal to buy offer quantity: %d != %d", pendingTokenBalance.Quantity, invoice.BuyOfferQuantity)
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
