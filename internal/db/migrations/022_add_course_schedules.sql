-- +goose Up
-- +goose StatementBegin

-- 1. Create course_schedules table
CREATE TABLE course_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    location_id UUID NOT NULL REFERENCES location_classes(id) ON DELETE CASCADE,
    vacancies INTEGER NOT NULL CHECK (vacancies >= 1 AND vacancies <= 1000),
    class_start_date TIMESTAMPTZ NOT NULL,
    class_end_date TIMESTAMPTZ NOT NULL,
    class_time VARCHAR(50) NOT NULL,
    class_days VARCHAR(200) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create index for better query performance
CREATE INDEX idx_course_schedules_location_id ON course_schedules(location_id);

-- 2. Migrate existing data from location_classes to course_schedules
-- Each location_class becomes one schedule
INSERT INTO course_schedules (
    location_id,
    vacancies,
    class_start_date,
    class_end_date,
    class_time,
    class_days,
    created_at,
    updated_at
)
SELECT
    id,
    COALESCE(vacancies, 30) as vacancies,
    class_start_date,
    class_end_date,
    class_time,
    class_days,
    created_at,
    updated_at
FROM location_classes
WHERE class_start_date IS NOT NULL
  AND class_end_date IS NOT NULL
  AND class_time IS NOT NULL
  AND class_days IS NOT NULL;

-- 3. Add schedule_id column to inscricoes table (nullable for now)
ALTER TABLE inscricoes
ADD COLUMN schedule_id UUID;

-- Add foreign key constraint (but don't require it yet for backward compatibility during migration)
ALTER TABLE inscricoes
ADD CONSTRAINT fk_inscricoes_schedule
FOREIGN KEY (schedule_id) REFERENCES course_schedules(id) ON DELETE SET NULL;

-- 4. Remove legacy schedule-related columns from location_classes
ALTER TABLE location_classes
    DROP COLUMN IF EXISTS vacancies,
    DROP COLUMN IF EXISTS class_start_date,
    DROP COLUMN IF EXISTS class_end_date,
    DROP COLUMN IF EXISTS class_time,
    DROP COLUMN IF EXISTS class_days;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Restore columns to location_classes
ALTER TABLE location_classes
    ADD COLUMN vacancies INTEGER,
    ADD COLUMN class_start_date TIMESTAMPTZ,
    ADD COLUMN class_end_date TIMESTAMPTZ,
    ADD COLUMN class_time VARCHAR(50),
    ADD COLUMN class_days VARCHAR(200);

-- Migrate data back from schedules to locations (take first schedule per location)
UPDATE location_classes lc
SET
    vacancies = cs.vacancies,
    class_start_date = cs.class_start_date,
    class_end_date = cs.class_end_date,
    class_time = cs.class_time,
    class_days = cs.class_days
FROM (
    SELECT DISTINCT ON (location_id)
        location_id,
        vacancies,
        class_start_date,
        class_end_date,
        class_time,
        class_days
    FROM course_schedules
    ORDER BY location_id, created_at
) cs
WHERE lc.id = cs.location_id;

-- Remove schedule_id from inscricoes
ALTER TABLE inscricoes
    DROP CONSTRAINT IF EXISTS fk_inscricoes_schedule,
    DROP COLUMN IF EXISTS schedule_id;

-- Drop course_schedules table
DROP INDEX IF EXISTS idx_course_schedules_location_id;
DROP TABLE IF EXISTS course_schedules;

-- +goose StatementEnd
