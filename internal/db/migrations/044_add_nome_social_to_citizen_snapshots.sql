-- +goose Up
-- +goose StatementBegin

-- Add nome_social column if it doesn't exist (fixes issue where table was created before this column was added)
ALTER TABLE citizen_snapshots ADD COLUMN IF NOT EXISTS nome_social VARCHAR(500);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove nome_social column
ALTER TABLE citizen_snapshots DROP COLUMN IF EXISTS nome_social;

-- +goose StatementEnd
