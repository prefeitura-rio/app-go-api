-- +goose Up
-- +goose StatementBegin

-- Add published_at field to cursos table (timestamp as integer/bigint for Unix timestamp)
ALTER TABLE cursos ADD COLUMN IF NOT EXISTS published_at BIGINT;

-- Add is_free field to cursos table
ALTER TABLE cursos ADD COLUMN IF NOT EXISTS is_free BOOLEAN;

-- Add comments for documentation
COMMENT ON COLUMN cursos.published_at IS 'Unix timestamp when the course was published';
COMMENT ON COLUMN cursos.is_free IS 'Indicates if the course is free of charge';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove the added columns
ALTER TABLE cursos DROP COLUMN IF EXISTS published_at;
ALTER TABLE cursos DROP COLUMN IF EXISTS is_free;

-- +goose StatementEnd