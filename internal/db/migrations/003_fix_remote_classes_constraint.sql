-- +goose Up
-- +goose StatementBegin

-- Check if the unique constraint exists, if not, create it
DO $$
BEGIN
    -- First check if the constraint already exists
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE table_name = 'remote_classes' 
        AND constraint_name = 'idx_remote_classes_curso_id' 
        AND table_schema = CURRENT_SCHEMA()
    ) THEN
        -- Create unique index if it doesn't exist
        CREATE UNIQUE INDEX IF NOT EXISTS idx_remote_classes_curso_id ON remote_classes(curso_id);
    END IF;
END $$;

-- Ensure the tables have proper UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Update any existing NULL UUIDs (if any)
UPDATE inscricoes SET id = gen_random_uuid() WHERE id IS NULL;
UPDATE custom_fields SET id = gen_random_uuid() WHERE id IS NULL;
UPDATE location_classes SET id = gen_random_uuid() WHERE id IS NULL;
UPDATE remote_classes SET id = gen_random_uuid() WHERE id IS NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop the unique index
DROP INDEX IF EXISTS idx_remote_classes_curso_id;

-- +goose StatementEnd