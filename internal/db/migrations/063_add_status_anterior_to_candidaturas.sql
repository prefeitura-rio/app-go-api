-- +goose Up
ALTER TABLE emp_candidaturas ADD COLUMN status_anterior VARCHAR(100);

-- +goose Down
ALTER TABLE emp_candidaturas DROP COLUMN status_anterior;
