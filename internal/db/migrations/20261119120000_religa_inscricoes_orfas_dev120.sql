-- +goose Up
-- DEV-120: antes da DEV-119, o PUT do curso apagava a turma reenviada sem `id`
-- e a recriava com outro id, e as inscrições ficavam com schedule_id de uma
-- turma que não existe mais. Por decisão registrada na issue, só as órfãs com
-- vínculo inequívoco voltam a apontar para a turma atual: as dos cursos 418 e
-- 420 de staging. As demais ficam como estão.
--
-- O escopo é travado pelo par (curso, turma remota atual) de staging; esses
-- UUIDs não existem em produção, onde a migration não muda nada. Uma inscrição
-- só é religada quando o seu schedule_id não existe no curso e o horário dele
-- no enrolled_unit gravado na inscrição (vagas, horário, dias e datas de início
-- e fim) bate com exatamente uma turma atual da mesma turma remota. Sem par, ou
-- com mais de um, ela fica como está. Só o schedule_id muda: o enrolled_unit é
-- o retrato do momento da inscrição e não é reescrito. Não usa a API de troca
-- de turma, que envia e-mail ao cidadão.
--
-- As datas do enrolled_unit vêm como YYYY-MM-DD (import em lote) ou RFC3339
-- (serviço), no fuso do processo (UTC ou America/Sao_Paulo): compara-se o dia
-- nos dois fusos, sem converter o texto (um valor fora do formato não casa e
-- não quebra a migration).
-- +goose StatementBegin
WITH escopo (curso_id, remote_class_id) AS (
    VALUES
        (418, '0f335b86-ab22-4c4a-b9cf-81bbc5be7713'::uuid),
        (420, 'df1f0580-2112-414e-8245-98a1563168d2'::uuid)
),
orfas AS (
    SELECT i.id, i.schedule_id, rc.id AS remote_class_id, snap.horario
    FROM inscricoes i
    JOIN remote_classes rc ON rc.curso_id = i.curso_id
    JOIN escopo e ON e.curso_id = rc.curso_id AND e.remote_class_id = rc.id
    CROSS JOIN LATERAL (
        SELECT s AS horario
        FROM jsonb_array_elements(
            CASE WHEN jsonb_typeof(i.enrolled_unit -> 'schedules') = 'array'
                 THEN i.enrolled_unit -> 'schedules'
                 ELSE '[]'::jsonb END
        ) AS s
        WHERE s ->> 'id' = i.schedule_id::text
        LIMIT 1
    ) snap
    WHERE i.schedule_id IS NOT NULL
      AND NOT EXISTS (
          SELECT 1 FROM remote_schedules rs
          WHERE rs.id = i.schedule_id AND rs.remote_class_id = rc.id)
      AND NOT EXISTS (
          SELECT 1 FROM course_schedules cs
          JOIN location_classes lc ON lc.id = cs.location_id
          WHERE cs.id = i.schedule_id AND lc.curso_id = i.curso_id)
),
pares AS (
    SELECT o.id AS inscricao_id, rs.id AS schedule_id
    FROM orfas o
    JOIN remote_schedules rs ON rs.remote_class_id = o.remote_class_id
    WHERE o.horario ->> 'vacancies' = rs.vacancies::text
      AND COALESCE(o.horario ->> 'class_time', '') = COALESCE(rs.class_time, '')
      AND COALESCE(o.horario ->> 'class_days', '') = COALESCE(rs.class_days, '')
      AND LEFT(COALESCE(o.horario ->> 'class_start_date', ''), 10) IN (
          COALESCE(TO_CHAR(rs.class_start_date AT TIME ZONE 'UTC', 'YYYY-MM-DD'), ''),
          COALESCE(TO_CHAR(rs.class_start_date AT TIME ZONE 'America/Sao_Paulo', 'YYYY-MM-DD'), ''))
      AND LEFT(COALESCE(o.horario ->> 'class_end_date', ''), 10) IN (
          COALESCE(TO_CHAR(rs.class_end_date AT TIME ZONE 'UTC', 'YYYY-MM-DD'), ''),
          COALESCE(TO_CHAR(rs.class_end_date AT TIME ZONE 'America/Sao_Paulo', 'YYYY-MM-DD'), ''))
),
inequivocos AS (
    SELECT inscricao_id, MIN(schedule_id::text)::uuid AS schedule_id
    FROM pares
    GROUP BY inscricao_id
    HAVING COUNT(*) = 1
)
UPDATE inscricoes i
SET schedule_id = q.schedule_id
FROM inequivocos q
WHERE i.id = q.inscricao_id;
-- +goose StatementEnd

-- +goose Down
-- Correção de dados: não há o que desfazer.
