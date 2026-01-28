-- +goose Up
-- +goose StatementBegin

-- Onboarding (primeiro login)
CREATE TABLE emp_onboarding (
    cpf VARCHAR(14) PRIMARY KEY,
    is_empregabilidade_first_login BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Termos de Uso
CREATE TABLE emp_termos_uso (
    cpf VARCHAR(14) PRIMARY KEY,
    user_consent BOOLEAN DEFAULT FALSE,
    aceito_em TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS emp_termos_uso;
DROP TABLE IF EXISTS emp_onboarding;

-- +goose StatementEnd
