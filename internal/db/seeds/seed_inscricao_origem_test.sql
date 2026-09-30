-- Seed: curso para testes de origem_inscricao
-- Cria um curso published sem período de inscrição e sem inscrições pré-existentes.
-- Usado junto com scripts/test_inscricao_origem.sh

-- Curso ID 950 — sem enrollment_start_date / enrollment_end_date para aceitar inscrições sempre
INSERT INTO cursos (id, titulo, orgao_id, status, modalidade, data_inicio, data_termino, data_limite_inscricoes, numero_vagas)
VALUES (
    950,
    '[SEED-TEST] Curso origem_inscricao',
    'orgao-seed-origem',
    'published',
    'presencial',
    NOW() + INTERVAL '30 days',
    NOW() + INTERVAL '60 days',
    NOW() + INTERVAL '25 days',
    100
)
ON CONFLICT (id) DO UPDATE
    SET titulo   = EXCLUDED.titulo,
        status   = EXCLUDED.status;

-- Remove inscrições anteriores de testes do curso 950 para garantir estado limpo
DELETE FROM inscricoes WHERE curso_id = 950;
