create table invoice_signatures (
    id UUID PRIMARY KEY,
    invoice_hash TEXT NOT NULL,
    signature TEXT NOT NULL,
    public_key TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL
);

alter table invoices add column signature_id UUID;
