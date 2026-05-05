-- +goose Up
-- Courses with legacy "opened" status are semantically equivalent to "published".
-- DeriveStatus already normalizes "opened" -> "published" at read time, but only
-- after this migration runs will courses without recent writes show derived statuses
-- (scheduled, accepting_enrollments, in_progress, finished) correctly.
UPDATE cursos SET status = 'published' WHERE status = 'opened';

-- +goose Down
-- Intentionally empty: reverting would incorrectly overwrite legitimately published
-- courses back to the legacy "opened" value.
