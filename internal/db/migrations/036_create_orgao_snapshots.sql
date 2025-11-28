-- +goose Up
-- Create orgao_snapshots table for caching external orgao data from RMI API
CREATE TABLE orgao_snapshots (
    id SERIAL PRIMARY KEY,
    orgao_id VARCHAR(100) UNIQUE NOT NULL,  -- External orgao ID reference
    name TEXT NOT NULL,                      -- Full orgao name (nome_ua)
    sigla VARCHAR(50),                       -- Orgao acronym (sigla_ua)
    metadata JSONB,                          -- Additional flexible data from API
    last_synced_at TIMESTAMP NOT NULL,       -- Last successful sync timestamp
    sync_status VARCHAR(50) DEFAULT 'synced' NOT NULL, -- synced, failed, pending
    sync_error TEXT,                         -- Error message if sync failed
    created_at TIMESTAMP DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP DEFAULT NOW() NOT NULL
);

-- Indexes for performance
CREATE INDEX idx_orgao_snapshots_orgao_id ON orgao_snapshots(orgao_id);
CREATE INDEX idx_orgao_snapshots_last_synced ON orgao_snapshots(last_synced_at);
CREATE INDEX idx_orgao_snapshots_status ON orgao_snapshots(sync_status);

-- +goose Down
-- Drop indexes
DROP INDEX IF EXISTS idx_orgao_snapshots_status;
DROP INDEX IF EXISTS idx_orgao_snapshots_last_synced;
DROP INDEX IF EXISTS idx_orgao_snapshots_orgao_id;

-- Drop table
DROP TABLE IF EXISTS orgao_snapshots;
