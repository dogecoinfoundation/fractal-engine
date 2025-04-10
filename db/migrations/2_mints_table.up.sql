create table mints (
    id text primary key,
    title text not null,
    description text not null,
    fraction_count integer not null,
    tags TEXT,
    metadata TEXT,
    verified boolean not null default false,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
