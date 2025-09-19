create table invoice_signatures (
    id UUID PRIMARY KEY,
    invoice_hash TEXT NOT NULL,
    signature TEXT NOT NULL,
    public_key TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL
);

alter table invoices add column signature_id UUID;

CREATE UNIQUE INDEX IF NOT EXISTS unique_invoice_hash_public_key_idx
  ON invoice_signatures (invoice_hash, public_key);