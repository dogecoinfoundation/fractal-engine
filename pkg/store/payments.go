package store

import (
	"fmt"
	"time"

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

	rows, err := tx.Query("SELECT id, hash, payment_address, buy_offer_offerer_address, buy_offer_hash, buy_offer_mint_hash, buy_offer_quantity, buy_offer_price, buy_offer_value, created_at, sell_offer_address FROM invoices WHERE hash = $1", onchainMessage.Hash)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		var invoice Invoice
		err := rows.Scan(&invoice.Id, &invoice.Hash, &invoice.PaymentAddress, &invoice.BuyOfferOffererAddress, &invoice.BuyOfferHash, &invoice.BuyOfferMintHash, &invoice.BuyOfferQuantity, &invoice.BuyOfferPrice, &invoice.BuyOfferValue, &invoice.CreatedAt, &invoice.SellOfferAddress)
		if err != nil {
			return err
		}

		if onchainTransaction.Value != invoice.BuyOfferValue {
			return fmt.Errorf("payment value is not equal to buy offer value: %f != %f", onchainTransaction.Value, invoice.BuyOfferValue)
		}

		_, err = tx.Exec("UPDATE invoices SET paid_at = $1 WHERE id = $2", time.Now(), invoice.Id)
		if err != nil {
			return err
		}

		_, err = tx.Exec("DELETE FROM onchain_transactions WHERE id = $1", onchainTransaction.Id)
		if err != nil {
			return err
		}

		pendingTokenBalance, err := s.GetPendingTokenBalance(invoice.Hash, invoice.BuyOfferMintHash)
		if err != nil {
			return err
		}

		if pendingTokenBalance.Quantity != invoice.BuyOfferQuantity {
			return fmt.Errorf("pending token balance quantity is not equal to buy offer quantity: %d != %d", pendingTokenBalance.Quantity, invoice.BuyOfferQuantity)
		}

		err = s.MovePendingToTokenBalance(pendingTokenBalance, invoice.BuyOfferOffererAddress, tx)
		if err != nil {
			return err
		}

		err = tx.Commit()
		if err != nil {
			return err
		}
	}

	return nil
}
