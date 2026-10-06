-- +goose Up
ALTER TABLE inscricoes
    ADD COLUMN IF NOT EXISTS origem_inscricao VARCHAR(64) NULL;

-- +goose Down
ALTER TABLE inscricoes
    DROP COLUMN IF EXISTS origem_inscricao;
