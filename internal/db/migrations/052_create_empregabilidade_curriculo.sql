-- +goose Up
-- +goose StatementBegin

-- Formação Acadêmica
CREATE TABLE emp_curriculo_formacoes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cpf VARCHAR(14) NOT NULL,
    id_escolaridade UUID REFERENCES emp_escolaridades(id),
    nome_instituicao VARCHAR(500),
    nome_curso VARCHAR(500),
    status VARCHAR(50), -- Completo, Em andamento, Incompleto
    ano_conclusao VARCHAR(4),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Idiomas do usuário
CREATE TABLE emp_curriculo_idiomas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cpf VARCHAR(14) NOT NULL,
    id_idioma UUID NOT NULL REFERENCES emp_idiomas(id),
    id_nivel UUID NOT NULL REFERENCES emp_niveis_idioma(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Cursos complementares
CREATE TABLE emp_curriculo_cursos_complementares (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cpf VARCHAR(14) NOT NULL,
    nome_curso VARCHAR(500) NOT NULL,
    nome_instituicao VARCHAR(500),
    ano_conclusao VARCHAR(4),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Experiências profissionais
CREATE TABLE emp_curriculo_experiencias (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cpf VARCHAR(14) NOT NULL,
    cargo VARCHAR(500) NOT NULL,
    empresa VARCHAR(500) NOT NULL,
    eh_trabalho_atual BOOLEAN DEFAULT FALSE,
    descricao_atividades TEXT,
    tempo_experiencia_meses INTEGER,
    experiencia_comprovada_ct BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Conquistas/Certificados
CREATE TABLE emp_curriculo_conquistas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cpf VARCHAR(14) NOT NULL,
    id_tipo_conquista UUID REFERENCES emp_tipos_conquista(id),
    titulo VARCHAR(500) NOT NULL,
    descricao TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Situação e Interesses
CREATE TABLE emp_curriculo_situacao_interesses (
    cpf VARCHAR(14) PRIMARY KEY,
    id_situacao UUID REFERENCES emp_situacoes_atual(id),
    tempo_procurando_emprego VARCHAR(50),
    id_disponibilidade UUID REFERENCES emp_disponibilidades(id),
    ids_tipos_vinculo_preferencia JSONB, -- array de UUIDs
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_emp_cur_formacoes_cpf ON emp_curriculo_formacoes(cpf);
CREATE INDEX idx_emp_cur_idiomas_cpf ON emp_curriculo_idiomas(cpf);
CREATE INDEX idx_emp_cur_cursos_cpf ON emp_curriculo_cursos_complementares(cpf);
CREATE INDEX idx_emp_cur_exp_cpf ON emp_curriculo_experiencias(cpf);
CREATE INDEX idx_emp_cur_conq_cpf ON emp_curriculo_conquistas(cpf);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_emp_cur_conq_cpf;
DROP INDEX IF EXISTS idx_emp_cur_exp_cpf;
DROP INDEX IF EXISTS idx_emp_cur_cursos_cpf;
DROP INDEX IF EXISTS idx_emp_cur_idiomas_cpf;
DROP INDEX IF EXISTS idx_emp_cur_formacoes_cpf;

DROP TABLE IF EXISTS emp_curriculo_situacao_interesses;
DROP TABLE IF EXISTS emp_curriculo_conquistas;
DROP TABLE IF EXISTS emp_curriculo_experiencias;
DROP TABLE IF EXISTS emp_curriculo_cursos_complementares;
DROP TABLE IF EXISTS emp_curriculo_idiomas;
DROP TABLE IF EXISTS emp_curriculo_formacoes;

-- +goose StatementEnd
