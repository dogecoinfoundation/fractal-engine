CREATE TABLE IF NOT EXISTS unconfirmed_invoices (
    id UUID PRIMARY KEY,
    hash TEXT NOT NULL,
    buyer_address TEXT NOT NULL,
    mint_hash TEXT NOT NULL,
    quantity INT NOT NULL,
    price INT NOT NULL,
    payment_address TEXT NOT NULL,
    seller_address TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    public_key TEXT NOT NULL,
    signature TEXT NOT NULL,
    status TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS token_balances (
    mint_hash TEXT NOT NULL,
    address TEXT NOT NULL,
    quantity INT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);


CREATE TABLE IF NOT EXISTS pending_token_balances (
    owner_address TEXT NOT NULL,
    invoice_hash TEXT NOT NULL,
    mint_hash TEXT NOT NULL,
    quantity INT NOT NULL,
    onchain_transaction_id TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    PRIMARY KEY (invoice_hash, mint_hash)
);

CREATE TABLE IF NOT EXISTS invoices (
    id UUID PRIMARY KEY,
    hash TEXT NOT NULL,
    block_height INT,
    transaction_hash TEXT,
    payment_address TEXT,
    buyer_address TEXT NOT NULL,
    mint_hash TEXT NOT NULL,
    quantity INT NOT NULL,
    price INT NOT NULL,
    paid_at TIMESTAMP,
    seller_address TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    public_key TEXT NOT NULL,
    signature TEXT NOT NULL
);
