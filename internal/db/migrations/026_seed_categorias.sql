-- +goose Up
-- +goose StatementBegin

-- Add UNIQUE constraint to categorias.nome if it doesn't exist
-- This prevents duplicate categories and enables ON CONFLICT
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'categorias_nome_key'
    ) THEN
        ALTER TABLE categorias ADD CONSTRAINT categorias_nome_key UNIQUE (nome);
    END IF;
END $$;

-- Seed categorias
-- Using ON CONFLICT DO NOTHING for idempotency
INSERT INTO categorias (nome) VALUES
('Tecnologia'),
('Marketing'),
('Finanças'),
('Gestão'),
('Games'),
('Nutrição'),
('Educação'),
('Gastronomia'),
('Saúde'),
('Veterinária'),
('Carreira'),
('Estética'),
('Sustentabilidade'),
('Artes'),
('Cultura'),
('Astronomia')
ON CONFLICT (nome) DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove seeded categories
DELETE FROM categorias
WHERE nome IN (
    'Tecnologia',
    'Marketing',
    'Finanças',
    'Gestão',
    'Games',
    'Nutrição',
    'Educação',
    'Gastronomia',
    'Saúde',
    'Veterinária',
    'Carreira',
    'Estética',
    'Sustentabilidade',
    'Artes',
    'Cultura',
    'Astronomia'
);

-- Remove UNIQUE constraint (optional, uncomment if you want to remove it on rollback)
-- ALTER TABLE categorias DROP CONSTRAINT IF EXISTS categorias_nome_key;

-- +goose StatementEnd
