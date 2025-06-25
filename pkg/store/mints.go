package store

import (
	"encoding/json"
	"fmt"

	"dogecoin.org/fractal-engine/pkg/protocol"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

func (s *TokenisationStore) GetMints(offset int, limit int) ([]Mint, error) {
	rows, err := s.DB.Query("SELECT id, created_at, title, description, fraction_count, tags, metadata, hash, transaction_hash, requirements, lockup_options, feed_url FROM mints LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mints []Mint
	for rows.Next() {
		var m Mint
		if err := rows.Scan(&m.Id, &m.CreatedAt, &m.Title, &m.Description, &m.FractionCount, &m.Tags, &m.Metadata, &m.Hash, &m.TransactionHash, &m.Requirements, &m.LockupOptions, &m.FeedURL); err != nil {
			return nil, err
		}
		mints = append(mints, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return mints, nil
}

func (s *TokenisationStore) SaveMint(mint *MintWithoutID) (string, error) {
	fmt.Println("Saving mint:", mint.Hash)

	id := uuid.New().String()

	metadata, err := json.Marshal(mint.Metadata)
	if err != nil {
		return "", err
	}

	requirements, err := json.Marshal(mint.Requirements)
	if err != nil {
		return "", err
	}

	lockupOptions, err := json.Marshal(mint.LockupOptions)
	if err != nil {
		return "", err
	}

	tags, err := json.Marshal(mint.Tags)
	if err != nil {
		return "", err
	}

	_, err = s.DB.Exec(`
	INSERT INTO mints (id, title, description, fraction_count, tags, metadata, hash, requirements, lockup_options, feed_url)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, id, mint.Title, mint.Description, mint.FractionCount, string(tags), string(metadata), mint.Hash, string(requirements), string(lockupOptions), mint.FeedURL)

	return id, err
}

func (s *TokenisationStore) TrimOldUnconfirmedMints(limit int) error {
	sqlQuery := fmt.Sprintf("DELETE FROM unconfirmed_mints WHERE id NOT IN (SELECT id FROM unconfirmed_mints ORDER BY id DESC LIMIT %d)", limit)
	_, err := s.DB.Exec(sqlQuery)
	if err != nil {
		return err
	}
	return nil
}

func (s *TokenisationStore) SaveUnconfirmedMint(mint *MintWithoutID) (string, error) {
	fmt.Println("Saving unconfirmed mint:", mint.Hash)

	id := uuid.New().String()

	metadata, err := json.Marshal(mint.Metadata)
	if err != nil {
		return "", err
	}

	requirements, err := json.Marshal(mint.Requirements)
	if err != nil {
		return "", err
	}

	lockupOptions, err := json.Marshal(mint.LockupOptions)
	if err != nil {
		return "", err
	}

	tags, err := json.Marshal(mint.Tags)
	if err != nil {
		return "", err
	}

	_, err = s.DB.Exec(`
	INSERT INTO unconfirmed_mints (id, title, description, fraction_count, tags, metadata, hash, requirements, lockup_options, feed_url)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, id, mint.Title, mint.Description, mint.FractionCount, string(tags), string(metadata), mint.Hash, string(requirements), string(lockupOptions), mint.FeedURL)

	return id, err
}

func (s *TokenisationStore) MatchMint(onchainTransaction OnChainTransaction) bool {
	if onchainTransaction.ActionType != protocol.ACTION_MINT {
		return false
	}

	var onchainMessage protocol.OnChainMintMessage
	err := proto.Unmarshal(onchainTransaction.ActionData, &onchainMessage)
	if err != nil {
		return false
	}

	if onchainMessage.Hash != onchainTransaction.TxHash {
		return false
	}

	rows, err := s.DB.Query("SELECT hash, transaction_hash FROM mints WHERE transaction_hash = $1 and block_height = $2 and hash = $3", onchainTransaction.TxHash, onchainTransaction.Height, onchainMessage.Hash)
	if err != nil {
		return false
	}
	defer rows.Close()

	exists := rows.Next()

	if exists {
		_, err = s.DB.Exec("DELETE FROM onchain_transactions WHERE $1", onchainTransaction.Id)
		if err != nil {
			return false
		}
	}

	return exists
}

func (s *TokenisationStore) MatchUnconfirmedMint(onchainTransaction OnChainTransaction) error {
	if onchainTransaction.ActionType != protocol.ACTION_MINT {
		return fmt.Errorf("action type is not mint: %d", onchainTransaction.ActionType)
	}

	var onchainMessage protocol.OnChainMintMessage
	err := proto.Unmarshal(onchainTransaction.ActionData, &onchainMessage)
	if err != nil {
		return err
	}

	rows, err := s.DB.Query("SELECT id, title, description, fraction_count, tags, metadata, hash, transaction_hash, requirements, lockup_options, feed_url FROM unconfirmed_mints WHERE hash = $1", onchainMessage.Hash)
	if err != nil {
		return err
	}

	var unconfirmedMint Mint
	if rows.Next() {
		if err := rows.Scan(
			&unconfirmedMint.Id, &unconfirmedMint.Title, &unconfirmedMint.Description,
			&unconfirmedMint.FractionCount, &unconfirmedMint.Tags, &unconfirmedMint.Metadata,
			&unconfirmedMint.Hash, &unconfirmedMint.TransactionHash, &unconfirmedMint.Requirements,
			&unconfirmedMint.LockupOptions, &unconfirmedMint.FeedURL); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("no unconfirmed mint found for hash: %s", onchainMessage.Hash)
	}

	rows.Close()

	id, err := s.SaveMint(&MintWithoutID{
		Hash:            unconfirmedMint.Hash,
		Title:           unconfirmedMint.Title,
		FractionCount:   unconfirmedMint.FractionCount,
		Description:     unconfirmedMint.Description,
		Tags:            unconfirmedMint.Tags,
		Metadata:        unconfirmedMint.Metadata,
		TransactionHash: unconfirmedMint.TransactionHash,
		BlockHeight:     unconfirmedMint.BlockHeight,
		CreatedAt:       unconfirmedMint.CreatedAt,
		Requirements:    unconfirmedMint.Requirements,
		LockupOptions:   unconfirmedMint.LockupOptions,
		FeedURL:         unconfirmedMint.FeedURL,
	})

	if err != nil {
		return err
	}

	fmt.Println("Saved mint:", id)

	_, err = s.DB.Exec("DELETE FROM unconfirmed_mints WHERE id = $1", unconfirmedMint.Id)
	if err != nil {
		return err
	}

	_, err = s.DB.Exec("DELETE FROM onchain_transactions WHERE $1", onchainTransaction.Id)
	if err != nil {
		return err
	}

	return nil
}

func getMintsCount(s *TokenisationStore) (int, error) {
	rows, err := s.DB.Query("SELECT COUNT(*) FROM mints")
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

func getUnconfirmedMintsCount(s *TokenisationStore) (int, error) {
	rows, err := s.DB.Query("SELECT COUNT(*) FROM unconfirmed_mints")
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

func (s *TokenisationStore) ClearMints() error {
	_, err := s.DB.Exec("DELETE FROM mints")
	if err != nil {
		return err
	}
	return nil
}
