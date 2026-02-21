-- +goose Up
-- +goose StatementBegin

-- Create enum type for acessibilidade PCD
CREATE TYPE acessibilidade_pcd_enum AS ENUM ('para_pcd', 'preferencial_pcd', 'exclusivo_pcd');

-- Add new column
ALTER TABLE emp_vagas ADD COLUMN acessibilidade_pcd acessibilidade_pcd_enum;

-- Migrate existing data: if vaga_pcd was true, set to 'para_pcd'
UPDATE emp_vagas SET acessibilidade_pcd = 'para_pcd' WHERE vaga_pcd = true;

-- Drop old column
ALTER TABLE emp_vagas DROP COLUMN vaga_pcd;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Add back the boolean column
ALTER TABLE emp_vagas ADD COLUMN vaga_pcd BOOLEAN DEFAULT FALSE;

-- Migrate data back: if acessibilidade_pcd is not null, set vaga_pcd to true
UPDATE emp_vagas SET vaga_pcd = true WHERE acessibilidade_pcd IS NOT NULL;

-- Drop the new column
ALTER TABLE emp_vagas DROP COLUMN acessibilidade_pcd;

-- Drop the enum type
DROP TYPE IF EXISTS acessibilidade_pcd_enum;

-- +goose StatementEnd
