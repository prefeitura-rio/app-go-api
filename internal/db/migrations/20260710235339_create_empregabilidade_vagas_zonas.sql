-- +goose Up
-- +goose StatementBegin

CREATE TABLE emp_vagas_zonas (
    id_vaga UUID REFERENCES emp_vagas(id) ON DELETE CASCADE,
    id_zona UUID REFERENCES emp_zonas(id) ON DELETE CASCADE,
    PRIMARY KEY (id_vaga, id_zona)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS emp_vagas_zonas;

-- +goose StatementEnd
