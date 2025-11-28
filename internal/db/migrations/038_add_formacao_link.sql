-- +goose Up
-- +goose StatementBegin
ALTER TABLE cursos ADD COLUMN IF NOT EXISTS formacao_link VARCHAR(20000);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE cursos DROP COLUMN IF EXISTS formacao_link;
-- +goose StatementEnd

