-- +goose Up
-- Enable unaccent extension for accent-insensitive text search
CREATE EXTENSION IF NOT EXISTS unaccent;

-- +goose Down
-- Drop unaccent extension
DROP EXTENSION IF EXISTS unaccent;
