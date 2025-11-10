-- +goose Up
-- +goose StatementBegin

-- Step 1: Drop foreign key constraints from all tables
ALTER TABLE cursos DROP CONSTRAINT IF EXISTS cursos_orgao_id_fkey;
ALTER TABLE empregos DROP CONSTRAINT IF EXISTS empregos_orgao_id_fkey;
ALTER TABLE oportunidades_mei DROP CONSTRAINT IF EXISTS oportunidades_mei_orgao_id_fkey;

-- Step 2: Drop indexes on orgao_id columns
DROP INDEX IF EXISTS idx_cursos_orgao;
DROP INDEX IF EXISTS idx_empregos_orgao;
DROP INDEX IF EXISTS idx_oportunidades_mei_orgao;

-- Step 3: Convert orgao_id columns from INTEGER to VARCHAR(100)
-- For cursos (nullable)
ALTER TABLE cursos ALTER COLUMN orgao_id TYPE VARCHAR(100) USING CASE
    WHEN orgao_id IS NULL THEN NULL
    ELSE orgao_id::TEXT
END;

-- For empregos (nullable)
ALTER TABLE empregos ALTER COLUMN orgao_id TYPE VARCHAR(100) USING CASE
    WHEN orgao_id IS NULL THEN NULL
    ELSE orgao_id::TEXT
END;

-- For oportunidades_mei (not nullable, has default constraint to handle)
ALTER TABLE oportunidades_mei ALTER COLUMN orgao_id DROP NOT NULL;
ALTER TABLE oportunidades_mei ALTER COLUMN orgao_id TYPE VARCHAR(100) USING orgao_id::TEXT;
ALTER TABLE oportunidades_mei ALTER COLUMN orgao_id SET NOT NULL;

-- Step 4: Recreate indexes for performance (using VARCHAR now)
CREATE INDEX idx_cursos_orgao ON cursos(orgao_id) WHERE orgao_id IS NOT NULL;
CREATE INDEX idx_empregos_orgao ON empregos(orgao_id) WHERE orgao_id IS NOT NULL;
CREATE INDEX idx_oportunidades_mei_orgao ON oportunidades_mei(orgao_id);

-- Step 5: Drop the orgaos table (no longer needed)
DROP TABLE IF EXISTS orgaos CASCADE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Recreate orgaos table
CREATE TABLE orgaos (
    id SERIAL PRIMARY KEY,
    nome VARCHAR(20000) NOT NULL
);

-- Drop VARCHAR indexes
DROP INDEX IF EXISTS idx_cursos_orgao;
DROP INDEX IF EXISTS idx_empregos_orgao;
DROP INDEX IF EXISTS idx_oportunidades_mei_orgao;

-- Convert orgao_id columns back from VARCHAR to INTEGER
-- Note: This assumes orgao_id values are numeric strings that can be converted back
ALTER TABLE cursos ALTER COLUMN orgao_id TYPE INTEGER USING CASE
    WHEN orgao_id IS NULL THEN NULL
    WHEN orgao_id ~ '^[0-9]+$' THEN orgao_id::INTEGER
    ELSE NULL
END;

ALTER TABLE empregos ALTER COLUMN orgao_id TYPE INTEGER USING CASE
    WHEN orgao_id IS NULL THEN NULL
    WHEN orgao_id ~ '^[0-9]+$' THEN orgao_id::INTEGER
    ELSE NULL
END;

ALTER TABLE oportunidades_mei ALTER COLUMN orgao_id DROP NOT NULL;
ALTER TABLE oportunidades_mei ALTER COLUMN orgao_id TYPE INTEGER USING CASE
    WHEN orgao_id ~ '^[0-9]+$' THEN orgao_id::INTEGER
    ELSE 0
END;
ALTER TABLE oportunidades_mei ALTER COLUMN orgao_id SET NOT NULL;

-- Recreate foreign key constraints
ALTER TABLE cursos ADD CONSTRAINT cursos_orgao_id_fkey FOREIGN KEY (orgao_id) REFERENCES orgaos(id);
ALTER TABLE empregos ADD CONSTRAINT empregos_orgao_id_fkey FOREIGN KEY (orgao_id) REFERENCES orgaos(id);
ALTER TABLE oportunidades_mei ADD CONSTRAINT oportunidades_mei_orgao_id_fkey FOREIGN KEY (orgao_id) REFERENCES orgaos(id);

-- Recreate indexes
CREATE INDEX idx_cursos_orgao ON cursos(orgao_id);
CREATE INDEX idx_empregos_orgao ON empregos(orgao_id);
CREATE INDEX idx_oportunidades_mei_orgao ON oportunidades_mei(orgao_id);

-- +goose StatementEnd
