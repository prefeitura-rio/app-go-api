-- +goose Up
-- Remove NOT NULL constraint from email column in inscricoes table
ALTER TABLE inscricoes ALTER COLUMN email DROP NOT NULL;

-- +goose Down
-- Restore NOT NULL constraint (set empty string for any null values first)
UPDATE inscricoes SET email = '' WHERE email IS NULL;
ALTER TABLE inscricoes ALTER COLUMN email SET NOT NULL;

