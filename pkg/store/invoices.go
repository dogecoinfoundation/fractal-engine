package store

import (
	"fmt"

	"dogecoin.org/fractal-engine/pkg/protocol"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

func (s *TokenisationStore) GetInvoices(offset int, limit int, mintHash string, offererAddress string) ([]Invoice, error) {
	rows, err := s.DB.Query("SELECT id, hash, payment_address, buy_offer_offerer_address, buy_offer_hash, buy_offer_mint_hash, buy_offer_quantity, buy_offer_price, buy_offer_value, created_at, sell_offer_address, public_key, signature FROM invoices WHERE buy_offer_mint_hash = $1 AND (buy_offer_offerer_address = $2 OR sell_offer_address = $2) LIMIT $3 OFFSET $4", mintHash, offererAddress, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoices []Invoice

	for rows.Next() {
		var invoice Invoice
		if err := rows.Scan(&invoice.Id, &invoice.Hash, &invoice.PaymentAddress, &invoice.BuyOfferOffererAddress, &invoice.BuyOfferHash, &invoice.BuyOfferMintHash, &invoice.BuyOfferQuantity, &invoice.BuyOfferPrice, &invoice.BuyOfferValue, &invoice.CreatedAt, &invoice.SellOfferAddress, &invoice.PublicKey, &invoice.Signature); err != nil {
			return nil, err
		}

		invoices = append(invoices, invoice)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return invoices, nil
}

func (s *TokenisationStore) CountUnconfirmedInvoices(mintHash string, offererAddress string) (int, error) {
	row := s.DB.QueryRow("SELECT COUNT(*) FROM unconfirmed_invoices WHERE buy_offer_mint_hash = $1 AND buy_offer_offerer_address = $2", mintHash, offererAddress)
	var count int
	err := row.Scan(&count)
	return count, err
}

func (s *TokenisationStore) GetUnconfirmedInvoices(offset int, limit int, mintHash string, offererAddress string) ([]UnconfirmedInvoice, error) {
	rows, err := s.DB.Query("SELECT id, hash, payment_address, buy_offer_offerer_address, buy_offer_hash, buy_offer_mint_hash, buy_offer_quantity, buy_offer_price, buy_offer_value, created_at, sell_offer_address, public_key, signature FROM unconfirmed_invoices WHERE buy_offer_mint_hash = $1 AND buy_offer_offerer_address = $2 LIMIT $3 OFFSET $4", mintHash, offererAddress, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoices []UnconfirmedInvoice

	for rows.Next() {
		var invoice UnconfirmedInvoice
		if err := rows.Scan(&invoice.Id, &invoice.Hash, &invoice.PaymentAddress, &invoice.BuyOfferOffererAddress, &invoice.BuyOfferHash, &invoice.BuyOfferMintHash, &invoice.BuyOfferQuantity, &invoice.BuyOfferPrice, &invoice.BuyOfferValue, &invoice.CreatedAt, &invoice.SellOfferAddress, &invoice.PublicKey, &invoice.Signature); err != nil {
			return nil, err
		}

		invoices = append(invoices, invoice)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return invoices, nil
}

func (s *TokenisationStore) SaveUnconfirmedInvoice(invoice *UnconfirmedInvoice) (string, error) {
	id := uuid.New().String()

	_, err := s.DB.Exec(`
	INSERT INTO unconfirmed_invoices (id, hash, payment_address, buy_offer_offerer_address, buy_offer_hash, buy_offer_mint_hash, buy_offer_quantity, buy_offer_price, buy_offer_value, created_at, sell_offer_address, public_key, signature)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, id, invoice.Hash, invoice.PaymentAddress, invoice.BuyOfferOffererAddress, invoice.BuyOfferHash, invoice.BuyOfferMintHash, invoice.BuyOfferQuantity, invoice.BuyOfferPrice, invoice.BuyOfferValue, invoice.CreatedAt, invoice.SellOfferAddress, invoice.PublicKey, invoice.Signature)

	return id, err
}

func (s *TokenisationStore) SaveInvoice(invoice *Invoice) (string, error) {
	id := uuid.New().String()

	_, err := s.DB.Exec(`
	INSERT INTO invoices (id, hash, payment_address, buy_offer_offerer_address, buy_offer_hash, buy_offer_mint_hash, buy_offer_quantity, buy_offer_price, buy_offer_value, created_at, sell_offer_address, block_height, transaction_hash, public_key, signature)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`, id, invoice.Hash, invoice.PaymentAddress, invoice.BuyOfferOffererAddress, invoice.BuyOfferHash, invoice.BuyOfferMintHash, invoice.BuyOfferQuantity, invoice.BuyOfferPrice, invoice.BuyOfferValue, invoice.CreatedAt, invoice.SellOfferAddress, invoice.BlockHeight, invoice.TransactionHash, invoice.PublicKey, invoice.Signature)

	return id, err
}

func (s *TokenisationStore) MatchInvoice(onchainTransaction OnChainTransaction) bool {
	if onchainTransaction.ActionType != protocol.ACTION_INVOICE {
		return false
	}

	var onchainMessage protocol.OnChainInvoiceMessage
	err := proto.Unmarshal(onchainTransaction.ActionData, &onchainMessage)
	if err != nil {
		return false
	}

	rows, err := s.DB.Query("SELECT hash, transaction_hash FROM invoices WHERE transaction_hash = $1 and block_height = $2 and hash = $3", onchainTransaction.TxHash, onchainTransaction.Height, onchainMessage.InvoiceHash)
	if err != nil {
		return false
	}
	defer rows.Close()

	exists := rows.Next()

	if exists {
		_, err = s.DB.Exec("DELETE FROM onchain_transactions WHERE id = $1", onchainTransaction.Id)
		if err != nil {
			return false
		}
	}

	return exists
}

func (s *TokenisationStore) MatchUnconfirmedInvoice(onchainTransaction OnChainTransaction) error {
	if onchainTransaction.ActionType != protocol.ACTION_INVOICE {
		return fmt.Errorf("action type is not invoice: %d", onchainTransaction.ActionType)
	}

	var onchainMessage protocol.OnChainInvoiceMessage
	err := proto.Unmarshal(onchainTransaction.ActionData, &onchainMessage)
	if err != nil {
		return err
	}

	rows, err := s.DB.Query("SELECT id, hash, buy_offer_offerer_address, buy_offer_hash, buy_offer_mint_hash, buy_offer_quantity, buy_offer_price, buy_offer_value, created_at, sell_offer_address, public_key, signature FROM unconfirmed_invoices WHERE hash = $1", onchainMessage.InvoiceHash)
	if err != nil {
		return err
	}

	var unconfirmedInvoice UnconfirmedInvoice
	if rows.Next() {
		if err := rows.Scan(
			&unconfirmedInvoice.Id, &unconfirmedInvoice.Hash, &unconfirmedInvoice.BuyOfferOffererAddress, &unconfirmedInvoice.BuyOfferHash, &unconfirmedInvoice.BuyOfferMintHash, &unconfirmedInvoice.BuyOfferQuantity, &unconfirmedInvoice.BuyOfferPrice, &unconfirmedInvoice.BuyOfferValue, &unconfirmedInvoice.CreatedAt, &unconfirmedInvoice.SellOfferAddress, &unconfirmedInvoice.PublicKey, &unconfirmedInvoice.Signature); err != nil {
			return err
		}
	} else {
		rows.Close()
		return fmt.Errorf("no unconfirmed invoice found for hash: %s", onchainMessage.InvoiceHash)
	}

	rows.Close()

	pendingTokenBalance, err := s.GetPendingTokenBalance(unconfirmedInvoice.Hash, unconfirmedInvoice.BuyOfferMintHash, nil)
	if err != nil {
		return err
	}

	if pendingTokenBalance.Quantity < unconfirmedInvoice.BuyOfferQuantity {
		return fmt.Errorf("pending token balance is less than the buy offer quantity: %d < %d", pendingTokenBalance.Quantity, unconfirmedInvoice.BuyOfferQuantity)
	}

	id, err := s.SaveInvoice(&Invoice{
		Hash:                   unconfirmedInvoice.Hash,
		PaymentAddress:         unconfirmedInvoice.PaymentAddress,
		BuyOfferOffererAddress: unconfirmedInvoice.BuyOfferOffererAddress,
		BuyOfferHash:           unconfirmedInvoice.BuyOfferHash,
		BuyOfferMintHash:       unconfirmedInvoice.BuyOfferMintHash,
		BuyOfferQuantity:       unconfirmedInvoice.BuyOfferQuantity,
		BuyOfferPrice:          unconfirmedInvoice.BuyOfferPrice,
		CreatedAt:              unconfirmedInvoice.CreatedAt,
		SellOfferAddress:       unconfirmedInvoice.SellOfferAddress,
		BuyOfferValue:          unconfirmedInvoice.BuyOfferValue,
		PublicKey:              unconfirmedInvoice.PublicKey,
		Signature:              unconfirmedInvoice.Signature,
		BlockHeight:            onchainTransaction.Height,
		TransactionHash:        onchainTransaction.TxHash,
	})

	if err != nil {
		return err
	}

	fmt.Println("Saved invoice:", id)

	_, err = s.DB.Exec("DELETE FROM unconfirmed_invoices WHERE id = $1", unconfirmedInvoice.Id)
	if err != nil {
		return err
	}

	_, err = s.DB.Exec("DELETE FROM onchain_transactions WHERE id = $1", onchainTransaction.Id)
	if err != nil {
		return err
	}

	return nil
}
