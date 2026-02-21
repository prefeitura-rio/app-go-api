-- +goose Up
-- +goose StatementBegin
INSERT INTO categorias (nome) VALUES ('Rio do Futuro')
ON CONFLICT (nome) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM categorias WHERE nome = 'Rio do Futuro';
-- +goose StatementEnd
