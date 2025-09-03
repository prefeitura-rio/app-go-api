-- +goose Up

-- Corrigir campos de data para TIMESTAMP WITH TIME ZONE para incluir horário
-- Os campos enrollment_start_date e enrollment_end_date precisam incluir horário
-- Os campos class_start_date e class_end_date nas tabelas de classes também

BEGIN;

-- Alterar campos na tabela cursos
ALTER TABLE cursos 
    ALTER COLUMN enrollment_start_date TYPE TIMESTAMP WITH TIME ZONE USING enrollment_start_date::timestamp with time zone,
    ALTER COLUMN enrollment_end_date TYPE TIMESTAMP WITH TIME ZONE USING enrollment_end_date::timestamp with time zone;

-- Alterar campos na tabela location_classes
ALTER TABLE location_classes 
    ALTER COLUMN class_start_date TYPE TIMESTAMP WITH TIME ZONE USING class_start_date::timestamp with time zone,
    ALTER COLUMN class_end_date TYPE TIMESTAMP WITH TIME ZONE USING class_end_date::timestamp with time zone;

-- Alterar campos na tabela remote_classes
ALTER TABLE remote_classes 
    ALTER COLUMN class_start_date TYPE TIMESTAMP WITH TIME ZONE USING class_start_date::timestamp with time zone,
    ALTER COLUMN class_end_date TYPE TIMESTAMP WITH TIME ZONE USING class_end_date::timestamp with time zone;

COMMIT;

-- +goose Down

-- Reverter para DATE (com perda de informação de horário)
BEGIN;

ALTER TABLE cursos 
    ALTER COLUMN enrollment_start_date TYPE DATE USING enrollment_start_date::date,
    ALTER COLUMN enrollment_end_date TYPE DATE USING enrollment_end_date::date;

ALTER TABLE location_classes 
    ALTER COLUMN class_start_date TYPE DATE USING class_start_date::date,
    ALTER COLUMN class_end_date TYPE DATE USING class_end_date::date;

ALTER TABLE remote_classes 
    ALTER COLUMN class_start_date TYPE DATE USING class_start_date::date,
    ALTER COLUMN class_end_date TYPE DATE USING class_end_date::date;

COMMIT;