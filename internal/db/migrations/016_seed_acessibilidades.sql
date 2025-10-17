-- +goose Up
-- +goose StatementBegin

-- Seed initial accessibility options
-- Using ON CONFLICT DO NOTHING for idempotency
INSERT INTO acessibilidades (nome) VALUES
('Acessível para pessoas com deficiência'),
('Exclusivo para pessoas com deficiência'),
('Não acessível para pessoas com deficiência')
ON CONFLICT DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove seeded accessibility options
DELETE FROM acessibilidades
WHERE nome IN (
    'Acessível para pessoas com deficiência',
    'Exclusivo para pessoas com deficiência',
    'Não acessível para pessoas com deficiência'
);

-- +goose StatementEnd
