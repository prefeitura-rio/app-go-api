-- +goose Up
-- +goose StatementBegin
-- Increase target_audience character limit from 200 to 600
ALTER TABLE cursos ALTER COLUMN target_audience TYPE VARCHAR(600);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Revert target_audience character limit back to 200
ALTER TABLE cursos ALTER COLUMN target_audience TYPE VARCHAR(200);
-- +goose StatementEnd