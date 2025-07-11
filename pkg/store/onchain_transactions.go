package store

import (
	"fmt"

	"github.com/google/uuid"
)

func getOnChainTransactionsCount(s *TokenisationStore) (int, error) {
	rows, err := s.DB.Query("SELECT COUNT(*) FROM onchain_transactions")
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	if !rows.Next() {
		return 0, nil
	}

	var count int
	err = rows.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *TokenisationStore) SaveOnChainTransaction(tx_hash string, height int64, transaction_number int, action_type uint8, action_version uint8, action_data []byte, address string, value float64) (string, error) {
	id := uuid.New().String()

	_, err := s.DB.Exec(`
	INSERT INTO onchain_transactions (id, tx_hash, block_height, transaction_number, action_type, action_version, action_data, address, value)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, id, tx_hash, height, transaction_number, action_type, action_version, action_data, address, value)

	return id, err
}

func (s *TokenisationStore) GetOldOnchainTransactions(blockHeight int) ([]OnChainTransaction, error) {
	rows, err := s.DB.Query("SELECT id, tx_hash, block_height, transaction_number, action_type, action_version, action_data, address, value FROM onchain_transactions WHERE block_height < $1", blockHeight)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []OnChainTransaction
	for rows.Next() {
		var transaction OnChainTransaction
		if err := rows.Scan(&transaction.Id, &transaction.TxHash, &transaction.Height, &transaction.TransactionNumber, &transaction.ActionType, &transaction.ActionVersion, &transaction.ActionData, &transaction.Address, &transaction.Value); err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

func (s *TokenisationStore) TrimOldOnChainTransactions(blockHeightToKeep int) error {
	sqlQuery := fmt.Sprintf("DELETE FROM onchain_transactions WHERE block_height < %d", blockHeightToKeep)
	_, err := s.DB.Exec(sqlQuery)
	if err != nil {
		return err
	}
	return nil
}

func (s *TokenisationStore) RemoveOnChainTransaction(id string) error {
	_, err := s.DB.Exec("DELETE FROM onchain_transactions WHERE id = $1", id)
	if err != nil {
		return err
	}
	return nil
}

func (s *TokenisationStore) CountOnChainTransactions(blockHeight int64) (int, error) {
	var count int
	err := s.DB.QueryRow("SELECT COUNT(*) FROM onchain_transactions WHERE block_height = $1", blockHeight).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *TokenisationStore) GetOnChainTransactions(offset int, limit int) ([]OnChainTransaction, error) {
	rows, err := s.DB.Query("SELECT id, tx_hash, block_height, transaction_number, action_type, action_version, action_data, address, value FROM onchain_transactions ORDER BY block_height ASC, transaction_number ASC LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []OnChainTransaction
	for rows.Next() {
		var transaction OnChainTransaction
		if err := rows.Scan(&transaction.Id, &transaction.TxHash, &transaction.Height, &transaction.TransactionNumber, &transaction.ActionType, &transaction.ActionVersion, &transaction.ActionData, &transaction.Address, &transaction.Value); err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}
