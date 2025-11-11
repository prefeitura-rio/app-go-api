-- +goose Up
-- +goose StatementBegin

-- Step 1: Drop ALL foreign key constraints that reference cnaes or mei_empresas tables
DO $$
DECLARE
    constraint_record RECORD;
BEGIN
    -- Drop constraints referencing cnaes
    FOR constraint_record IN
        SELECT conname, conrelid::regclass AS table_name
        FROM pg_constraint
        WHERE confrelid = 'cnaes'::regclass
    LOOP
        EXECUTE format('ALTER TABLE %s DROP CONSTRAINT IF EXISTS %I',
                      constraint_record.table_name,
                      constraint_record.conname);
    END LOOP;

    -- Drop constraints referencing mei_empresas
    FOR constraint_record IN
        SELECT conname, conrelid::regclass AS table_name
        FROM pg_constraint
        WHERE confrelid = 'mei_empresas'::regclass
    LOOP
        EXECUTE format('ALTER TABLE %s DROP CONSTRAINT IF EXISTS %I',
                      constraint_record.table_name,
                      constraint_record.conname);
    END LOOP;
END $$;

-- Step 2: Drop many-to-many join table
DROP TABLE IF EXISTS mei_empresas_cnaes CASCADE;

-- Step 3: Convert oportunidades_mei.cnae_id from INTEGER to VARCHAR(20)
-- This will store the CNAE code (e.g., "4322-3/01") instead of the database ID
-- Using temp column approach to avoid subquery limitation in USING clause
ALTER TABLE oportunidades_mei ADD COLUMN cnae_id_temp VARCHAR(20);

UPDATE oportunidades_mei
SET cnae_id_temp = (SELECT codigo FROM cnaes WHERE id = oportunidades_mei.cnae_id LIMIT 1)
WHERE cnae_id IS NOT NULL;

ALTER TABLE oportunidades_mei DROP COLUMN cnae_id;
ALTER TABLE oportunidades_mei RENAME COLUMN cnae_id_temp TO cnae_id;

-- Step 4: Convert propostas_mei.mei_empresa_id from INTEGER to VARCHAR(18)
-- This will store the CNPJ (e.g., "12.345.678/0001-90") instead of the database ID
-- Using temp column approach to avoid subquery limitation in USING clause
ALTER TABLE propostas_mei ADD COLUMN mei_empresa_id_temp VARCHAR(18);

UPDATE propostas_mei
SET mei_empresa_id_temp = (SELECT cnpj FROM mei_empresas WHERE id = propostas_mei.mei_empresa_id LIMIT 1)
WHERE mei_empresa_id IS NOT NULL;

ALTER TABLE propostas_mei DROP COLUMN mei_empresa_id;
ALTER TABLE propostas_mei RENAME COLUMN mei_empresa_id_temp TO mei_empresa_id;

-- Step 5: Drop the CNAE and MEI Empresa tables
DROP TABLE IF EXISTS cnaes CASCADE;
DROP TABLE IF EXISTS mei_empresas CASCADE;

-- Step 6: Recreate indexes on the converted columns
CREATE INDEX idx_oportunidades_mei_cnae ON oportunidades_mei(cnae_id);
CREATE INDEX idx_propostas_mei_empresa ON propostas_mei(mei_empresa_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Note: This migration cannot be fully reversed as it drops tables and converts data
-- The down migration recreates the tables but data will be lost

-- Recreate CNAE table
CREATE TABLE cnaes (
    id SERIAL PRIMARY KEY,
    codigo VARCHAR(20) NOT NULL,
    ocupacao VARCHAR(255) NOT NULL,
    servico VARCHAR(500) NOT NULL,
    UNIQUE(codigo, servico)
);

CREATE INDEX idx_cnaes_codigo ON cnaes(codigo);
CREATE INDEX idx_cnaes_ocupacao ON cnaes(ocupacao);

-- Recreate MEI Empresas table
CREATE TABLE mei_empresas (
    id SERIAL PRIMARY KEY,
    cnpj VARCHAR(18) UNIQUE NOT NULL,
    razao_social VARCHAR(255) NOT NULL,
    porte VARCHAR(100),
    nome_fantasia VARCHAR(255),
    tipo VARCHAR(100),
    natureza_juridica VARCHAR(100),
    situacao_cadastral VARCHAR(50),
    cep VARCHAR(10),
    logradouro VARCHAR(255),
    numero VARCHAR(20),
    bairro VARCHAR(100),
    cidade VARCHAR(100),
    estado VARCHAR(2),
    email VARCHAR(255),
    whatsapp VARCHAR(20),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_mei_empresas_cnpj ON mei_empresas(cnpj);
CREATE INDEX idx_mei_empresas_situacao ON mei_empresas(situacao_cadastral);

-- Recreate many-to-many join table
CREATE TABLE mei_empresas_cnaes (
    mei_empresa_id INTEGER NOT NULL REFERENCES mei_empresas(id) ON DELETE CASCADE,
    cnae_id INTEGER NOT NULL REFERENCES cnaes(id) ON DELETE CASCADE,
    PRIMARY KEY (mei_empresa_id, cnae_id)
);

CREATE INDEX idx_mei_empresas_cnaes_mei ON mei_empresas_cnaes(mei_empresa_id);
CREATE INDEX idx_mei_empresas_cnaes_cnae ON mei_empresas_cnaes(cnae_id);

-- Drop indexes on VARCHAR columns
DROP INDEX IF EXISTS idx_oportunidades_mei_cnae;
DROP INDEX IF EXISTS idx_propostas_mei_empresa;

-- Convert back to INTEGER (data will be lost - this is just structure)
ALTER TABLE oportunidades_mei ALTER COLUMN cnae_id TYPE INTEGER USING NULL;
ALTER TABLE propostas_mei ALTER COLUMN mei_empresa_id TYPE INTEGER USING NULL;

-- Recreate foreign keys
ALTER TABLE oportunidades_mei ADD CONSTRAINT oportunidades_mei_cnae_id_fkey
    FOREIGN KEY (cnae_id) REFERENCES cnaes(id);

ALTER TABLE propostas_mei ADD CONSTRAINT propostas_mei_mei_empresa_id_fkey
    FOREIGN KEY (mei_empresa_id) REFERENCES mei_empresas(id) ON DELETE CASCADE;

-- Recreate indexes
CREATE INDEX idx_oportunidades_mei_cnae ON oportunidades_mei(cnae_id);
CREATE INDEX idx_propostas_mei_empresa ON propostas_mei(mei_empresa_id);

-- +goose StatementEnd
