-- +goose Up
-- +goose StatementBegin
CREATE TABLE emp_curriculo_perfil (
    cpf VARCHAR(14) PRIMARY KEY,
    resumo_profissional TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS emp_curriculo_perfil;
-- +goose StatementEnd
