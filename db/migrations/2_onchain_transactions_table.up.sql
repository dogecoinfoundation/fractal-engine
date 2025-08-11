CREATE TABLE IF NOT EXISTS onchain_transactions (
    id TEXT PRIMARY KEY,
    tx_hash TEXT NOT NULL,
    block_height BIGINT NOT NULL,
    block_hash TEXT NOT NULL,
    transaction_number INTEGER NOT NULL,
    action_type TEXT NOT NULL,
    action_version INTEGER NOT NULL,
    action_data BYTEA NOT NULL,
    address TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    value DOUBLE PRECISION NOT NULL
);
