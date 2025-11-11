-- +goose Up
-- +goose StatementBegin

-- This migration fixes the incomplete 024 migration
-- It's idempotent and can be run safely even if some steps were already completed

-- Step 1: Drop ALL foreign key constraints that reference orgaos table (if it still exists)
DO $$
DECLARE
    constraint_record RECORD;
    table_exists BOOLEAN;
BEGIN
    -- Check if orgaos table exists
    SELECT EXISTS (
        SELECT FROM pg_tables
        WHERE schemaname = 'public'
        AND tablename = 'orgaos'
    ) INTO table_exists;

    IF table_exists THEN
        -- Drop all constraints referencing orgaos
        FOR constraint_record IN
            SELECT conname, conrelid::regclass AS table_name
            FROM pg_constraint
            WHERE confrelid = 'orgaos'::regclass
        LOOP
            EXECUTE format('ALTER TABLE %s DROP CONSTRAINT IF EXISTS %I',
                          constraint_record.table_name,
                          constraint_record.conname);
        END LOOP;
    END IF;
END $$;

-- Step 2: Drop indexes on orgao_id columns (if they still exist with old names)
DROP INDEX IF EXISTS idx_cursos_orgao;
DROP INDEX IF EXISTS idx_empregos_orgao;
DROP INDEX IF EXISTS idx_oportunidades_mei_orgao;

-- Step 3: Convert orgao_id columns from INTEGER to VARCHAR(100) if not already converted
DO $$
BEGIN
    -- Check and convert cursos.orgao_id
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'cursos'
        AND column_name = 'orgao_id'
        AND data_type = 'integer'
    ) THEN
        ALTER TABLE cursos ALTER COLUMN orgao_id TYPE VARCHAR(100) USING CASE
            WHEN orgao_id IS NULL THEN NULL
            ELSE orgao_id::TEXT
        END;
    END IF;

    -- Check and convert empregos.orgao_id
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'empregos'
        AND column_name = 'orgao_id'
        AND data_type = 'integer'
    ) THEN
        ALTER TABLE empregos ALTER COLUMN orgao_id TYPE VARCHAR(100) USING CASE
            WHEN orgao_id IS NULL THEN NULL
            ELSE orgao_id::TEXT
        END;
    END IF;

    -- Check and convert oportunidades_mei.orgao_id
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'oportunidades_mei'
        AND column_name = 'orgao_id'
        AND data_type = 'integer'
    ) THEN
        ALTER TABLE oportunidades_mei ALTER COLUMN orgao_id DROP NOT NULL;
        ALTER TABLE oportunidades_mei ALTER COLUMN orgao_id TYPE VARCHAR(100) USING orgao_id::TEXT;
        ALTER TABLE oportunidades_mei ALTER COLUMN orgao_id SET NOT NULL;
    END IF;
END $$;

-- Step 4: Recreate indexes for performance (only if they don't exist)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_indexes
        WHERE indexname = 'idx_cursos_orgao'
    ) THEN
        CREATE INDEX idx_cursos_orgao ON cursos(orgao_id) WHERE orgao_id IS NOT NULL;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_indexes
        WHERE indexname = 'idx_empregos_orgao'
    ) THEN
        CREATE INDEX idx_empregos_orgao ON empregos(orgao_id) WHERE orgao_id IS NOT NULL;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_indexes
        WHERE indexname = 'idx_oportunidades_mei_orgao'
    ) THEN
        CREATE INDEX idx_oportunidades_mei_orgao ON oportunidades_mei(orgao_id);
    END IF;
END $$;

-- Step 5: Drop the orgaos table (if it still exists)
DROP TABLE IF EXISTS orgaos CASCADE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- This is a fix migration, down migration is not applicable
-- If you need to rollback, you should rollback the original 024 migration

-- +goose StatementEnd
