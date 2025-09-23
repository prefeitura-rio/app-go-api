-- +goose Up
-- +goose StatementBegin

-- Add external partner fields to cursos table
ALTER TABLE cursos ADD COLUMN IF NOT EXISTS is_external_partner BOOLEAN;
ALTER TABLE cursos ADD COLUMN IF NOT EXISTS external_partner_name VARCHAR(255);
ALTER TABLE cursos ADD COLUMN IF NOT EXISTS external_partner_url VARCHAR(500);
ALTER TABLE cursos ADD COLUMN IF NOT EXISTS external_partner_logo_url VARCHAR(500);
ALTER TABLE cursos ADD COLUMN IF NOT EXISTS external_partner_contact VARCHAR(255);

-- Create indexes for better performance on external partner queries
CREATE INDEX IF NOT EXISTS idx_cursos_is_external_partner ON cursos(is_external_partner);
CREATE INDEX IF NOT EXISTS idx_cursos_external_partner_name ON cursos(external_partner_name);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove external partner fields from cursos table
ALTER TABLE cursos DROP COLUMN IF EXISTS external_partner_contact;
ALTER TABLE cursos DROP COLUMN IF EXISTS external_partner_logo_url;
ALTER TABLE cursos DROP COLUMN IF EXISTS external_partner_url;
ALTER TABLE cursos DROP COLUMN IF EXISTS external_partner_name;
ALTER TABLE cursos DROP COLUMN IF EXISTS is_external_partner;

-- +goose StatementEnd