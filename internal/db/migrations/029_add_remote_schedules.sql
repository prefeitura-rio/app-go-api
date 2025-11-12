-- +goose Up
-- +goose StatementBegin

-- 1. Create remote_schedules table
CREATE TABLE remote_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    remote_class_id UUID NOT NULL REFERENCES remote_classes(id) ON DELETE CASCADE,
    vacancies INTEGER NOT NULL CHECK (vacancies >= 1 AND vacancies <= 1000),
    class_start_date TIMESTAMPTZ NOT NULL,
    class_end_date TIMESTAMPTZ NOT NULL,
    class_time VARCHAR(20000) NOT NULL,
    class_days VARCHAR(20000) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create index for better query performance
CREATE INDEX idx_remote_schedules_remote_class_id ON remote_schedules(remote_class_id);

-- 2. Remove legacy schedule-related columns from remote_classes
ALTER TABLE remote_classes
    DROP COLUMN IF EXISTS vacancies,
    DROP COLUMN IF EXISTS class_start_date,
    DROP COLUMN IF EXISTS class_end_date,
    DROP COLUMN IF EXISTS class_time,
    DROP COLUMN IF EXISTS class_days;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Restore columns to remote_classes
ALTER TABLE remote_classes
    ADD COLUMN vacancies INTEGER,
    ADD COLUMN class_start_date TIMESTAMPTZ,
    ADD COLUMN class_end_date TIMESTAMPTZ,
    ADD COLUMN class_time VARCHAR(20000),
    ADD COLUMN class_days VARCHAR(20000);

-- Drop remote_schedules table
DROP INDEX IF EXISTS idx_remote_schedules_remote_class_id;
DROP TABLE IF EXISTS remote_schedules;

-- +goose StatementEnd
