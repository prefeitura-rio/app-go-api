-- +goose Up
-- +goose StatementBegin

-- Registra tentativas de candidatura bloqueadas por não atenderem aos critérios
-- de elegibilidade configurados na vaga. Não há FK para emp_vagas nem para
-- candidaturas: o candidato bloqueado nunca chega a se candidatar de fato,
-- então nenhuma linha em emp_candidaturas é criada para esse fluxo.
CREATE TABLE emp_candidatura_bloqueios (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cpf VARCHAR(14) NOT NULL,
    id_vaga UUID NOT NULL,
    criterios_nao_atendidos JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_emp_candidatura_bloqueios_id_vaga ON emp_candidatura_bloqueios(id_vaga);
CREATE INDEX idx_emp_candidatura_bloqueios_cpf ON emp_candidatura_bloqueios(cpf);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_emp_candidatura_bloqueios_cpf;
DROP INDEX IF EXISTS idx_emp_candidatura_bloqueios_id_vaga;
DROP TABLE IF EXISTS emp_candidatura_bloqueios;

-- +goose StatementEnd
