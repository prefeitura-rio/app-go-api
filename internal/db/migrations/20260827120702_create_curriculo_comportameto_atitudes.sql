-- +goose Up
-- +goose StatementBegin

-- 1. Criação da tabela pivot entre currículo (CPF) e comportamentos/atitudes
CREATE TABLE emp_curriculo_comportamento_atitudes (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    cpf CHAR(11) NOT NULL,
    id_comportamento_atitudes BIGINT NOT NULL REFERENCES emp_comportamento_atitudes(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT uk_emp_curriculo_comportamento_atitudes_cpf_comportamento UNIQUE (cpf, id_comportamento_atitudes)
);

-- 2. Criação da Trigger vinculada à tabela
CREATE TRIGGER update_emp_curriculo_comportamento_atitudes_updated_at
    BEFORE UPDATE ON emp_curriculo_comportamento_atitudes
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS update_emp_curriculo_comportamento_atitudes_updated_at ON emp_curriculo_comportamento_atitudes;
DROP TABLE IF EXISTS emp_curriculo_comportamento_atitudes;

-- +goose StatementEnd