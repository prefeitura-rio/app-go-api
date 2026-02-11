-- +goose Up
-- +goose StatementBegin

ALTER TABLE emp_candidaturas ADD COLUMN nome VARCHAR(255);
ALTER TABLE emp_candidaturas ADD COLUMN email VARCHAR(255);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE emp_candidaturas DROP COLUMN IF EXISTS email;
ALTER TABLE emp_candidaturas DROP COLUMN IF EXISTS nome;

-- +goose StatementEnd
