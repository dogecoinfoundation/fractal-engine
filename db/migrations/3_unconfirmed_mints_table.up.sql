CREATE TABLE IF NOT EXISTS unconfirmed_mints (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    fraction_count INTEGER NOT NULL,
    tags TEXT,
    transaction_hash TEXT,
    block_height INTEGER,
    owner_address TEXT,
    metadata TEXT,
    hash TEXT,
    requirements TEXT,
    lockup_options TEXT,
    feed_url TEXT,
    public_key TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
