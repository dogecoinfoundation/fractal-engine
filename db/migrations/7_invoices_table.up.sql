CREATE TABLE IF NOT EXISTS unconfirmed_invoices (
    id UUID PRIMARY KEY,
    hash TEXT NOT NULL,
    buy_offer_offerer_address TEXT NOT NULL,
    buy_offer_hash TEXT NOT NULL,
    buy_offer_mint_hash TEXT NOT NULL,
    buy_offer_quantity INT NOT NULL,
    buy_offer_price INT NOT NULL,
    buy_offer_value DOUBLE PRECISION NOT NULL,
    payment_address TEXT NOT NULL,
    sell_offer_address TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    public_key TEXT NOT NULL,
    signature TEXT NOT NULL
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
    buy_offer_offerer_address TEXT NOT NULL,
    buy_offer_hash TEXT NOT NULL,
    buy_offer_mint_hash TEXT NOT NULL,
    buy_offer_quantity INT NOT NULL,
    buy_offer_price INT NOT NULL,
    paid_at TIMESTAMP,
    buy_offer_value DOUBLE PRECISION NOT NULL,
    sell_offer_address TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    public_key TEXT NOT NULL,
    signature TEXT NOT NULL
);