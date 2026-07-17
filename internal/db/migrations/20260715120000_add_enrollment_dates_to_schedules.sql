-- +goose Up
-- +goose StatementBegin

-- Enrollment period moves from the course level down to each turma (schedule).
-- Nullable so existing rows can be backfilled from the parent course below;
-- new writes always provide both dates (enforced in the service layer).
ALTER TABLE course_schedules ADD COLUMN IF NOT EXISTS enrollment_start_date TIMESTAMPTZ;
ALTER TABLE course_schedules ADD COLUMN IF NOT EXISTS enrollment_end_date TIMESTAMPTZ;
ALTER TABLE remote_schedules ADD COLUMN IF NOT EXISTS enrollment_start_date TIMESTAMPTZ;
ALTER TABLE remote_schedules ADD COLUMN IF NOT EXISTS enrollment_end_date TIMESTAMPTZ;

-- Backfill each turma from its parent course's enrollment window
-- (course_schedules -> location_classes -> cursos).
UPDATE course_schedules cs
SET enrollment_start_date = c.enrollment_start_date,
    enrollment_end_date   = c.enrollment_end_date
FROM location_classes lc
JOIN cursos c ON c.id = lc.curso_id
WHERE cs.location_id = lc.id
  AND cs.enrollment_start_date IS NULL;

-- (remote_schedules -> remote_classes -> cursos).
UPDATE remote_schedules rs
SET enrollment_start_date = c.enrollment_start_date,
    enrollment_end_date   = c.enrollment_end_date
FROM remote_classes rc
JOIN cursos c ON c.id = rc.curso_id
WHERE rs.remote_class_id = rc.id
  AND rs.enrollment_start_date IS NULL;

-- Index the turma enrollment end date used by "open turmas" / status queries.
CREATE INDEX IF NOT EXISTS idx_course_schedules_enrollment_end ON course_schedules(enrollment_end_date);
CREATE INDEX IF NOT EXISTS idx_remote_schedules_enrollment_end ON remote_schedules(enrollment_end_date);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_course_schedules_enrollment_end;
DROP INDEX IF EXISTS idx_remote_schedules_enrollment_end;

ALTER TABLE course_schedules DROP COLUMN IF EXISTS enrollment_start_date;
ALTER TABLE course_schedules DROP COLUMN IF EXISTS enrollment_end_date;
ALTER TABLE remote_schedules DROP COLUMN IF EXISTS enrollment_start_date;
ALTER TABLE remote_schedules DROP COLUMN IF EXISTS enrollment_end_date;

-- +goose StatementEnd
