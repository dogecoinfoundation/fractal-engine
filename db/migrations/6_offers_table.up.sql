create table offers (
    id uuid primary key,
    type int not null,
    offerer_address text not null,
    hash text not null,
    mint_hash text not null,
    quantity int not null,
    price int not null,
    created_at timestamp not null
);
