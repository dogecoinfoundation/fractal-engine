package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/protocol"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/protobuf/proto"
)

type TokenisationStore struct {
	DB      *sql.DB
	backend string
	cfg     config.Config
}

func (s *TokenisationStore) GetStats() (map[string]int, error) {
	stats := make(map[string]int)

	mints, err := getMintsCount(s)
	if err != nil {
		fmt.Println("Error getting mints count:", err)
		return nil, err
	}

	unconfirmedMints, err := getUnconfirmedMintsCount(s)
	if err != nil {
		fmt.Println("Error getting unconfirmed mints count:", err)
		return nil, err
	}

	onChainTransactions, err := getOnChainTransactionsCount(s)
	if err != nil {
		fmt.Println("Error getting onchain transactions count:", err)
		return nil, err
	}

	stats["mints"] = mints
	stats["unconfirmed_mints"] = unconfirmedMints
	stats["onchain_transactions"] = onChainTransactions

	return stats, nil

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

func (s *TokenisationStore) SaveOnChainTransaction(tx_hash string, height int64, action_type uint8, action_version uint8, action_data []byte) error {
	id := uuid.New().String()

	_, err := s.DB.Exec(`
	INSERT INTO onchain_transactions (id, tx_hash, block_height, action_type, action_version, action_data)
	VALUES ($1, $2, $3, $4, $5, $6)
	`, id, tx_hash, height, action_type, action_version, action_data)

	return err
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

func NewTokenisationStore(dbUrl string, cfg config.Config) (*TokenisationStore, error) {
	u, err := url.Parse(dbUrl)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "memory" {
		sqlite, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			return nil, err
		}

		return &TokenisationStore{DB: sqlite, backend: "sqlite", cfg: cfg}, nil
	} else if u.Scheme == "sqlite" {
		var url string
		if u.Host == "" {
			url = u.Path
		} else {
			url = u.Host
		}

		sqlite, err := sql.Open("sqlite3", url)
		if err != nil {
			return nil, err
		}

		return &TokenisationStore{DB: sqlite, backend: "sqlite", cfg: cfg}, nil
	} else if u.Scheme == "postgres" {
		postgres, err := sql.Open("postgres", dbUrl)
		if err != nil {
			return nil, err
		}
		return &TokenisationStore{DB: postgres, backend: "postgres", cfg: cfg}, nil
	}

	return nil, fmt.Errorf("unsupported database scheme: %s", u.Scheme)
}

func (s *TokenisationStore) Migrate() error {
	driver, err := s.getMigrationDriver()
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance("file://"+s.cfg.MigrationsPath, s.backend, driver)
	if err != nil {
		return err
	}

	return m.Up()
}

func ProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		// Check if go.mod exists in this directory
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory, cannot find go.mod
			return "", os.ErrNotExist
		}
		dir = parent
	}
}

func MigrationsPath() (string, error) {
	root, err := ProjectRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "db", "migrations"), nil
}

func (s *TokenisationStore) getMigrationDriver() (database.Driver, error) {
	if s.backend == "postgres" {
		driver, err := postgres.WithInstance(s.DB, &postgres.Config{})
		if err != nil {
			return nil, err
		}

		return driver, nil
	}

	if s.backend == "sqlite" {
		driver, err := sqlite.WithInstance(s.DB, &sqlite.Config{})
		if err != nil {
			return nil, err
		}

		return driver, nil
	}

	return nil, fmt.Errorf("unsupported database scheme: %s", s.backend)
}

func (s *TokenisationStore) ClearMints() error {
	_, err := s.DB.Exec("DELETE FROM mints")
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

func (s *TokenisationStore) GetOnChainTransactions(limit int) ([]OnChainTransaction, error) {
	rows, err := s.DB.Query("SELECT id, tx_hash, block_height, action_type, action_version, action_data FROM onchain_transactions LIMIT $1", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []OnChainTransaction
	for rows.Next() {
		var transaction OnChainTransaction
		if err := rows.Scan(&transaction.Id, &transaction.TxHash, &transaction.Height, &transaction.ActionType, &transaction.ActionVersion, &transaction.ActionData); err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

func (s *TokenisationStore) GetMints(offset int, limit int) ([]Mint, error) {
	fmt.Println("Getting mints:", limit, offset)

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

func (s *TokenisationStore) GetChainPosition() (int64, string, bool, error) {
	var blockHeight int64
	var blockHash string
	var waitingForNextHash bool

	err := s.DB.QueryRow("SELECT block_height, block_hash, waiting_for_next_hash FROM chain_position").Scan(&blockHeight, &blockHash, &waitingForNextHash)
	if err == sql.ErrNoRows {
		return 0, "", false, nil
	}

	if err != nil {
		return 0, "", false, err
	}

	return blockHeight, blockHash, waitingForNextHash, nil
}

func (s *TokenisationStore) UpsertChainPosition(blockHeight int64, blockHash string, waitingForNextHash bool) error {

	_, err := s.DB.Exec(`
	INSERT INTO chain_position (id, block_height, block_hash, waiting_for_next_hash)
	VALUES (1, $1, $2, $3)
	ON CONFLICT (id)
	DO UPDATE SET block_height = EXCLUDED.block_height,
				  block_hash = EXCLUDED.block_hash,
				  waiting_for_next_hash = EXCLUDED.waiting_for_next_hash
`, blockHeight, blockHash, waitingForNextHash)

	return err
}

func (s *TokenisationStore) Close() error {
	fmt.Println("Closing store")
	return s.DB.Close()
}
