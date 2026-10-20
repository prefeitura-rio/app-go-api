-- +goose Up
-- Origem da candidatura: o próprio cidadão pelo app (portal_cidadao), o
-- cadastro manual na vaga (secretaria) ou a associação feita a partir do banco
-- de currículos (banco_curriculos). Candidaturas antigas ficam sem origem.
ALTER TABLE emp_candidaturas
    ADD COLUMN IF NOT EXISTS origem VARCHAR(64) NULL,
    ADD COLUMN IF NOT EXISTS associado_por_cpf VARCHAR(14) NULL;

-- +goose Down
ALTER TABLE emp_candidaturas
    DROP COLUMN IF EXISTS associado_por_cpf,
    DROP COLUMN IF EXISTS origem;
