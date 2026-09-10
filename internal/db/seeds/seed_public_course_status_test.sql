-- Seed: status visibility scenarios for GET /api/public/courses/:courseId
-- One course per relevant status. IDs 910–916.
-- Used with scripts/test_public_course_status.sh

INSERT INTO cursos (id, titulo, orgao_id, status, modalidade, data_inicio, data_termino, data_limite_inscricoes, numero_vagas)
VALUES
    (910, '[SEED-TEST] Curso published',        'orgao-seed-a', 'published',        'presencial', NOW() + INTERVAL '30 days', NOW() + INTERVAL '60 days', NOW() + INTERVAL '25 days', 10),
    (911, '[SEED-TEST] Curso draft',            'orgao-seed-a', 'draft',            'presencial', NOW() + INTERVAL '30 days', NOW() + INTERVAL '60 days', NOW() + INTERVAL '25 days', 10),
    (912, '[SEED-TEST] Curso in_review',        'orgao-seed-a', 'in_review',        'presencial', NOW() + INTERVAL '30 days', NOW() + INTERVAL '60 days', NOW() + INTERVAL '25 days', 10),
    (913, '[SEED-TEST] Curso needs_changes',    'orgao-seed-a', 'needs_changes',    'presencial', NOW() + INTERVAL '30 days', NOW() + INTERVAL '60 days', NOW() + INTERVAL '25 days', 10),
    (914, '[SEED-TEST] Curso approved',         'orgao-seed-a', 'approved',         'presencial', NOW() + INTERVAL '30 days', NOW() + INTERVAL '60 days', NOW() + INTERVAL '25 days', 10),
    (915, '[SEED-TEST] Curso pending_deletion', 'orgao-seed-a', 'pending_deletion', 'presencial', NOW() + INTERVAL '30 days', NOW() + INTERVAL '60 days', NOW() + INTERVAL '25 days', 10),
    (916, '[SEED-TEST] Curso closed',           'orgao-seed-a', 'closed',           'presencial', NOW() - INTERVAL '60 days', NOW() - INTERVAL '30 days', NOW() - INTERVAL '35 days', 10)
ON CONFLICT (id) DO UPDATE
    SET status = EXCLUDED.status,
        titulo = EXCLUDED.titulo;
