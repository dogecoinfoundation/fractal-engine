create table unconfirmed_mints (
    id text primary key,
    title text not null,
    description text not null,
    fraction_count integer not null,
    tags TEXT,
    transaction_hash TEXT,
    output_address TEXT,
    metadata TEXT,
    hash TEXT,
    requirements TEXT,
    lockup_options TEXT,
    feed_url TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

 