-- +goose Up
-- +goose StatementBegin

-- Requisitos de idioma + nível mínimo exigidos por uma vaga (critério de elegibilidade).
-- Entidade própria (não many2many simples) por carregar a coluna id_nivel_minimo.
CREATE TABLE emp_vagas_idiomas_requisitos (
    id_vaga UUID REFERENCES emp_vagas(id) ON DELETE CASCADE,
    id_idioma UUID REFERENCES emp_idiomas(id) ON DELETE CASCADE,
    id_nivel_minimo UUID NOT NULL REFERENCES emp_niveis_idioma(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id_vaga, id_idioma)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS emp_vagas_idiomas_requisitos;

-- +goose StatementEnd
