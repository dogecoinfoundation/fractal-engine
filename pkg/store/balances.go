package store

import (
	"database/sql"
	"errors"
	"log"
	"time"
)

func (s *TokenisationStore) UpsertTokenBalance(address, mintHash string, quantity int) error {
	log.Println("Upserting token balance:", address, mintHash, quantity)

	res, err := s.DB.Exec(`
	INSERT INTO token_balances (address, mint_hash, quantity, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5)		
	ON CONFLICT (address, mint_hash)
	DO UPDATE SET quantity = EXCLUDED.quantity, updated_at = EXCLUDED.updated_at
	`, address, mintHash, quantity, time.Now(), time.Now())

	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("no rows affected")
	}

	return nil
}

func (s *TokenisationStore) UpsertPendingTokenBalance(invoiceHash, mintHash string, quantity int, txHash string, ownerAddress string) error {
	_, err := s.DB.Exec(`
	INSERT INTO pending_token_balances (invoice_hash, mint_hash, quantity, onchain_transaction_id, created_at, owner_address)
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (invoice_hash, mint_hash)
	DO UPDATE SET quantity = EXCLUDED.quantity + $3
	`, invoiceHash, mintHash, quantity, txHash, time.Now(), ownerAddress)
	return err
}

func (s *TokenisationStore) HasPendingTokenBalance(invoiceHash, mintHash string, onChainTransactionId string) (bool, error) {
	rows, err := s.DB.Query(`
		SELECT COUNT(*) FROM pending_token_balances WHERE invoice_hash = $1 AND mint_hash = $2 AND onchain_transaction_id = $3
	`, invoiceHash, mintHash, onChainTransactionId)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	if !rows.Next() {
		return false, nil
	}

	var count int
	err = rows.Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s *TokenisationStore) GetTokenBalance(address, mintHash string) (int, error) {
	log.Println("Getting token balance:", address, mintHash)

	rows, err := s.DB.Query(`
		SELECT quantity FROM token_balances WHERE address = $1 AND mint_hash = $2
	`, address, mintHash)

	if err != nil {
		return 0, err
	}

	defer rows.Close()

	if rows.Next() {
		var quantity int
		err := rows.Scan(&quantity)
		if err != nil {
			return 0, err
		}
		return quantity, nil
	}

	return 0, nil
}

func (s *TokenisationStore) UpsertTokenBalanceWithTransaction(address, mintHash string, quantity int, tx *sql.Tx) error {
	log.Println("Upserting token balance with transaction:", address, mintHash, quantity)

	_, err := tx.Exec(`
	INSERT INTO token_balances (address, mint_hash, quantity, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5)		
	ON CONFLICT (address, mint_hash)
	DO UPDATE SET quantity = EXCLUDED.quantity + $3, updated_at = EXCLUDED.updated_at
	`, address, mintHash, quantity, time.Now(), time.Now())

	return err
}
