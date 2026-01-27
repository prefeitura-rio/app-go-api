-- +goose Up
-- +goose StatementBegin

-- Etapas do processo seletivo
CREATE TABLE emp_etapas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    id_vaga UUID NOT NULL REFERENCES emp_vagas(id) ON DELETE CASCADE,
    titulo VARCHAR(500) NOT NULL,
    descricao TEXT,
    ordem INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_emp_etapas_vaga ON emp_etapas(id_vaga);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_emp_etapas_vaga;
DROP TABLE IF EXISTS emp_etapas;

-- +goose StatementEnd
