-- +goose Up
-- Critérios de elegibilidade com bloqueio de candidatura ligado (DEV-392):
-- lista JSON das chaves ligadas ("escolaridade", "idioma"; PcD, idade e
-- experiência entram como chaves novas). NULL = nenhum ligado, que é o padrão
-- de vaga nova; o desligado nunca é gravado.
ALTER TABLE emp_vagas
    ADD COLUMN IF NOT EXISTS criterios_bloqueio JSONB NULL;

-- Vagas existentes continuam bloqueando pelo que já bloqueiam hoje: liga o
-- critério de escolaridade onde há escolaridade mínima e o de idioma onde há
-- requisito de idioma. Só preenche quem ainda não tem estado gravado, então
-- rodar de novo não muda nada.
UPDATE emp_vagas v
SET criterios_bloqueio =
        CASE WHEN v.id_escolaridade_minima IS NOT NULL
             THEN '["escolaridade"]'::jsonb ELSE '[]'::jsonb END
        ||
        CASE WHEN EXISTS (SELECT 1 FROM emp_vagas_idiomas_requisitos r WHERE r.id_vaga = v.id)
             THEN '["idioma"]'::jsonb ELSE '[]'::jsonb END
WHERE v.criterios_bloqueio IS NULL
  AND (v.id_escolaridade_minima IS NOT NULL
       OR EXISTS (SELECT 1 FROM emp_vagas_idiomas_requisitos r WHERE r.id_vaga = v.id));

-- +goose Down
ALTER TABLE emp_vagas
    DROP COLUMN IF EXISTS criterios_bloqueio;
