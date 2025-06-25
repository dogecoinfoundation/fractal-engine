package store

import "github.com/google/uuid"

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

func (s *TokenisationStore) SaveOnChainTransaction(tx_hash string, height int64, action_type uint8, action_version uint8, action_data []byte, address string) error {
	id := uuid.New().String()

	_, err := s.DB.Exec(`
	INSERT INTO onchain_transactions (id, tx_hash, block_height, action_type, action_version, action_data, address)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, id, tx_hash, height, action_type, action_version, action_data, address)

	return err
}

func (s *TokenisationStore) CountOnChainTransactions(blockHeight int64) (int, error) {
	var count int
	err := s.DB.QueryRow("SELECT COUNT(*) FROM onchain_transactions WHERE block_height = $1", blockHeight).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *TokenisationStore) GetOnChainTransactions(limit int) ([]OnChainTransaction, error) {
	rows, err := s.DB.Query("SELECT id, tx_hash, block_height, action_type, action_version, action_data, address FROM onchain_transactions LIMIT $1", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []OnChainTransaction
	for rows.Next() {
		var transaction OnChainTransaction
		if err := rows.Scan(&transaction.Id, &transaction.TxHash, &transaction.Height, &transaction.ActionType, &transaction.ActionVersion, &transaction.ActionData, &transaction.Address); err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}
