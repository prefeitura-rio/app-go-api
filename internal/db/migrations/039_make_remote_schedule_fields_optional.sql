-- +goose Up
-- +goose StatementBegin

-- Make schedule fields optional for Online (Remoto) courses
-- Vacancies remains required
ALTER TABLE remote_schedules
    ALTER COLUMN class_start_date DROP NOT NULL,
    ALTER COLUMN class_end_date DROP NOT NULL,
    ALTER COLUMN class_time DROP NOT NULL,
    ALTER COLUMN class_days DROP NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Restore NOT NULL constraints
-- Note: This will fail if there are NULL values in the database
ALTER TABLE remote_schedules
    ALTER COLUMN class_start_date SET NOT NULL,
    ALTER COLUMN class_end_date SET NOT NULL,
    ALTER COLUMN class_time SET NOT NULL,
    ALTER COLUMN class_days SET NOT NULL;

-- +goose StatementEnd
