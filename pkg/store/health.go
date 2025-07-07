package store

import "database/sql"

func (s *TokenisationStore) GetHealth() (int64, int64, error) {
	rows, err := s.DB.Query("SELECT current_block_height, latest_block_height FROM health")
	if err != nil {
		return 0, 0, err
	}

	defer rows.Close()

	var currentBlockHeight int64
	var latestBlockHeight int64

	if rows.Next() {
		err := rows.Scan(&currentBlockHeight, &latestBlockHeight)
		if err != nil {
			return 0, 0, err
		}
	} else {
		return 0, 0, sql.ErrNoRows
	}

	return currentBlockHeight, latestBlockHeight, nil
}

func (s *TokenisationStore) UpsertHealth(currentBlockHeight int64, latestBlockHeight int64) error {
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

	stmt, err = tx.Prepare("INSERT INTO health (current_block_height, latest_block_height) VALUES (?, ?)")
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(currentBlockHeight, latestBlockHeight)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
