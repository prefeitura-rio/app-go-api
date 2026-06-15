-- +goose Up
-- +goose StatementBegin
ALTER TABLE custom_fields ADD COLUMN format_type VARCHAR(50);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE custom_fields DROP COLUMN IF EXISTS format_type;
-- +goose StatementEnd
