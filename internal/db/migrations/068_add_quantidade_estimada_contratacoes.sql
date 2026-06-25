-- +goose Up
-- +goose StatementBegin
ALTER TABLE emp_vagas
    ADD COLUMN quantidade_estimada_contratacoes INTEGER;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE emp_vagas
    DROP COLUMN IF EXISTS quantidade_estimada_contratacoes;
-- +goose StatementEnd
