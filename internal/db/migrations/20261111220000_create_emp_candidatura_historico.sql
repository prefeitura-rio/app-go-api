-- +goose Up
-- Histórico de etapas e status de cada candidatura (DEV-532). A tabela é só de
-- inclusão: a API grava um evento na mesma transação de cada transição, e o
-- banco recusa UPDATE, DELETE e TRUNCATE. Guarda só ids (nada de CPF, nome ou
-- e-mail) e não tem chave estrangeira, para a exclusão de uma vaga (que apaga
-- candidaturas e etapas em cascata) não apagar o histórico. Candidaturas
-- antigas não são migradas: o primeiro evento novo de cada uma traz o estado
-- de origem em status_anterior / id_etapa_anterior.
CREATE TABLE IF NOT EXISTS emp_candidatura_historico (
    id BIGSERIAL PRIMARY KEY,
    id_candidatura UUID NOT NULL,
    id_vaga UUID NOT NULL,
    tipo_evento VARCHAR(64) NOT NULL,
    ator VARCHAR(32) NOT NULL,
    status_anterior VARCHAR(100) NULL,
    status_novo VARCHAR(100) NOT NULL,
    id_etapa_anterior UUID NULL,
    id_etapa_nova UUID NULL,
    motivo VARCHAR(100) NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_emp_candidatura_historico_tipo_evento CHECK (tipo_evento IN (
        'candidatura_criada', 'mudanca_etapa', 'volta_etapa', 'etapa_removida', 'candidatura_aprovada',
        'candidatura_reprovada', 'candidatura_cancelada', 'status_alterado'
    )),
    CONSTRAINT chk_emp_candidatura_historico_ator CHECK (ator IN ('candidato', 'operador', 'automacao')),
    -- Motivo é um código (vaga_congelada, ...), nunca texto livre.
    CONSTRAINT chk_emp_candidatura_historico_motivo CHECK (motivo IS NULL OR motivo ~ '^[a-z0-9_]{1,100}$')
);

CREATE INDEX IF NOT EXISTS idx_emp_candidatura_historico_candidatura
    ON emp_candidatura_historico (id_candidatura, created_at, id);
CREATE INDEX IF NOT EXISTS idx_emp_candidatura_historico_vaga
    ON emp_candidatura_historico (id_vaga, created_at);
CREATE INDEX IF NOT EXISTS idx_emp_candidatura_historico_tipo_evento
    ON emp_candidatura_historico (tipo_evento);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION emp_candidatura_historico_somente_inclusao() RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'emp_candidatura_historico é só de inclusão: % não é permitido', TG_OP;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

DROP TRIGGER IF EXISTS trg_emp_candidatura_historico_imutavel ON emp_candidatura_historico;
CREATE TRIGGER trg_emp_candidatura_historico_imutavel
    BEFORE UPDATE OR DELETE ON emp_candidatura_historico
    FOR EACH ROW EXECUTE FUNCTION emp_candidatura_historico_somente_inclusao();

DROP TRIGGER IF EXISTS trg_emp_candidatura_historico_sem_truncate ON emp_candidatura_historico;
CREATE TRIGGER trg_emp_candidatura_historico_sem_truncate
    BEFORE TRUNCATE ON emp_candidatura_historico
    FOR EACH STATEMENT EXECUTE FUNCTION emp_candidatura_historico_somente_inclusao();

-- Classificação opcional da etapa (código definido pela gestão) para comparar o
-- funil entre vagas. Etapas existentes ficam sem tipo.
ALTER TABLE emp_etapas
    ADD COLUMN IF NOT EXISTS tipo_etapa VARCHAR(64) NULL;

-- +goose Down
ALTER TABLE emp_etapas
    DROP COLUMN IF EXISTS tipo_etapa;
DROP TABLE IF EXISTS emp_candidatura_historico;
DROP FUNCTION IF EXISTS emp_candidatura_historico_somente_inclusao();
