-- +goose Up
-- +goose StatementBegin

-- Empresas
CREATE TABLE emp_empresas (
    cnpj VARCHAR(18) PRIMARY KEY,
    razao_social VARCHAR(500),
    nome_fantasia VARCHAR(500),
    descricao TEXT,
    url_logo VARCHAR(1000),
    website VARCHAR(1000),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_emp_empresas_razao ON emp_empresas(razao_social);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_emp_empresas_razao;
DROP TABLE IF EXISTS emp_empresas;

-- +goose StatementEnd
