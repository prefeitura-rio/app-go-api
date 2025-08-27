-- +goose Up

-- Converter ENUMs para VARCHAR para melhor manutenibilidade
-- Isso resolve problemas de migração e permite maior flexibilidade

BEGIN;

-- Alterar colunas de ENUM para VARCHAR
ALTER TABLE cursos 
    ALTER COLUMN modalidade TYPE VARCHAR(50),
    ALTER COLUMN status TYPE VARCHAR(50),
    ALTER COLUMN turno TYPE VARCHAR(50),
    ALTER COLUMN formato_aula TYPE VARCHAR(50);

-- Alterar outras tabelas que possam usar ENUMs
ALTER TABLE empregos 
    ALTER COLUMN turno TYPE VARCHAR(50);

-- Normalizar valores existentes para o novo formato
UPDATE cursos SET 
    status = CASE 
        WHEN status = 'CRIADO' THEN 'draft'
        WHEN status = 'ABERTO' THEN 'opened'
        WHEN status = 'ENCERRADO' THEN 'closed'
        ELSE status
    END,
    modalidade = CASE 
        WHEN modalidade = 'PRESENCIAL' THEN 'Presencial'
        WHEN modalidade = 'ONLINE' THEN 'Remoto'
        WHEN modalidade = 'HIBRIDO' THEN 'Semipresencial'
        ELSE modalidade
    END;

COMMIT;

-- +goose Down

-- Reverter para ENUMs (complexo, melhor evitar)
-- Esta migração é irreversível por simplicidade
-- Em caso de necessidade, criar nova migração específica