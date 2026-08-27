-- +goose Up
-- +goose StatementBegin

-- Tabela de relacionamento entre currículo (CPF) e habilidades
CREATE TABLE emp_curriculo_habilidades (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cpf CHAR(11) NOT NULL,
    id_habilidade UUID NOT NULL REFERENCES emp_habilidades(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraint definida diretamente na criação da tabela
    CONSTRAINT uk_emp_curriculo_habilidades_cpf_habilidade UNIQUE (cpf, id_habilidade)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS emp_curriculo_habilidades;

-- +goose StatementEnd
