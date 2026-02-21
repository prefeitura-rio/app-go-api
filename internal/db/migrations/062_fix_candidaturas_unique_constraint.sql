-- +goose Up
-- +goose StatementBegin

-- Remove a constraint UNIQUE(cpf, id_vaga) que não respeita soft delete.
-- Adiciona índice parcial que garante unicidade apenas entre candidaturas ativas (deleted_at IS NULL),
-- permitindo re-candidatura após exclusão.
ALTER TABLE emp_candidaturas DROP CONSTRAINT IF EXISTS emp_candidaturas_cpf_id_vaga_key;

CREATE UNIQUE INDEX uq_emp_candidaturas_cpf_vaga_active
    ON emp_candidaturas(cpf, id_vaga)
    WHERE deleted_at IS NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS uq_emp_candidaturas_cpf_vaga_active;

ALTER TABLE emp_candidaturas ADD CONSTRAINT emp_candidaturas_cpf_id_vaga_key UNIQUE(cpf, id_vaga);

-- +goose StatementEnd
