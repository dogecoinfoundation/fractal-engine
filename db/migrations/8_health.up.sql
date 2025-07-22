CREATE TABLE IF NOT EXISTS health (
    id SERIAL PRIMARY KEY,
    current_block_height BIGINT NOT NULL,
    latest_block_height BIGINT NOT NULL,
    chain TEXT NOT NULL,
    wallets_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);