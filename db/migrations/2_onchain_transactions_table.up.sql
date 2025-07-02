create table onchain_transactions (
    id text primary key,
    tx_hash text not null,
    block_height bigint not null,
    transaction_number integer not null,
    action_type text not null,
    action_version integer not null,
    action_data blob not null,
    address text not null,
    created_at datetime not null default current_timestamp,
    value float not null
);


