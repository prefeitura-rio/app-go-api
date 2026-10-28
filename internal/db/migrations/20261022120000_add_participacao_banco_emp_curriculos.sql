-- +goose Up
-- Participação no banco de currículos: todo currículo fica visível para os
-- recrutadores, e o cidadão pode se ocultar no app a qualquer momento. Quem já
-- está na base segue visível.
--
-- origem_cadastro registra como o currículo chegou à base, só para controle (o
-- recrutador não vê): pela candidatura a uma vaga (vaga), quando a pessoa monta o
-- currículo na hora de se candidatar, ou pelo preenchimento direto do currículo
-- no app (banco_curriculos). É gravada na primeira candidatura e fica nula
-- enquanto a pessoa não se candidata — nesse caso ela veio pelo app.
ALTER TABLE emp_curriculos
    ADD COLUMN IF NOT EXISTS disponivel_banco BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS disponivel_banco_alterado_em TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS origem_cadastro VARCHAR(32) NULL;

-- Mesma regra da aplicação (RegistrarOrigemCadastro): currículo criado até 24h
-- antes da primeira candidatura própria veio pela vaga; antes disso, pelo app.
-- A candidatura criada pelo backoffice pode ter o CPF com máscara, e as
-- associações feitas pelo recrutador não contam. As colunas de candidatura não
-- têm fuso e guardam o horário de Brasília.
UPDATE emp_curriculos c
SET origem_cadastro = CASE
    WHEN c.created_at >= pc.primeira - INTERVAL '24 hours' THEN 'vaga'
    ELSE 'banco_curriculos'
END
FROM (
    SELECT regexp_replace(cpf, '\D', '', 'g') AS cpf,
           MIN(created_at) AT TIME ZONE 'America/Sao_Paulo' AS primeira
    FROM emp_candidaturas
    WHERE origem IS DISTINCT FROM 'banco_curriculos'
    GROUP BY 1
) pc
WHERE pc.cpf = c.cpf;

CREATE INDEX IF NOT EXISTS idx_emp_curriculos_disponiveis
    ON emp_curriculos (created_at DESC)
    WHERE disponivel_banco;

-- +goose Down
DROP INDEX IF EXISTS idx_emp_curriculos_disponiveis;
ALTER TABLE emp_curriculos
    DROP COLUMN IF EXISTS origem_cadastro,
    DROP COLUMN IF EXISTS disponivel_banco_alterado_em,
    DROP COLUMN IF EXISTS disponivel_banco;
