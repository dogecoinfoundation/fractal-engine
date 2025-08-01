package store

import (
	"database/sql"
	"errors"
	"log"
	"time"
)

func (s *TokenisationStore) UpsertTokenBalance(address, mintHash string, quantity int) error {
	log.Println("Upserting token balance:", address, mintHash, quantity)

	_, err := s.DB.Exec(`
	INSERT INTO token_balances (address, mint_hash, quantity, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5)
	`, address, mintHash, quantity, time.Now(), time.Now())

	if err != nil {
		return err
	}

	return nil
}

func (s *TokenisationStore) UpsertPendingTokenBalance(invoiceHash, mintHash string, quantity int, onchainTransactionId string, ownerAddress string) error {
	log.Println("Upserting pending token balance:", invoiceHash, mintHash, quantity, onchainTransactionId, ownerAddress)

	_, err := s.DB.Exec(`
	INSERT INTO pending_token_balances (invoice_hash, mint_hash, quantity, onchain_transaction_id, created_at, owner_address)
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (invoice_hash, mint_hash)
	DO UPDATE SET quantity = $3
	`, invoiceHash, mintHash, quantity, onchainTransactionId, time.Now(), ownerAddress)

	if err != nil {
		return err
	}

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

func (s *TokenisationStore) RemovePendingTokenBalance(invoiceHash, mintHash string) error {
	_, err := s.DB.Exec(`
		DELETE FROM pending_token_balances WHERE invoice_hash = $1 AND mint_hash = $2
	`, invoiceHash, mintHash)
	return err
}

func (s *TokenisationStore) GetPendingTokenBalance(invoiceHash, mintHash string, tx *sql.Tx) (PendingTokenBalance, error) {
	var rows *sql.Rows
	var err error

	if tx == nil {
		rows, err = s.DB.Query(`
		SELECT quantity, invoice_hash, mint_hash, owner_address FROM pending_token_balances WHERE invoice_hash = $1 AND mint_hash = $2
	`, invoiceHash, mintHash)
	} else {
		rows, err = tx.Query(`
			SELECT quantity, invoice_hash, mint_hash, owner_address FROM pending_token_balances WHERE invoice_hash = $1 AND mint_hash = $2
		`, invoiceHash, mintHash)
	}

	defer rows.Close()

	if err != nil {
		return PendingTokenBalance{}, err
	}

	if rows.Next() {
		var pendingTokenBalance PendingTokenBalance
		err := rows.Scan(&pendingTokenBalance.Quantity, &pendingTokenBalance.InvoiceHash, &pendingTokenBalance.MintHash, &pendingTokenBalance.OwnerAddress)
		if err != nil {
			return PendingTokenBalance{}, err
		}

		return pendingTokenBalance, nil
	}

	return PendingTokenBalance{}, errors.New("no pending token balance found")
}

func (s *TokenisationStore) GetPendingTokenBalanceTotalForMintAndOwner(mintHash string, ownerAddress string) (int, error) {
	rows, err := s.DB.Query(`
		SELECT COALESCE(SUM(quantity), 0) FROM pending_token_balances WHERE mint_hash = $1 AND owner_address = $2
	`, mintHash, ownerAddress)
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

func (s *TokenisationStore) GetTokenBalances(address string, mintHash string) ([]TokenBalance, error) {
	log.Println("Getting token balance: ADDRESS", address, "MINT HASH", mintHash)

	rows, err := s.DB.Query(`
		SELECT quantity FROM token_balances WHERE address = $1 AND mint_hash = $2
	`, address, mintHash)

	if err != nil {
		return []TokenBalance{}, err
	}

	defer rows.Close()

	tokenBalances := []TokenBalance{}

	for rows.Next() {

		var quantity int
		err := rows.Scan(&quantity)
		if err != nil {
			return []TokenBalance{}, err
		}
		tokenBalances = append(tokenBalances, TokenBalance{
			Address:  address,
			MintHash: mintHash,
			Quantity: quantity,
		})
	}

	return tokenBalances, nil
}

func (s *TokenisationStore) UpsertTokenBalanceWithTransaction(address, mintHash string, quantity int, tx *sql.Tx) error {
	log.Println("Upserting token balance with transaction:", address, mintHash, quantity)

	_, err := tx.Exec(`
	INSERT INTO token_balances (address, mint_hash, quantity, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5)
	`, address, mintHash, quantity, time.Now(), time.Now())

	return err
}

func (s *TokenisationStore) MovePendingToTokenBalance(pendingTokenBalance PendingTokenBalance, buyerAddress string, tx *sql.Tx) error {
	_, err := tx.Exec(`
	INSERT INTO token_balances (address, mint_hash, quantity, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5)
	`, buyerAddress, pendingTokenBalance.MintHash, pendingTokenBalance.Quantity, time.Now(), time.Now())

	if err != nil {
		return err
	}

	err = s.UpsertTokenBalanceWithTransaction(pendingTokenBalance.OwnerAddress, pendingTokenBalance.MintHash, -pendingTokenBalance.Quantity, tx)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
	DELETE FROM pending_token_balances WHERE invoice_hash = $1 AND mint_hash = $2
	`, pendingTokenBalance.InvoiceHash, pendingTokenBalance.MintHash)
	if err != nil {
		return err
	}

	return err
}
