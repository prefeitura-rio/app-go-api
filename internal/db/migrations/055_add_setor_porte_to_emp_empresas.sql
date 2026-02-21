-- +goose Up
-- +goose StatementBegin
ALTER TABLE emp_empresas ADD COLUMN IF NOT EXISTS setor VARCHAR(500);
ALTER TABLE emp_empresas ADD COLUMN IF NOT EXISTS porte VARCHAR(100);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE emp_empresas DROP COLUMN IF EXISTS setor;
ALTER TABLE emp_empresas DROP COLUMN IF EXISTS porte;
-- +goose StatementEnd
