-- +goose Up
-- +goose StatementBegin

-- This migration fixes the incomplete 025 migration
-- It's idempotent and can be run safely even if some steps were already completed

-- Step 1: Drop ALL foreign key constraints that reference cnaes or mei_empresas tables (if they still exist)
DO $$
DECLARE
    constraint_record RECORD;
    cnaes_exists BOOLEAN;
    mei_empresas_exists BOOLEAN;
BEGIN
    -- Check if tables exist
    SELECT EXISTS (
        SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'cnaes'
    ) INTO cnaes_exists;

    SELECT EXISTS (
        SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'mei_empresas'
    ) INTO mei_empresas_exists;

    -- Drop constraints referencing cnaes
    IF cnaes_exists THEN
        FOR constraint_record IN
            SELECT conname, conrelid::regclass AS table_name
            FROM pg_constraint
            WHERE confrelid = 'cnaes'::regclass
        LOOP
            EXECUTE format('ALTER TABLE %s DROP CONSTRAINT IF EXISTS %I',
                          constraint_record.table_name,
                          constraint_record.conname);
        END LOOP;
    END IF;

    -- Drop constraints referencing mei_empresas
    IF mei_empresas_exists THEN
        FOR constraint_record IN
            SELECT conname, conrelid::regclass AS table_name
            FROM pg_constraint
            WHERE confrelid = 'mei_empresas'::regclass
        LOOP
            EXECUTE format('ALTER TABLE %s DROP CONSTRAINT IF EXISTS %I',
                          constraint_record.table_name,
                          constraint_record.conname);
        END LOOP;
    END IF;
END $$;

-- Step 2: Drop many-to-many join table
DROP TABLE IF EXISTS mei_empresas_cnaes CASCADE;

-- Step 3: Convert oportunidades_mei.cnae_id from INTEGER to VARCHAR(20)
-- Using a temp column approach to avoid subquery in USING clause
DO $$
DECLARE
    cnaes_exists BOOLEAN;
BEGIN
    -- Check if cnaes table exists
    SELECT EXISTS (
        SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'cnaes'
    ) INTO cnaes_exists;

    -- Check if cnae_id is still INTEGER
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'oportunidades_mei'
        AND column_name = 'cnae_id'
        AND data_type = 'integer'
    ) THEN
        -- Add temporary column
        ALTER TABLE oportunidades_mei ADD COLUMN cnae_id_temp VARCHAR(20);

        -- Populate temp column with CNAE codes if cnaes table exists
        IF cnaes_exists THEN
            UPDATE oportunidades_mei
            SET cnae_id_temp = (SELECT codigo FROM cnaes WHERE id = oportunidades_mei.cnae_id LIMIT 1)
            WHERE cnae_id IS NOT NULL;
        END IF;

        -- Drop old column and rename temp
        ALTER TABLE oportunidades_mei DROP COLUMN cnae_id;
        ALTER TABLE oportunidades_mei RENAME COLUMN cnae_id_temp TO cnae_id;
    END IF;
END $$;

-- Step 4: Convert propostas_mei.mei_empresa_id from INTEGER to VARCHAR(18)
DO $$
DECLARE
    mei_empresas_exists BOOLEAN;
BEGIN
    -- Check if mei_empresas table exists
    SELECT EXISTS (
        SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'mei_empresas'
    ) INTO mei_empresas_exists;

    -- Check if mei_empresa_id is still INTEGER
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'propostas_mei'
        AND column_name = 'mei_empresa_id'
        AND data_type = 'integer'
    ) THEN
        -- Add temporary column
        ALTER TABLE propostas_mei ADD COLUMN mei_empresa_id_temp VARCHAR(18);

        -- Populate temp column with CNPJs if mei_empresas table exists
        IF mei_empresas_exists THEN
            UPDATE propostas_mei
            SET mei_empresa_id_temp = (SELECT cnpj FROM mei_empresas WHERE id = propostas_mei.mei_empresa_id LIMIT 1)
            WHERE mei_empresa_id IS NOT NULL;
        END IF;

        -- Drop old column and rename temp
        ALTER TABLE propostas_mei DROP COLUMN mei_empresa_id;
        ALTER TABLE propostas_mei RENAME COLUMN mei_empresa_id_temp TO mei_empresa_id;
    END IF;
END $$;

-- Step 5: Drop the CNAE and MEI Empresa tables
DROP TABLE IF EXISTS cnaes CASCADE;
DROP TABLE IF EXISTS mei_empresas CASCADE;

-- Step 6: Recreate indexes on the converted columns (only if they don't exist)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_indexes
        WHERE indexname = 'idx_oportunidades_mei_cnae'
    ) THEN
        CREATE INDEX idx_oportunidades_mei_cnae ON oportunidades_mei(cnae_id);
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_indexes
        WHERE indexname = 'idx_propostas_mei_empresa'
    ) THEN
        CREATE INDEX idx_propostas_mei_empresa ON propostas_mei(mei_empresa_id);
    END IF;
END $$;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- This is a fix migration, down migration is not applicable
-- If you need to rollback, you should rollback the original 025 migration

-- +goose StatementEnd
