-- +goose Up
-- Convert single CNAE field to array of CNAEs for OportunidadeMEI
-- This allows opportunities to accept proposals from multiple CNAE codes

-- Add new array column for CNAEs
ALTER TABLE oportunidades_mei ADD COLUMN cnae_ids VARCHAR(20)[] DEFAULT '{}';

-- Migrate existing single CNAE to array (if not null/empty)
UPDATE oportunidades_mei
SET cnae_ids = ARRAY[cnae_id]::VARCHAR(20)[]
WHERE cnae_id IS NOT NULL AND cnae_id != '';

-- Drop the old single CNAE column
ALTER TABLE oportunidades_mei DROP COLUMN cnae_id;

-- Make cnae_ids not null with default empty array
ALTER TABLE oportunidades_mei ALTER COLUMN cnae_ids SET NOT NULL;

-- Add index for array contains queries (useful for filtering)
CREATE INDEX idx_oportunidades_mei_cnae_ids ON oportunidades_mei USING GIN(cnae_ids);

-- +goose Down
-- Restore single CNAE column
ALTER TABLE oportunidades_mei ADD COLUMN cnae_id VARCHAR(20);

-- Take first CNAE from array if exists
UPDATE oportunidades_mei
SET cnae_id = cnae_ids[1]
WHERE array_length(cnae_ids, 1) > 0;

-- Drop array column and index
DROP INDEX IF EXISTS idx_oportunidades_mei_cnae_ids;
ALTER TABLE oportunidades_mei DROP COLUMN cnae_ids;

-- Restore not null constraint
ALTER TABLE oportunidades_mei ALTER COLUMN cnae_id SET NOT NULL;
