-- +goose Up
-- +goose StatementBegin

-- Add display_order column to course_schedules
ALTER TABLE course_schedules ADD COLUMN display_order INTEGER NOT NULL DEFAULT 0;

-- Add display_order column to remote_schedules
ALTER TABLE remote_schedules ADD COLUMN display_order INTEGER NOT NULL DEFAULT 0;

-- Backfill display_order for existing course_schedules
UPDATE course_schedules cs
SET display_order = sub.rn
FROM (
    SELECT id, ROW_NUMBER() OVER (PARTITION BY location_id ORDER BY created_at, id) AS rn
    FROM course_schedules
) sub
WHERE cs.id = sub.id;

-- Backfill display_order for existing remote_schedules
UPDATE remote_schedules rs
SET display_order = sub.rn
FROM (
    SELECT id, ROW_NUMBER() OVER (PARTITION BY remote_class_id ORDER BY created_at, id) AS rn
    FROM remote_schedules
) sub
WHERE rs.id = sub.id;

-- Create indexes for ordering queries
CREATE INDEX idx_course_schedules_location_display_order ON course_schedules(location_id, display_order);
CREATE INDEX idx_remote_schedules_remote_class_display_order ON remote_schedules(remote_class_id, display_order);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_course_schedules_location_display_order;
DROP INDEX IF EXISTS idx_remote_schedules_remote_class_display_order;

ALTER TABLE course_schedules DROP COLUMN IF EXISTS display_order;
ALTER TABLE remote_schedules DROP COLUMN IF EXISTS display_order;

-- +goose StatementEnd
