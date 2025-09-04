-- +goose Up
-- +goose StatementBegin

-- Remove old course data that uses legacy enum values
-- This allows us to start fresh with the new enum values
DELETE FROM cursos_categorias;
DELETE FROM cursos_acessibilidades;
DELETE FROM cursos;

-- Reset auto-increment sequence
ALTER SEQUENCE cursos_id_seq RESTART WITH 1;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Cannot restore deleted data, this migration is destructive
-- Consider backing up data before running this migration

-- +goose StatementEnd