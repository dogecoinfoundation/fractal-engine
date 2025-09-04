package store

import (
	"database/sql"
	"fmt"

	"dogecoin.org/fractal-engine/pkg/protocol"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

func (s *TokenisationStore) ChooseInvoice() (Invoice, error) {
	row := s.DB.QueryRow("SELECT id, hash, payment_address, buyer_address, mint_hash, quantity, price, created_at, seller_address, public_key, signature, paid_at FROM invoices WHERE hash IN (SELECT hash FROM invoices ORDER BY RANDOM() LIMIT 1)")
	var invoice Invoice
	if err := row.Scan(&invoice.Id, &invoice.Hash, &invoice.PaymentAddress, &invoice.BuyerAddress, &invoice.MintHash, &invoice.Quantity, &invoice.Price, &invoice.CreatedAt, &invoice.SellerAddress, &invoice.PublicKey, &invoice.Signature, &invoice.PaidAt); err != nil {
		return Invoice{}, err
	}
	return invoice, nil
}

func (s *TokenisationStore) GetInvoicesForMe(offset int, limit int, myAddress string) ([]Invoice, error) {
	rows, err := s.DB.Query("SELECT id, hash, payment_address, buyer_address, mint_hash, quantity, price, created_at, seller_address, public_key, signature, paid_at FROM invoices WHERE (buyer_address = $1 OR seller_address = $1) LIMIT $2 OFFSET $3", myAddress, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoices []Invoice

	for rows.Next() {
		var invoice Invoice
		if err := rows.Scan(&invoice.Id, &invoice.Hash, &invoice.PaymentAddress, &invoice.BuyerAddress, &invoice.MintHash, &invoice.Quantity, &invoice.Price, &invoice.CreatedAt, &invoice.SellerAddress, &invoice.PublicKey, &invoice.Signature, &invoice.PaidAt); err != nil {
			return nil, err
		}

		invoices = append(invoices, invoice)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return invoices, nil
}

func (s *TokenisationStore) GetInvoices(offset int, limit int, mintHash string, offererAddress string) ([]Invoice, error) {
	rows, err := s.DB.Query("SELECT id, hash, payment_address, buyer_address, mint_hash, quantity, price, created_at, seller_address, public_key, signature, paid_at FROM invoices WHERE mint_hash = $1 AND (buyer_address = $2 OR seller_address = $2) LIMIT $3 OFFSET $4", mintHash, offererAddress, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoices []Invoice

	for rows.Next() {
		var invoice Invoice
		if err := rows.Scan(&invoice.Id, &invoice.Hash, &invoice.PaymentAddress, &invoice.BuyerAddress, &invoice.MintHash, &invoice.Quantity, &invoice.Price, &invoice.CreatedAt, &invoice.SellerAddress, &invoice.PublicKey, &invoice.Signature, &invoice.PaidAt); err != nil {
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
	row := s.DB.QueryRow("SELECT COUNT(*) FROM unconfirmed_invoices WHERE mint_hash = $1 AND buyer_address = $2", mintHash, offererAddress)
	var count int
	err := row.Scan(&count)
	return count, err
}

func (s *TokenisationStore) GetUnconfirmedInvoices(offset int, limit int, mintHash string, offererAddress string) ([]UnconfirmedInvoice, error) {
	rows, err := s.DB.Query("SELECT id, hash, payment_address, buyer_address, mint_hash, quantity, price, created_at, seller_address, public_key, signature FROM unconfirmed_invoices WHERE mint_hash = $1 AND buyer_address = $2 LIMIT $3 OFFSET $4", mintHash, offererAddress, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoices []UnconfirmedInvoice

	for rows.Next() {
		var invoice UnconfirmedInvoice
		if err := rows.Scan(&invoice.Id, &invoice.Hash, &invoice.PaymentAddress, &invoice.BuyerAddress, &invoice.MintHash, &invoice.Quantity, &invoice.Price, &invoice.CreatedAt, &invoice.SellerAddress, &invoice.PublicKey, &invoice.Signature); err != nil {
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
	INSERT INTO unconfirmed_invoices (id, hash, payment_address, buyer_address, mint_hash, quantity, price, created_at, seller_address, public_key, signature)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, id, invoice.Hash, invoice.PaymentAddress, invoice.BuyerAddress, invoice.MintHash, invoice.Quantity, invoice.Price, invoice.CreatedAt, invoice.SellerAddress, invoice.PublicKey, invoice.Signature)

	return id, err
}

func (s *TokenisationStore) SaveInvoice(invoice *Invoice) (string, error) {
	return s.SaveInvoiceWithTx(invoice, nil)
}

func (s *TokenisationStore) SaveInvoiceWithTx(invoice *Invoice, tx *sql.Tx) (string, error) {
	id := uuid.New().String()

	query := `
	INSERT INTO invoices (id, hash, payment_address, buyer_address, mint_hash, quantity, price, created_at, seller_address, block_height, transaction_hash, public_key, signature)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	var err error
	if tx != nil {
		_, err = tx.Exec(query, id, invoice.Hash, invoice.PaymentAddress, invoice.BuyerAddress, invoice.MintHash, invoice.Quantity, invoice.Price, invoice.CreatedAt, invoice.SellerAddress, invoice.BlockHeight, invoice.TransactionHash, invoice.PublicKey, invoice.Signature)
	} else {
		_, err = s.DB.Exec(query, id, invoice.Hash, invoice.PaymentAddress, invoice.BuyerAddress, invoice.MintHash, invoice.Quantity, invoice.Price, invoice.CreatedAt, invoice.SellerAddress, invoice.BlockHeight, invoice.TransactionHash, invoice.PublicKey, invoice.Signature)
	}

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

	// Start transaction for atomic operations
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	rows, err := tx.Query("SELECT id, hash, buyer_address, mint_hash, quantity, price, created_at, seller_address, public_key, signature FROM unconfirmed_invoices WHERE hash = $1", onchainMessage.InvoiceHash)
	if err != nil {
		return err
	}

	var unconfirmedInvoice UnconfirmedInvoice
	if rows.Next() {
		if err := rows.Scan(
			&unconfirmedInvoice.Id, &unconfirmedInvoice.Hash, &unconfirmedInvoice.BuyerAddress, &unconfirmedInvoice.MintHash, &unconfirmedInvoice.Quantity, &unconfirmedInvoice.Price, &unconfirmedInvoice.CreatedAt, &unconfirmedInvoice.SellerAddress, &unconfirmedInvoice.PublicKey, &unconfirmedInvoice.Signature); err != nil {
			return err
		}
	} else {
		rows.Close()
		return fmt.Errorf("no unconfirmed invoice found for hash: %s", onchainMessage.InvoiceHash)
	}

	rows.Close()

	pendingTokenBalance, err := s.GetPendingTokenBalance(unconfirmedInvoice.Hash, unconfirmedInvoice.MintHash, tx)
	if err != nil {
		return err
	}

	if pendingTokenBalance.Quantity < unconfirmedInvoice.Quantity {
		return fmt.Errorf("pending token balance is less than the buy offer quantity: %d < %d", pendingTokenBalance.Quantity, unconfirmedInvoice.Quantity)
	}

	// Use transaction-aware SaveInvoice
	id, err := s.SaveInvoiceWithTx(&Invoice{
		Hash:            unconfirmedInvoice.Hash,
		PaymentAddress:  unconfirmedInvoice.PaymentAddress,
		BuyerAddress:    unconfirmedInvoice.BuyerAddress,
		MintHash:        unconfirmedInvoice.MintHash,
		Quantity:        unconfirmedInvoice.Quantity,
		Price:           unconfirmedInvoice.Price,
		CreatedAt:       unconfirmedInvoice.CreatedAt,
		SellerAddress:   unconfirmedInvoice.SellerAddress,
		PublicKey:       unconfirmedInvoice.PublicKey,
		Signature:       unconfirmedInvoice.Signature,
		BlockHeight:     onchainTransaction.Height,
		TransactionHash: onchainTransaction.TxHash,
	}, tx)

	if err != nil {
		return err
	}

	fmt.Println("Saved invoice:", id)

	_, err = tx.Exec("DELETE FROM unconfirmed_invoices WHERE id = $1", unconfirmedInvoice.Id)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM onchain_transactions WHERE id = $1", onchainTransaction.Id)
	if err != nil {
		return err
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
