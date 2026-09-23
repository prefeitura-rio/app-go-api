-- +goose Up
-- Registro raiz do currículo: guarda a data de inclusão no banco de currículos.
-- As tabelas de seção não servem para isso porque são apagadas e recriadas a
-- cada salvamento do formulário, o que reseta o created_at delas.
--
-- A coluna tem fuso (TIMESTAMPTZ): a API grava no horário de Brasília, e uma
-- coluna sem fuso seria lida de volta como UTC, deslocando a data em 3 horas.
-- +goose StatementBegin
CREATE TABLE emp_curriculos (
    cpf VARCHAR(14) PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

CREATE INDEX idx_emp_curriculos_created_at ON emp_curriculos (created_at DESC);

-- Backfill dos currículos existentes. Como as seções têm o created_at resetado a
-- cada salvamento, a primeira candidatura (cujo created_at nunca muda) aproxima a
-- data real de quem editou o currículo depois de se candidatar. As colunas das
-- seções e das candidaturas não têm fuso e guardam o horário de Brasília.
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
    SELECT cpf, MIN(created_at) AS created_at FROM secoes GROUP BY cpf
),
primeira_candidatura AS (
    SELECT cpf, MIN(created_at) AS created_at FROM emp_candidaturas GROUP BY cpf
)
INSERT INTO emp_curriculos (cpf, created_at)
SELECT c.cpf, COALESCE(LEAST(c.created_at, pc.created_at) AT TIME ZONE 'America/Sao_Paulo', CURRENT_TIMESTAMP)
FROM curriculos c
LEFT JOIN primeira_candidatura pc ON pc.cpf = c.cpf;
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS emp_curriculos;
