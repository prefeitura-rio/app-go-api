-- +goose Up
-- +goose StatementBegin
ALTER TABLE emp_candidaturas
ADD COLUMN IF NOT EXISTS curriculo_snapshot JSONB;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE emp_candidaturas
DROP COLUMN IF EXISTS curriculo_snapshot;
-- +goose StatementEnd
