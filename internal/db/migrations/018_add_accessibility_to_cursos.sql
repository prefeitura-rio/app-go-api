-- +goose Up
-- +goose StatementBegin
ALTER TABLE cursos ADD COLUMN accessibility VARCHAR(255);
COMMENT ON COLUMN cursos.accessibility IS 'Campo de texto livre para informações de acessibilidade do curso';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE cursos DROP COLUMN accessibility;
-- +goose StatementEnd
