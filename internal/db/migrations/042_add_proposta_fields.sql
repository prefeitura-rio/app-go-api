-- +goose Up
-- Add new fields to propostas_mei table:
-- - prazo_execucao: execution deadline text (e.g., "30 dias", "2 semanas")
-- - aceita_custos_integrais: boolean indicating if the proposal accepts full costs coverage

ALTER TABLE propostas_mei
ADD COLUMN prazo_execucao VARCHAR(255),
ADD COLUMN aceita_custos_integrais BOOLEAN;

-- +goose Down
-- Remove the new fields
ALTER TABLE propostas_mei
DROP COLUMN IF EXISTS prazo_execucao,
DROP COLUMN IF EXISTS aceita_custos_integrais;
