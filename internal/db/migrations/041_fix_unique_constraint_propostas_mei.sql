-- +goose Up
-- Replace unique constraint with partial unique index to allow resubmitting proposals after deletion
-- The old constraint prevented the same empresa from creating a new proposal for the same opportunity
-- even after deleting their previous proposal (soft delete)

-- Drop the old unique constraint
ALTER TABLE propostas_mei
DROP CONSTRAINT IF EXISTS unique_proposta_per_empresa_oportunidade;

-- Create a partial unique index that only applies to non-deleted records
CREATE UNIQUE INDEX unique_active_proposta_per_empresa_oportunidade
ON propostas_mei (oportunidade_mei_id, mei_empresa_id)
WHERE deleted_at IS NULL;

-- +goose Down
-- Remove the partial unique index
DROP INDEX IF EXISTS unique_active_proposta_per_empresa_oportunidade;

-- Restore the old unique constraint
ALTER TABLE propostas_mei
ADD CONSTRAINT unique_proposta_per_empresa_oportunidade
UNIQUE (oportunidade_mei_id, mei_empresa_id);
