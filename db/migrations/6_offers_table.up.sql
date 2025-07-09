create table sell_offers (
    id uuid primary key,
    offerer_address text not null,
    hash text not null,
    mint_hash text not null,
    quantity int not null,
    price int not null,
    created_at timestamp not null
);

create table buy_offers (
    id uuid primary key,
    offerer_address text not null,
    seller_address text not null,
    hash text not null,
    mint_hash text not null,
    quantity int not null,
    price int not null,
    created_at timestamp not null
);