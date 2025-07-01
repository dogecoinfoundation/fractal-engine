create table ufeo (
    id uuid primary key,
    quantity int not null,
    mint_hash text not null,
    created_at timestamp not null,
    block_height int not null
);

