-- +goose Up
-- +goose StatementBegin

-- Zonas (região geográfica de elegibilidade da vaga)
CREATE TABLE emp_zonas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    descricao VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO emp_zonas (descricao) VALUES
    ('Zona Norte'),
    ('Zona Sul'),
    ('Zona Oeste'),
    ('Centro');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS emp_zonas;

-- +goose StatementEnd
