-- +goose Up
-- +goose StatementBegin

-- Adicionar campo certificate_url na tabela inscricoes
ALTER TABLE inscricoes ADD COLUMN IF NOT EXISTS certificate_url VARCHAR(500);

-- Adicionar novo status 'concluded' para inscrições
-- Como estamos usando VARCHAR ao invés de ENUM (migração 006), não precisamos alterar o tipo
-- Apenas documentamos que 'concluded' é um valor válido

-- Criar índice para otimizar consultas por certificados
CREATE INDEX IF NOT EXISTS idx_inscricoes_certificate ON inscricoes(certificate_url) WHERE certificate_url IS NOT NULL;

-- Adicionar campo concluded_at para registrar quando a inscrição foi concluída
ALTER TABLE inscricoes ADD COLUMN IF NOT EXISTS concluded_at TIMESTAMP WITH TIME ZONE;

-- Criar índice para otimizar consultas por data de conclusão
CREATE INDEX IF NOT EXISTS idx_inscricoes_concluded_at ON inscricoes(concluded_at) WHERE concluded_at IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remover índices
DROP INDEX IF EXISTS idx_inscricoes_concluded_at;
DROP INDEX IF EXISTS idx_inscricoes_certificate;

-- Remover colunas
ALTER TABLE inscricoes DROP COLUMN IF EXISTS concluded_at;
ALTER TABLE inscricoes DROP COLUMN IF EXISTS certificate_url;

-- +goose StatementEnd