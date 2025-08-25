-- +goose Up
-- +goose StatementBegin

-- Update course status enum to include draft and match specification
ALTER TYPE status_curso_enum ADD VALUE 'DRAFT';
ALTER TYPE status_curso_enum ADD VALUE 'OPENED';
ALTER TYPE status_curso_enum ADD VALUE 'CLOSED';
ALTER TYPE status_curso_enum ADD VALUE 'CANCELED';

-- Create enrollment status enum
CREATE TYPE status_inscricao_enum AS ENUM ('pending', 'approved', 'rejected', 'cancelled');

-- Create custom fields table for enrollment forms
CREATE TABLE custom_fields (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    curso_id INTEGER NOT NULL REFERENCES cursos(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    field_type VARCHAR(50) NOT NULL DEFAULT 'text', -- text, textarea, select, checkbox, etc.
    required BOOLEAN NOT NULL DEFAULT false,
    options JSONB, -- For select fields, store options as JSON array
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create enrollments table
CREATE TABLE inscricoes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    curso_id INTEGER NOT NULL REFERENCES cursos(id) ON DELETE CASCADE,
    cpf VARCHAR(11) NOT NULL,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    status status_inscricao_enum NOT NULL DEFAULT 'pending',
    custom_fields_data JSONB, -- Store custom field responses as JSON
    admin_notes TEXT,
    reason TEXT, -- Reason for approval/rejection
    enrolled_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Ensure one enrollment per person per course
    UNIQUE(curso_id, cpf)
);

-- Create location classes table for presential/semi-presential courses
CREATE TABLE location_classes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    curso_id INTEGER NOT NULL REFERENCES cursos(id) ON DELETE CASCADE,
    address VARCHAR(500) NOT NULL,
    neighborhood VARCHAR(100) NOT NULL,
    vacancies INTEGER NOT NULL CHECK (vacancies > 0 AND vacancies <= 1000),
    class_start_date DATE NOT NULL,
    class_end_date DATE NOT NULL,
    class_time VARCHAR(50) NOT NULL, -- "HH:MM" or "HH:MM - HH:MM"
    class_days VARCHAR(200) NOT NULL, -- "Segunda, Quarta, Sexta"
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create remote classes table for online courses
CREATE TABLE remote_classes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    curso_id INTEGER NOT NULL REFERENCES cursos(id) ON DELETE CASCADE,
    vacancies INTEGER NOT NULL CHECK (vacancies > 0 AND vacancies <= 1000),
    class_start_date DATE NOT NULL,
    class_end_date DATE NOT NULL,
    class_time VARCHAR(50) NOT NULL, -- "HH:MM" or "HH:MM - HH:MM"
    class_days VARCHAR(200) NOT NULL, -- "Segunda, Quarta, Sexta"
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Only one remote class per course
    UNIQUE(curso_id)
);

-- Add new fields to courses table to match specification
ALTER TABLE cursos ADD COLUMN IF NOT EXISTS organization VARCHAR(255);
ALTER TABLE cursos ADD COLUMN IF NOT EXISTS theme VARCHAR(100);
ALTER TABLE cursos ADD COLUMN IF NOT EXISTS workload VARCHAR(50);
ALTER TABLE cursos ADD COLUMN IF NOT EXISTS target_audience VARCHAR(200);
ALTER TABLE cursos ADD COLUMN IF NOT EXISTS institutional_logo VARCHAR(500);
ALTER TABLE cursos ADD COLUMN IF NOT EXISTS cover_image VARCHAR(500);
ALTER TABLE cursos ADD COLUMN IF NOT EXISTS enrollment_start_date DATE;
ALTER TABLE cursos ADD COLUMN IF NOT EXISTS enrollment_end_date DATE;
ALTER TABLE cursos ADD COLUMN IF NOT EXISTS has_certificate BOOLEAN DEFAULT false;
ALTER TABLE cursos ADD COLUMN IF NOT EXISTS facilitator VARCHAR(255);
ALTER TABLE cursos ADD COLUMN IF NOT EXISTS objectives TEXT;
ALTER TABLE cursos ADD COLUMN IF NOT EXISTS expected_results TEXT;
ALTER TABLE cursos ADD COLUMN IF NOT EXISTS program_content TEXT;
ALTER TABLE cursos ADD COLUMN IF NOT EXISTS methodology TEXT;
ALTER TABLE cursos ADD COLUMN IF NOT EXISTS resources_used TEXT;
ALTER TABLE cursos ADD COLUMN IF NOT EXISTS material_used TEXT;
ALTER TABLE cursos ADD COLUMN IF NOT EXISTS teaching_material TEXT;

-- Create indexes for better performance
CREATE INDEX idx_inscricoes_curso ON inscricoes(curso_id);
CREATE INDEX idx_inscricoes_cpf ON inscricoes(cpf);
CREATE INDEX idx_inscricoes_status ON inscricoes(status);
CREATE INDEX idx_inscricoes_enrolled_at ON inscricoes(enrolled_at);

CREATE INDEX idx_custom_fields_curso ON custom_fields(curso_id);
CREATE INDEX idx_location_classes_curso ON location_classes(curso_id);
CREATE INDEX idx_remote_classes_curso ON remote_classes(curso_id);

-- Update existing indexes
CREATE INDEX IF NOT EXISTS idx_cursos_organization ON cursos(organization);
CREATE INDEX IF NOT EXISTS idx_cursos_theme ON cursos(theme);
CREATE INDEX IF NOT EXISTS idx_cursos_enrollment_dates ON cursos(enrollment_start_date, enrollment_end_date);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop new tables
DROP TABLE IF EXISTS remote_classes;
DROP TABLE IF EXISTS location_classes;
DROP TABLE IF EXISTS inscricoes;
DROP TABLE IF EXISTS custom_fields;

-- Drop new enum
DROP TYPE IF EXISTS status_inscricao_enum;

-- Remove added columns from cursos
ALTER TABLE cursos DROP COLUMN IF EXISTS teaching_material;
ALTER TABLE cursos DROP COLUMN IF EXISTS material_used;
ALTER TABLE cursos DROP COLUMN IF EXISTS resources_used;
ALTER TABLE cursos DROP COLUMN IF EXISTS methodology;
ALTER TABLE cursos DROP COLUMN IF EXISTS program_content;
ALTER TABLE cursos DROP COLUMN IF EXISTS expected_results;
ALTER TABLE cursos DROP COLUMN IF EXISTS objectives;
ALTER TABLE cursos DROP COLUMN IF EXISTS facilitator;
ALTER TABLE cursos DROP COLUMN IF EXISTS has_certificate;
ALTER TABLE cursos DROP COLUMN IF EXISTS enrollment_end_date;
ALTER TABLE cursos DROP COLUMN IF EXISTS enrollment_start_date;
ALTER TABLE cursos DROP COLUMN IF EXISTS cover_image;
ALTER TABLE cursos DROP COLUMN IF EXISTS institutional_logo;
ALTER TABLE cursos DROP COLUMN IF EXISTS target_audience;
ALTER TABLE cursos DROP COLUMN IF EXISTS workload;
ALTER TABLE cursos DROP COLUMN IF EXISTS theme;
ALTER TABLE cursos DROP COLUMN IF EXISTS organization;

-- Note: Cannot easily remove enum values in PostgreSQL, would require recreating the type
-- This is a limitation of PostgreSQL enum types

-- +goose StatementEnd