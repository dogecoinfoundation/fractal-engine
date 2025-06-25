package store

import "time"

func (s *TokenisationStore) UpsertTokenBalance(address, mintHash string, quantity int) error {
	_, err := s.DB.Exec(`
	INSERT INTO token_balances (address, mint_hash, quantity, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5)		
	ON CONFLICT (address, mint_hash)
	DO UPDATE SET quantity = EXCLUDED.quantity + $3, updated_at = EXCLUDED.updated_at
	`, address, mintHash, quantity, time.Now(), time.Now())

	return err
}
