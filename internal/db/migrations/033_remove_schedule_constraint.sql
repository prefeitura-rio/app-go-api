-- +goose Up
-- +goose StatementBegin

-- Remove the foreign key constraint that only references course_schedules
-- This constraint prevents remote course schedules from being used in enrollments
-- Validation is now handled at the application layer in validateScheduleID function
ALTER TABLE inscricoes
DROP CONSTRAINT IF EXISTS fk_inscricoes_schedule;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Restore the constraint (but this will fail if there are remote schedules in use)
-- This is intentional - the down migration should only be used if no remote schedules exist
ALTER TABLE inscricoes
ADD CONSTRAINT fk_inscricoes_schedule
FOREIGN KEY (schedule_id) REFERENCES course_schedules(id) ON DELETE SET NULL;

-- +goose StatementEnd
