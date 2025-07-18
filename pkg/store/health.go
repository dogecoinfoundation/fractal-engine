package store

import (
	"database/sql"
	"time"
)

func (s *TokenisationStore) GetHealth() (int64, int64, string, bool, time.Time, error) {
	rows, err := s.DB.Query("SELECT current_block_height, latest_block_height, chain, wallets_enabled, updated_at FROM health")
	if err != nil {
		return 0, 0, "", false, time.Time{}, err
	}

	defer rows.Close()

	var currentBlockHeight int64
	var latestBlockHeight int64
	var chain string
	var walletsEnabled bool
	var updatedAt time.Time

	if rows.Next() {
		err := rows.Scan(&currentBlockHeight, &latestBlockHeight, &chain, &walletsEnabled, &updatedAt)
		if err != nil {
			return 0, 0, "", false, time.Time{}, err
		}
	} else {
		return 0, 0, "", false, time.Time{}, sql.ErrNoRows
	}

	return currentBlockHeight, latestBlockHeight, chain, walletsEnabled, updatedAt, nil
}

func (s *TokenisationStore) UpsertHealth(currentBlockHeight int64, latestBlockHeight int64, chain string, walletsEnabled bool) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("DELETE FROM health")
	if err != nil {
		return err
	}

	_, err = stmt.Exec()
	if err != nil {
		return err
	}

	stmt, err = tx.Prepare("INSERT INTO health (current_block_height, latest_block_height, chain, wallets_enabled, updated_at) VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)")
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(currentBlockHeight, latestBlockHeight, chain, walletsEnabled)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
