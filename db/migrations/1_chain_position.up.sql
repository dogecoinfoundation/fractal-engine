CREATE TABLE IF NOT EXISTS chain_position (
    id INTEGER PRIMARY KEY,
	block_height TEXT NOT NULL,
	block_hash TEXT NOT NULL,
	waiting_for_next_hash BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE UNIQUE INDEX IF NOT EXISTS chain_position_key ON chain_position (id);
