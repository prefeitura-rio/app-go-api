-- +goose Up
-- +goose StatementBegin

-- Rename anexo_url to cover_image
ALTER TABLE oportunidades_mei
RENAME COLUMN anexo_url TO cover_image;

-- Add gallery_images column as JSONB array
ALTER TABLE oportunidades_mei
ADD COLUMN gallery_images JSONB DEFAULT '[]'::jsonb;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove gallery_images column
ALTER TABLE oportunidades_mei
DROP COLUMN IF EXISTS gallery_images;

-- Rename cover_image back to anexo_url
ALTER TABLE oportunidades_mei
RENAME COLUMN cover_image TO anexo_url;

-- +goose StatementEnd
