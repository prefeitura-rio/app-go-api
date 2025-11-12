-- +goose Up
-- Add unique constraint to prevent duplicate proposals from same empresa for same opportunity
-- This ensures data integrity at database level in addition to application-level checks

-- First, remove any existing duplicates (keep the oldest one)
DELETE FROM propostas_mei a
WHERE a.id NOT IN (
    SELECT MIN(id)
    FROM propostas_mei
    GROUP BY oportunidade_mei_id, mei_empresa_id
);

-- Now add the unique constraint
ALTER TABLE propostas_mei
ADD CONSTRAINT unique_proposta_per_empresa_oportunidade
UNIQUE (oportunidade_mei_id, mei_empresa_id);

-- +goose Down
-- Remove the unique constraint
ALTER TABLE propostas_mei
DROP CONSTRAINT IF EXISTS unique_proposta_per_empresa_oportunidade;
