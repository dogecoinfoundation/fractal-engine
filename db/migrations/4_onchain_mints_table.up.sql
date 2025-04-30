create table onchain_mints (
    id text primary key,
    hash TEXT,
    transaction_hash TEXT,
    output_address TEXT,
    verified boolean not null default false,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
