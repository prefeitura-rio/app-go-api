-- +goose Up
-- +goose StatementBegin

ALTER TABLE cursos ADD COLUMN auto_approve_enrollments BOOLEAN DEFAULT FALSE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE cursos DROP COLUMN IF EXISTS auto_approve_enrollments;

-- +goose StatementEnd
