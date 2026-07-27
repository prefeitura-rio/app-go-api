-- +goose Up
-- +goose StatementBegin

-- Normaliza campos de data de cursos gravados em MEIA-NOITE UTC (00:00:00Z) para
-- o dia correto no fuso de Brasília (America/Sao_Paulo).
--
-- Contexto: a app-go-api passou a serializar timestamptz com offset -03:00 (antes: "Z").
-- O client (superapp) exibe essas datas truncando a string (split('T')[0]).
--   - enrollment_end_date = 2026-08-23T00:00:00Z
--       antes (UTC):    "2026-08-23" -> exibia 23/08  (batia com a intenção por acaso)
--       depois (-03:00): "2026-08-22T21:00:00-03:00" -> "2026-08-22" -> exibe 22/08 (-1 dia)
-- Causa-raiz: datas de dia inteiro foram gravadas em 00:00 UTC (= 21:00 BRT do dia anterior),
-- e não na meia-noite de Brasília. Esta migration reancora o INSTANTE ao dia pretendido:
--   - *_start_date  -> INÍCIO do dia em BRT  (00:00:00-03:00 == 03:00Z)
--   - *_end_date    -> FIM do dia em BRT     (23:59:59-03:00 == 02:59:59Z do dia seguinte)
--
-- Guarda: só toca valores cujo horário em UTC é EXATAMENTE 00:00:00. Timestamps reais
-- (created_at/updated_at, ou datas já com hora, como 22:00Z) têm fração/hora != 0 e são ignorados.
-- A lógica funcional (abrir/encerrar inscrição) compara instantes e não muda; isto corrige só a EXIBIÇÃO.
--
-- Idempotente: após a correção os valores ficam em 03:00Z (start) ou 02:59:59Z (end),
-- que não casam mais com a guarda `::time = '00:00:00'`; re-executar é no-op.
--
-- Ordem importa: normaliza as TURMAS primeiro e só então re-deriva o período
-- denormalizado do curso (min/max das turmas), espelhando ApplyDerivedEnrollmentPeriod
-- do runtime. Sem isso, um curso com turmas mistas (uma em 00:00Z, outra com hora real)
-- ficaria com o curso-nível defasado em relação à turma reancorada.

DO $$
DECLARE
    v_cursos_end   int;
    v_cs_end       int;
    v_cs_start     int;
    v_rs_end       int;
    v_rs_start     int;
    v_cursos_start int;
BEGIN
    SELECT count(*) INTO v_cursos_start FROM cursos
        WHERE enrollment_start_date IS NOT NULL AND (enrollment_start_date AT TIME ZONE 'UTC')::time = time '00:00:00';
    SELECT count(*) INTO v_cursos_end FROM cursos
        WHERE enrollment_end_date   IS NOT NULL AND (enrollment_end_date   AT TIME ZONE 'UTC')::time = time '00:00:00';
    SELECT count(*) INTO v_cs_start FROM course_schedules
        WHERE (class_start_date      IS NOT NULL AND (class_start_date      AT TIME ZONE 'UTC')::time = time '00:00:00')
           OR (enrollment_start_date IS NOT NULL AND (enrollment_start_date AT TIME ZONE 'UTC')::time = time '00:00:00');
    SELECT count(*) INTO v_cs_end FROM course_schedules
        WHERE (class_end_date        IS NOT NULL AND (class_end_date        AT TIME ZONE 'UTC')::time = time '00:00:00')
           OR (enrollment_end_date   IS NOT NULL AND (enrollment_end_date   AT TIME ZONE 'UTC')::time = time '00:00:00');
    SELECT count(*) INTO v_rs_start FROM remote_schedules
        WHERE (class_start_date      IS NOT NULL AND (class_start_date      AT TIME ZONE 'UTC')::time = time '00:00:00')
           OR (enrollment_start_date IS NOT NULL AND (enrollment_start_date AT TIME ZONE 'UTC')::time = time '00:00:00');
    SELECT count(*) INTO v_rs_end FROM remote_schedules
        WHERE (class_end_date        IS NOT NULL AND (class_end_date        AT TIME ZONE 'UTC')::time = time '00:00:00')
           OR (enrollment_end_date   IS NOT NULL AND (enrollment_end_date   AT TIME ZONE 'UTC')::time = time '00:00:00');

    RAISE NOTICE 'normalize_course_dates_timezone: cursos(start=%, end=%) course_schedules(start=%, end=%) remote_schedules(start=%, end=%)',
        v_cursos_start, v_cursos_end, v_cs_start, v_cs_end, v_rs_start, v_rs_end;
END $$;

-- ---------- END dates -> fim do dia em BRT (23:59:59 America/Sao_Paulo) ----------

UPDATE cursos
SET enrollment_end_date =
    ((enrollment_end_date AT TIME ZONE 'UTC')::date + time '23:59:59') AT TIME ZONE 'America/Sao_Paulo'
WHERE enrollment_end_date IS NOT NULL
  AND (enrollment_end_date AT TIME ZONE 'UTC')::time = time '00:00:00';

UPDATE course_schedules
SET class_end_date =
    ((class_end_date AT TIME ZONE 'UTC')::date + time '23:59:59') AT TIME ZONE 'America/Sao_Paulo'
