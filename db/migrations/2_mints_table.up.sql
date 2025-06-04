create table mints (
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
    resellable boolean not null default true,
    lockup_options TEXT,
    gossiped boolean not null default false,
    verified boolean not null default false,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
