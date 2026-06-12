-- +goose Up
-- +goose StatementBegin
ALTER TABLE orgao_snapshots ADD COLUMN IF NOT EXISTS cd_ua VARCHAR(50);
CREATE INDEX IF NOT EXISTS idx_orgao_snapshots_cd_ua ON orgao_snapshots(cd_ua);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_orgao_snapshots_cd_ua;
ALTER TABLE orgao_snapshots DROP COLUMN IF EXISTS cd_ua;
-- +goose StatementEnd