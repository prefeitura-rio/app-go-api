-- +goose Up
-- +goose NO TRANSACTION
-- Add performance indexes for high-load optimization
-- These indexes address full table scans on filtered queries
-- CONCURRENTLY requires running outside transaction block

-- Course listings (hottest endpoint)
-- Composite index for status + visibility filtering
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_cursos_status_visible
ON cursos(status, is_visible)
WHERE is_visible = true;

-- Index for organization filtering
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_cursos_orgao_status
ON cursos(orgao_id, status);

-- Index for modalidade filtering
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_cursos_modalidade
ON cursos(modalidade)
WHERE modalidade IS NOT NULL;

-- Index for sorting by creation date
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_cursos_created_at
ON cursos(created_at DESC);

-- Job listings
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_empregos_status_orgao
ON empregos(status, orgao_id);

-- Enrollments (user dashboard - high frequency)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_inscricoes_cpf_status
ON inscricoes(cpf, status);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_inscricoes_curso_id
ON inscricoes(curso_id);

-- MEI opportunities
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_oportunidades_mei_status
ON oportunidades_mei(status);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_oportunidades_mei_orgao
ON oportunidades_mei(orgao_id);

-- Foreign key indexes for JOIN optimization
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_empregos_empresa_id
ON empregos(empresa_id)
WHERE empresa_id IS NOT NULL;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_empregos_escolaridade_id
ON empregos(escolaridade_id)
WHERE escolaridade_id IS NOT NULL;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_cursos_instituicao_id
ON cursos(instituicao_id)
WHERE instituicao_id IS NOT NULL;

-- +goose Down
-- Remove performance indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_cursos_status_visible;
DROP INDEX CONCURRENTLY IF EXISTS idx_cursos_orgao_status;
DROP INDEX CONCURRENTLY IF EXISTS idx_cursos_modalidade;
DROP INDEX CONCURRENTLY IF EXISTS idx_cursos_created_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_empregos_status_orgao;
DROP INDEX CONCURRENTLY IF EXISTS idx_inscricoes_cpf_status;
DROP INDEX CONCURRENTLY IF EXISTS idx_inscricoes_curso_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_oportunidades_mei_status;
DROP INDEX CONCURRENTLY IF EXISTS idx_oportunidades_mei_orgao;
DROP INDEX CONCURRENTLY IF EXISTS idx_empregos_empresa_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_empregos_escolaridade_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_cursos_instituicao_id;
