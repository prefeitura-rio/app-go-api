-- +goose Up
-- +goose StatementBegin
ALTER TABLE cursos ADD COLUMN is_visible BOOLEAN DEFAULT true NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE cursos DROP COLUMN is_visible;
-- +goose StatementEnd
