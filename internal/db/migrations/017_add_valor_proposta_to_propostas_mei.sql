-- +goose Up
ALTER TABLE propostas_mei ADD COLUMN valor_proposta DECIMAL(10,2);

-- +goose Down
ALTER TABLE propostas_mei DROP COLUMN valor_proposta;
