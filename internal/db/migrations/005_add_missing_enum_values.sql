-- +goose Up

-- Add missing enum values (ignore errors if they already exist)
DO $$
BEGIN
    -- Try to add each value, ignore if it already exists
    BEGIN
        ALTER TYPE status_curso_enum ADD VALUE 'draft';
    EXCEPTION 
        WHEN duplicate_object THEN NULL;
    END;
    
    BEGIN
        ALTER TYPE status_curso_enum ADD VALUE 'opened';
    EXCEPTION 
        WHEN duplicate_object THEN NULL;
    END;
    
    BEGIN
        ALTER TYPE status_curso_enum ADD VALUE 'closed';
    EXCEPTION 
        WHEN duplicate_object THEN NULL;
    END;
    
    BEGIN
        ALTER TYPE status_curso_enum ADD VALUE 'canceled';
    EXCEPTION 
        WHEN duplicate_object THEN NULL;
    END;
END $$;

-- +goose Down
-- No rollback needed - enum values cannot be removed easily