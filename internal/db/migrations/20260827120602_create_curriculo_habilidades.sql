-- +goose Up
-- +goose StatementBegin

-- 1. Tabela de relacionamento entre currículo (CPF) e habilidades
CREATE TABLE emp_curriculo_habilidades (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    cpf CHAR(11) NOT NULL,
    id_habilidade BIGINT NOT NULL REFERENCES emp_habilidades(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT uk_emp_curriculo_habilidades_cpf_habilidade UNIQUE (cpf, id_habilidade)
);

-- 2. Criação da Trigger vinculada à tabela para o updated_at
CREATE TRIGGER update_emp_curriculo_habilidades_updated_at
    BEFORE UPDATE ON emp_curriculo_habilidades
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS update_emp_curriculo_habilidades_updated_at ON emp_curriculo_habilidades;
DROP TABLE IF EXISTS emp_curriculo_habilidades;

-- +goose StatementEnd