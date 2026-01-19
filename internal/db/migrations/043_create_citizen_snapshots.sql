-- +goose Up
-- +goose StatementBegin

-- Create citizen_snapshots table to cache citizen data from RMI
CREATE TABLE citizen_snapshots (
    cpf VARCHAR(11) PRIMARY KEY,
    nome VARCHAR(500),
    email VARCHAR(500),
    celular VARCHAR(50),
    data_nascimento DATE,
    endereco JSONB,
    raca VARCHAR(100),
    genero VARCHAR(100),
    renda_familiar VARCHAR(100),
    escolaridade VARCHAR(100),
    deficiencia VARCHAR(500),
    last_synced_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create index for faster lookups by last_synced_at (for sync worker)
CREATE INDEX idx_citizen_snapshots_last_synced ON citizen_snapshots(last_synced_at);

-- Create index for email lookups
CREATE INDEX idx_citizen_snapshots_email ON citizen_snapshots(email) WHERE email IS NOT NULL AND email != '';

COMMENT ON TABLE citizen_snapshots IS 'Cache of citizen personal data from RMI API, synced periodically';
COMMENT ON COLUMN citizen_snapshots.cpf IS 'Primary key - CPF without formatting';
COMMENT ON COLUMN citizen_snapshots.endereco IS 'JSONB with logradouro, numero, complemento, bairro, cidade, uf, cep';
COMMENT ON COLUMN citizen_snapshots.last_synced_at IS 'Last time this record was synced with RMI';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_citizen_snapshots_email;
DROP INDEX IF EXISTS idx_citizen_snapshots_last_synced;
DROP TABLE IF EXISTS citizen_snapshots;

-- +goose StatementEnd
