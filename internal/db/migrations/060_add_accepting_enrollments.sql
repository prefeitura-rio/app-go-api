-- +goose Up
-- +goose StatementBegin

ALTER TABLE course_schedules ADD COLUMN accepting_enrollments BOOLEAN DEFAULT TRUE;
ALTER TABLE remote_schedules ADD COLUMN accepting_enrollments BOOLEAN DEFAULT TRUE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE course_schedules DROP COLUMN IF EXISTS accepting_enrollments;
ALTER TABLE remote_schedules DROP COLUMN IF EXISTS accepting_enrollments;

-- +goose StatementEnd
