
CREATE TABLE transfer_requests (
    id text primary key,
    sender_address BIGINT NOT NULL,
    receiver_address BIGINT NOT NULL,
    mint_id text NOT NULL,
    amount BIGINT NOT NULL,
    price_per_token BIGINT NOT NULL,
    transaction_hash text,
    approved boolean NOT NULL DEFAULT false,
    approved_at TIMESTAMP,
    verified boolean not null default false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE transfers (
    id text primary key,
    transfer_request_id text NOT NULL,
    transaction_hash text NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE accounts (
    id text primary key,
    address VARCHAR(255) NOT NULL,
    balance BIGINT NOT NULL DEFAULT 0,
    mint_id text NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

