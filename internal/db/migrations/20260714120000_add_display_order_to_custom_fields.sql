-- +goose Up
-- +goose StatementBegin

-- Add display_order column to custom_fields.
-- IF NOT EXISTS makes this resilient to GORM AutoMigrate, which may add the column
-- first (as a plain default-0 column, without the backfill below). In that case goose
-- skips the add, still runs the backfill, and creates the index — self-healing.
ALTER TABLE custom_fields ADD COLUMN IF NOT EXISTS display_order INTEGER NOT NULL DEFAULT 0;

-- Backfill display_order for existing custom_fields using current insertion order
UPDATE custom_fields cf
SET display_order = sub.rn
FROM (
    SELECT id, ROW_NUMBER() OVER (PARTITION BY curso_id ORDER BY created_at, id) AS rn
    FROM custom_fields
) sub
WHERE cf.id = sub.id;

-- Create index for ordering queries
CREATE INDEX IF NOT EXISTS idx_custom_fields_curso_display_order ON custom_fields(curso_id, display_order);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_custom_fields_curso_display_order;

ALTER TABLE custom_fields DROP COLUMN IF EXISTS display_order;

-- +goose StatementEnd
