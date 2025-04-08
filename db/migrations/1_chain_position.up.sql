CREATE TABLE IF NOT EXISTS chain_position (
    id INTEGER PRIMARY KEY,
	block_height TEXT NOT NULL,
	block_hash TEXT NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS chain_position_key ON chain_position (id);
