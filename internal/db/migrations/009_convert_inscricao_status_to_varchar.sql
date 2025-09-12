-- +goose Up
-- +goose StatementBegin

-- Converter coluna status de ENUM para VARCHAR para permitir novos valores como 'concluded'
ALTER TABLE inscricoes 
    ALTER COLUMN status TYPE VARCHAR(50) USING status::text;

-- Adicionar constraint de check para validar valores permitidos
ALTER TABLE inscricoes 
    ADD CONSTRAINT check_inscricao_status 
    CHECK (status IN ('pending', 'approved', 'rejected', 'cancelled', 'concluded'));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remover constraint
ALTER TABLE inscricoes DROP CONSTRAINT IF EXISTS check_inscricao_status;

-- Reverter para ENUM (removendo registros com 'concluded' primeiro)
UPDATE inscricoes SET status = 'approved' WHERE status = 'concluded';

-- Recriar o tipo ENUM
CREATE TYPE status_inscricao_enum_temp AS ENUM ('pending', 'approved', 'rejected', 'cancelled');

ALTER TABLE inscricoes 
    ALTER COLUMN status TYPE status_inscricao_enum_temp USING status::status_inscricao_enum_temp;

DROP TYPE IF EXISTS status_inscricao_enum;
ALTER TYPE status_inscricao_enum_temp RENAME TO status_inscricao_enum;

-- +goose StatementEnd