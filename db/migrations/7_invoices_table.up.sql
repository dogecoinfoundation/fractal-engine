create table unconfirmed_invoices (
    id uuid primary key,
    hash text not null,
    buy_offer_offerer_address text not null,
    buy_offer_hash text not null,
    buy_offer_mint_hash text not null,
    buy_offer_quantity int not null,
    buy_offer_price int not null,
    buy_offer_value float not null,
    payment_address text not null,
    sell_offer_address text not null,
    created_at timestamp not null
);

create table token_balances (
    id uuid primary key,
    mint_hash text not null,
    address text not null,
    quantity int not null,
    created_at timestamp not null,
    updated_at timestamp not null
);
 
create table invoices (
    id uuid primary key,
    hash text not null,
    block_height int,
    transaction_hash text,
    payment_address text,
    buy_offer_offerer_address text not null,
    buy_offer_hash text not null,
    buy_offer_mint_hash text not null,
    buy_offer_quantity int not null,
    buy_offer_price int not null,
    paid_at timestamp,
    buy_offer_value float not null,
    sell_offer_address text not null,
    created_at timestamp not null
);