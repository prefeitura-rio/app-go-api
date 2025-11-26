-- +goose Up
-- +goose StatementBegin
ALTER TABLE location_classes ADD COLUMN neighborhood_zone VARCHAR(20000);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE location_classes DROP COLUMN IF EXISTS neighborhood_zone;
-- +goose StatementEnd

