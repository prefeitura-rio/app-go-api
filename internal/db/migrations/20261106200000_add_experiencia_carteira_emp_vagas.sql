-- +goose Up
-- Tempo de experiência comprovada em carteira exigido pela vaga (DEV-424,
-- história PREF-633): qualquer_tempo ou minimo_6_meses (prazo fixo). NULL =
-- sem exigência, que é o padrão; o bloqueio de candidatura é a chave
-- "experiencia_carteira" de criterios_bloqueio. Só acrescenta uma coluna
-- nullable (expand): o código anterior a esta migração segue funcionando e
-- nenhuma vaga existente muda.
ALTER TABLE emp_vagas
    ADD COLUMN IF NOT EXISTS experiencia_carteira VARCHAR(20) NULL;

ALTER TABLE emp_vagas
    DROP CONSTRAINT IF EXISTS chk_emp_vagas_experiencia_carteira;
ALTER TABLE emp_vagas
    ADD CONSTRAINT chk_emp_vagas_experiencia_carteira
    CHECK (experiencia_carteira IS NULL
           OR experiencia_carteira IN ('qualquer_tempo', 'minimo_6_meses'));

-- +goose Down
ALTER TABLE emp_vagas
    DROP CONSTRAINT IF EXISTS chk_emp_vagas_experiencia_carteira;
ALTER TABLE emp_vagas
    DROP COLUMN IF EXISTS experiencia_carteira;
