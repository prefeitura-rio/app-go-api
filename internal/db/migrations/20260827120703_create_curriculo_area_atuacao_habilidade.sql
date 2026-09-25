-- +goose Up
-- +goose StatementBegin

-- 1. Tabela de relacionamento entre currículo (CPF) e habilidades por área
CREATE TABLE IF NOT EXISTS emp_curriculo_area_atuacao_habilidade (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    cpf char(11) NOT NULL,
    id_area_atuacao_habilidade BIGINT NOT NULL REFERENCES area_atuacao_habilidade(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Nome da constraint único para esta tabela
    CONSTRAINT uk_emp_curriculo_aah_cpf_habilidade UNIQUE (cpf, id_area_atuacao_habilidade)
);

-- 2. Trigger para atualização automática do updated_at
CREATE TRIGGER update_emp_curriculo_area_atuacao_habilidade_updated_at
    BEFORE UPDATE ON emp_curriculo_area_atuacao_habilidade
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 3. Índices de performance
CREATE INDEX IF NOT EXISTS idx_emp_curriculo_aah_id_aah 
ON emp_curriculo_area_atuacao_habilidade (id_area_atuacao_habilidade);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- 1. Remoção do Trigger
DROP TRIGGER IF EXISTS update_emp_curriculo_area_atuacao_habilidade_updated_at ON emp_curriculo_area_atuacao_habilidade;

-- 2. Remoção da Tabela
DROP TABLE IF EXISTS emp_curriculo_area_atuacao_habilidade;

-- +goose StatementEnd