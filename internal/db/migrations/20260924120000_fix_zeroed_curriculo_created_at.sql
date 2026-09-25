-- +goose Up
-- Os updates de seção do currículo usavam o Save do GORM, que regrava todas as
-- colunas: quem salvava a seção de novo ficava com created_at = 0001-01-01 (o
-- zero do Go). O backfill de emp_curriculos pega o menor created_at das seções e
-- herdou essa data, que aparecia como "31/12/1" no banco de currículos.
--
-- 1. Nas seções, a data real se perdeu; updated_at é a melhor aproximação.
UPDATE emp_curriculo_formacoes SET created_at = COALESCE(updated_at, CURRENT_TIMESTAMP) WHERE created_at < '1900-01-01';
UPDATE emp_curriculo_idiomas SET created_at = COALESCE(updated_at, CURRENT_TIMESTAMP) WHERE created_at < '1900-01-01';
UPDATE emp_curriculo_cursos_complementares SET created_at = COALESCE(updated_at, CURRENT_TIMESTAMP) WHERE created_at < '1900-01-01';
UPDATE emp_curriculo_experiencias SET created_at = COALESCE(updated_at, CURRENT_TIMESTAMP) WHERE created_at < '1900-01-01';
UPDATE emp_curriculo_conquistas SET created_at = COALESCE(updated_at, CURRENT_TIMESTAMP) WHERE created_at < '1900-01-01';
UPDATE emp_curriculo_situacao_interesses SET created_at = COALESCE(updated_at, CURRENT_TIMESTAMP) WHERE created_at < '1900-01-01';
UPDATE emp_curriculo_perfil SET created_at = COALESCE(updated_at, CURRENT_TIMESTAMP) WHERE created_at < '1900-01-01';

-- 2. Refaz a data de inclusão só dos currículos zerados, com a mesma regra do
-- backfill original (20260910120000_create_emp_curriculos), agora sobre as
-- seções corrigidas.
-- +goose StatementBegin
WITH secoes AS (
    SELECT cpf, created_at FROM emp_curriculo_formacoes
    UNION ALL SELECT cpf, created_at FROM emp_curriculo_idiomas
    UNION ALL SELECT cpf, created_at FROM emp_curriculo_cursos_complementares
    UNION ALL SELECT cpf, created_at FROM emp_curriculo_experiencias
    UNION ALL SELECT cpf, created_at FROM emp_curriculo_conquistas
    UNION ALL SELECT cpf, created_at FROM emp_curriculo_situacao_interesses
    UNION ALL SELECT cpf, created_at FROM emp_curriculo_perfil
),
curriculos AS (
    SELECT cpf, MIN(created_at) AS created_at FROM secoes
    WHERE created_at >= '1900-01-01'
    GROUP BY cpf
),
primeira_candidatura AS (
    SELECT cpf, MIN(created_at) AS created_at FROM emp_candidaturas
    WHERE created_at >= '1900-01-01'
    GROUP BY cpf
)
UPDATE emp_curriculos ec
SET created_at = COALESCE(LEAST(c.created_at, pc.created_at) AT TIME ZONE 'America/Sao_Paulo', CURRENT_TIMESTAMP)
FROM emp_curriculos zerado
LEFT JOIN curriculos c ON c.cpf = zerado.cpf
LEFT JOIN primeira_candidatura pc ON pc.cpf = zerado.cpf
WHERE ec.cpf = zerado.cpf
  AND ec.created_at < '1900-01-01';
-- +goose StatementEnd

-- +goose Down
-- Correção de dados: não há o que desfazer.
