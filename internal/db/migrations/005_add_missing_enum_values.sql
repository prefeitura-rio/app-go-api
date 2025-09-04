-- +goose Up

-- Add missing enum values (outside transaction)
ALTER TYPE status_curso_enum ADD VALUE IF NOT EXISTS 'draft';
ALTER TYPE status_curso_enum ADD VALUE IF NOT EXISTS 'opened';
ALTER TYPE status_curso_enum ADD VALUE IF NOT EXISTS 'closed';
ALTER TYPE status_curso_enum ADD VALUE IF NOT EXISTS 'canceled';

-- +goose Down
-- No rollback needed - enum values cannot be removed easily