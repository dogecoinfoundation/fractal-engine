CREATE TABLE IF NOT EXISTS sell_offers (
    id UUID PRIMARY KEY,
    offerer_address TEXT NOT NULL,
    hash TEXT NOT NULL,
    mint_hash text not null,
    quantity int not null,
    price int not null,
    created_at timestamp not null,
    public_key text not null,
    signature text not null
);

CREATE TABLE IF NOT EXISTS buy_offers (
    id UUID PRIMARY KEY,
    offerer_address TEXT NOT NULL,
    seller_address TEXT NOT NULL,
    hash text not null,
    mint_hash text not null,
    quantity int not null,
    price int not null,
    created_at timestamp not null,
    public_key text not null,
    signature text not null
);