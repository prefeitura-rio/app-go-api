-- +goose Up
-- +goose StatementBegin

-- Vagas
CREATE TABLE emp_vagas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    titulo VARCHAR(500) NOT NULL,
    descricao TEXT NOT NULL,
    id_contratante VARCHAR(18) NOT NULL REFERENCES emp_empresas(cnpj),
    id_regime_contratacao UUID NOT NULL REFERENCES emp_regimes_contratacao(id),
    id_modelo_trabalho UUID NOT NULL REFERENCES emp_modelos_trabalho(id),
    vaga_pcd BOOLEAN DEFAULT FALSE,
    valor_vaga DECIMAL(12,2),
    bairro VARCHAR(255),
    data_limite TIMESTAMP WITH TIME ZONE,
    requisitos TEXT,
    diferenciais TEXT,
    responsabilidades TEXT,
    beneficios TEXT,
    id_orgao_parceiro VARCHAR(50),
    status VARCHAR(50) NOT NULL DEFAULT 'em_edicao',
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- M:N Vagas <-> Tipos PCD
CREATE TABLE emp_vagas_tipos_pcd (
    id_vaga UUID REFERENCES emp_vagas(id) ON DELETE CASCADE,
    id_tipo_pcd UUID REFERENCES emp_tipos_pcd(id) ON DELETE CASCADE,
    PRIMARY KEY (id_vaga, id_tipo_pcd)
);

CREATE INDEX idx_emp_vagas_contratante ON emp_vagas(id_contratante);
CREATE INDEX idx_emp_vagas_status ON emp_vagas(status);
CREATE INDEX idx_emp_vagas_data_limite ON emp_vagas(data_limite);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_emp_vagas_data_limite;
DROP INDEX IF EXISTS idx_emp_vagas_status;
DROP INDEX IF EXISTS idx_emp_vagas_contratante;
DROP TABLE IF EXISTS emp_vagas_tipos_pcd;
DROP TABLE IF EXISTS emp_vagas;

-- +goose StatementEnd