WHERE class_end_date IS NOT NULL
  AND (class_end_date AT TIME ZONE 'UTC')::time = time '00:00:00';

UPDATE course_schedules
SET enrollment_end_date =
    ((enrollment_end_date AT TIME ZONE 'UTC')::date + time '23:59:59') AT TIME ZONE 'America/Sao_Paulo'
WHERE enrollment_end_date IS NOT NULL
  AND (enrollment_end_date AT TIME ZONE 'UTC')::time = time '00:00:00';

UPDATE remote_schedules
SET class_end_date =
    ((class_end_date AT TIME ZONE 'UTC')::date + time '23:59:59') AT TIME ZONE 'America/Sao_Paulo'
WHERE class_end_date IS NOT NULL
  AND (class_end_date AT TIME ZONE 'UTC')::time = time '00:00:00';

UPDATE remote_schedules
SET enrollment_end_date =
    ((enrollment_end_date AT TIME ZONE 'UTC')::date + time '23:59:59') AT TIME ZONE 'America/Sao_Paulo'
WHERE enrollment_end_date IS NOT NULL
  AND (enrollment_end_date AT TIME ZONE 'UTC')::time = time '00:00:00';

-- ---------- START dates -> início do dia em BRT (00:00:00 America/Sao_Paulo) ----------

UPDATE cursos
SET enrollment_start_date =
    ((enrollment_start_date AT TIME ZONE 'UTC')::date)::timestamp AT TIME ZONE 'America/Sao_Paulo'
WHERE enrollment_start_date IS NOT NULL
  AND (enrollment_start_date AT TIME ZONE 'UTC')::time = time '00:00:00';

UPDATE course_schedules
SET class_start_date =
    ((class_start_date AT TIME ZONE 'UTC')::date)::timestamp AT TIME ZONE 'America/Sao_Paulo'
WHERE class_start_date IS NOT NULL
  AND (class_start_date AT TIME ZONE 'UTC')::time = time '00:00:00';

UPDATE course_schedules
SET enrollment_start_date =
    ((enrollment_start_date AT TIME ZONE 'UTC')::date)::timestamp AT TIME ZONE 'America/Sao_Paulo'
WHERE enrollment_start_date IS NOT NULL
  AND (enrollment_start_date AT TIME ZONE 'UTC')::time = time '00:00:00';

UPDATE remote_schedules
SET class_start_date =
    ((class_start_date AT TIME ZONE 'UTC')::date)::timestamp AT TIME ZONE 'America/Sao_Paulo'
WHERE class_start_date IS NOT NULL
  AND (class_start_date AT TIME ZONE 'UTC')::time = time '00:00:00';

UPDATE remote_schedules
SET enrollment_start_date =
    ((enrollment_start_date AT TIME ZONE 'UTC')::date)::timestamp AT TIME ZONE 'America/Sao_Paulo'
WHERE enrollment_start_date IS NOT NULL
  AND (enrollment_start_date AT TIME ZONE 'UTC')::time = time '00:00:00';

-- ---------- Re-deriva o período de inscrição denormalizado do curso ----------
-- Espelha models.ApplyDerivedEnrollmentPeriod: curso.enrollment_{start,end} =
-- (menor abertura, maior fechamento) entre as turmas. Roda DEPOIS das turmas já
-- normalizadas para não deixar o curso-nível defasado no caso de turmas mistas.
-- Não toca LIVRE_FORMACAO_ONLINE (sem turmas; mantém as próprias datas — igual ao runtime).
-- `IS DISTINCT FROM` restringe a escrita apenas aos cursos que realmente divergem.
WITH turma_enroll AS (
    SELECT lc.curso_id AS curso_id,
           cs.enrollment_start_date AS es,
           cs.enrollment_end_date   AS ee
      FROM course_schedules cs
      JOIN location_classes lc ON lc.id = cs.location_id
     WHERE cs.enrollment_start_date IS NOT NULL
       AND cs.enrollment_end_date   IS NOT NULL
    UNION ALL
    SELECT rc.curso_id,
           rs.enrollment_start_date,
           rs.enrollment_end_date
      FROM remote_schedules rs
      JOIN remote_classes rc ON rc.id = rs.remote_class_id
     WHERE rs.enrollment_start_date IS NOT NULL
       AND rs.enrollment_end_date   IS NOT NULL
),
agg AS (
    SELECT curso_id, MIN(es) AS min_start, MAX(ee) AS max_end
      FROM turma_enroll
     GROUP BY curso_id
)
UPDATE cursos c
SET enrollment_start_date = agg.min_start,
    enrollment_end_date   = agg.max_end
FROM agg
WHERE c.id = agg.curso_id
  AND c.modalidade <> 'LIVRE_FORMACAO_ONLINE'
  AND (c.enrollment_start_date IS DISTINCT FROM agg.min_start
       OR c.enrollment_end_date IS DISTINCT FROM agg.max_end);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Migration de normalização de dados: não é reversível de forma segura, pois o valor
-- original (00:00:00Z) é indistinguível de datas legitimamente gravadas como início/fim
-- de dia em BRT após esta correção. Faça snapshot/backup das tabelas cursos,
-- course_schedules e remote_schedules antes de aplicar, caso precise reverter.
SELECT 1;

-- +goose StatementEnd
