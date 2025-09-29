-- +goose Up
-- +goose StatementBegin
-- Add enrolled_unit column to inscricoes table to store unit enrollment information
ALTER TABLE inscricoes ADD COLUMN enrolled_unit JSONB;

-- Add comment explaining the enrolled_unit structure
COMMENT ON COLUMN inscricoes.enrolled_unit IS 'Optional JSONB field storing unit enrollment details: {id, curso_id, address, neighborhood, vacancies, class_start_date, class_end_date, class_time, class_days, created_at, updated_at}';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Remove enrolled_unit column
ALTER TABLE inscricoes DROP COLUMN IF EXISTS enrolled_unit;
-- +goose StatementEnd