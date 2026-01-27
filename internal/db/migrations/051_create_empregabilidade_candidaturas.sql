-- +goose Up
-- +goose StatementBegin

-- Candidaturas
CREATE TABLE emp_candidaturas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cpf VARCHAR(14) NOT NULL,
    id_vaga UUID NOT NULL REFERENCES emp_vagas(id) ON DELETE CASCADE,
    status VARCHAR(100) NOT NULL DEFAULT 'candidatura_enviada',
    id_etapa_atual UUID REFERENCES emp_etapas(id),
    respostas_info_complementares JSONB, -- array de {id_info, resposta}
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(cpf, id_vaga)
);

CREATE INDEX idx_emp_candidaturas_cpf ON emp_candidaturas(cpf);
CREATE INDEX idx_emp_candidaturas_vaga ON emp_candidaturas(id_vaga);
CREATE INDEX idx_emp_candidaturas_status ON emp_candidaturas(status);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_emp_candidaturas_status;
DROP INDEX IF EXISTS idx_emp_candidaturas_vaga;
DROP INDEX IF EXISTS idx_emp_candidaturas_cpf;
DROP TABLE IF EXISTS emp_candidaturas;

-- +goose StatementEnd
