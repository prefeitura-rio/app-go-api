-- +goose Up
-- +goose StatementBegin

-- 1. Extensões e Função Imutável para Busca Textual (Garantia de Existência)
CREATE EXTENSION IF NOT EXISTS unaccent;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE OR REPLACE FUNCTION immutable_unaccent(text)
RETURNS text AS $$
    SELECT public.unaccent($1);
$$ LANGUAGE sql IMMUTABLE PARALLEL SAFE STRICT;

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 2. Tabela de comportamentos e atitudes
CREATE TABLE emp_comportamento_atitudes (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    nome VARCHAR(250) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 3. Trigger vinculada à tabela
CREATE TRIGGER update_comportamento_atitudes_updated_at
    BEFORE UPDATE ON emp_comportamento_atitudes
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 4. Índice de Busca Textual corrigido para a tabela correta
CREATE INDEX IF NOT EXISTS idx_emp_comportamento_atitudes_nome_unaccent_trgm 
ON emp_comportamento_atitudes 
USING gin (lower(immutable_unaccent(nome)) gin_trgm_ops);

-- 5. Carga de dados inicial
INSERT INTO emp_comportamento_atitudes (nome) VALUES
('Proatividade'),
('Organização'),
('Liderança'),
('Resolução de problemas'),
('Pensamento analítico'),
('Responsabilidade'),
('Empatia'),
('Inteligência emocional')
ON CONFLICT (nome) DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_emp_comportamento_atitudes_nome_unaccent_trgm;
DROP TRIGGER IF EXISTS update_comportamento_atitudes_updated_at ON emp_comportamento_atitudes;
DROP TABLE IF EXISTS emp_comportamento_atitudes;

-- NOTA: Funções globais (immutable_unaccent, update_updated_at_column) 
-- e extensões NÃO devem ser excluídas no Down pois servem a outras tabelas.

-- +goose StatementEnd