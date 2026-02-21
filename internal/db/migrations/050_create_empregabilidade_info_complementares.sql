-- +goose Up
-- +goose StatementBegin

-- Informações complementares (perguntas customizadas)
CREATE TABLE emp_informacoes_complementares (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    id_vaga UUID NOT NULL REFERENCES emp_vagas(id) ON DELETE CASCADE,
    titulo VARCHAR(500) NOT NULL,
    obrigatorio BOOLEAN DEFAULT FALSE,
    tipo_campo VARCHAR(50) NOT NULL, -- resposta_curta, resposta_numerica, selecao_unica, selecao_multipla
    valor_minimo INTEGER,
    valor_maximo INTEGER,
    opcoes JSONB, -- array de strings para seleção
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_emp_info_comp_vaga ON emp_informacoes_complementares(id_vaga);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_emp_info_comp_vaga;
DROP TABLE IF EXISTS emp_informacoes_complementares;

-- +goose StatementEnd
