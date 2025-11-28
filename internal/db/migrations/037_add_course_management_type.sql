-- +goose Up
-- +goose StatementBegin
ALTER TABLE cursos ADD COLUMN IF NOT EXISTS course_management_type VARCHAR(50) DEFAULT 'OWN_ORG';

UPDATE cursos
SET course_management_type = CASE
    WHEN is_external_partner = true AND external_partner_url IS NOT NULL AND external_partner_url != ''
        THEN 'EXTERNAL_MANAGED_BY_PARTNER'
    WHEN is_external_partner = true AND external_partner_name IS NOT NULL AND external_partner_name != ''
        THEN 'EXTERNAL_MANAGED_BY_ORG'
    ELSE 'OWN_ORG'
END
WHERE course_management_type IS NULL OR course_management_type = '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE cursos DROP COLUMN IF EXISTS course_management_type;
-- +goose StatementEnd

