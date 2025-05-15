-- +goose Up
-- +goose StatementBegin
CREATE TABLE orgaos (
    id SERIAL PRIMARY KEY,
    nome VARCHAR(255) NOT NULL
);

CREATE TABLE instituicoes_ensino (
    id SERIAL PRIMARY KEY,
    nome VARCHAR(255) NOT NULL
);

CREATE TABLE empresas (
    id SERIAL PRIMARY KEY,
    nome VARCHAR(255) NOT NULL
);

CREATE TABLE escolaridades (
    id SERIAL PRIMARY KEY,
    nivel VARCHAR(100) NOT NULL
);

CREATE TABLE categorias (
    id SERIAL PRIMARY KEY,
    nome VARCHAR(100) NOT NULL
);

CREATE TABLE acessibilidades (
    id SERIAL PRIMARY KEY,
    nome VARCHAR(100) NOT NULL
);

CREATE TYPE modalidade_enum AS ENUM ('PRESENCIAL', 'ONLINE', 'HIBRIDO');
CREATE TYPE status_curso_enum AS ENUM ('CRIADO', 'ABERTO', 'ENCERRADO');
CREATE TYPE turno_enum AS ENUM ('MANHA', 'TARDE', 'NOITE', 'INTEGRAL', 'LIVRE');
CREATE TYPE formato_aula_enum AS ENUM ('GRAVADO', 'AO_VIVO');
CREATE TYPE tipo_contratacao_enum AS ENUM ('CLT', 'ESTAGIO', 'JOVEM_APRENDIZ', 'MEI', 'PJ');
CREATE TYPE status_emprego_enum AS ENUM ('CRIADO', 'ABERTO', 'EM_ANALISE', 'ENCERRADO');
CREATE TYPE jornada_trabalho_enum AS ENUM ('INTEGRAL', 'MEIO_PERIODO', 'FLEXIVEL', 'ESTAGIO', 'HOME_OFFICE');

CREATE TABLE cursos (
    id SERIAL PRIMARY KEY,
    titulo VARCHAR(255) NOT NULL,
    descricao TEXT,
    orgao_id INTEGER REFERENCES orgaos(id),
    instituicao_id INTEGER REFERENCES instituicoes_ensino(id),
    modalidade modalidade_enum NOT NULL,
    local_realizacao VARCHAR(255),
    data_inicio TIMESTAMP,
    data_termino TIMESTAMP,
    data_limite_inscricoes TIMESTAMP,
    numero_vagas INTEGER,
    carga_horaria INTEGER,
    pre_requisitos TEXT,
    certificacao_oferecida BOOLEAN DEFAULT false,
    status status_curso_enum NOT NULL,
    turno turno_enum,
    formato_aula formato_aula_enum,
    link_inscricao VARCHAR(255),
    contato_duvidas VARCHAR(255),
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE empregos (
    id SERIAL PRIMARY KEY,
    titulo VARCHAR(255) NOT NULL,
    descricao TEXT,
    orgao_id INTEGER REFERENCES orgaos(id),
    empresa_id INTEGER REFERENCES empresas(id),
    tipo_contratacao tipo_contratacao_enum NOT NULL,
    numero_vagas INTEGER,
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    salario_min INTEGER,
    salario_max INTEGER,
    beneficios TEXT,
    pre_requisitos TEXT,
    data_inicio_prevista TIMESTAMP,
    data_limite_candidatura TIMESTAMP,
    status status_emprego_enum NOT NULL,
    escolaridade_id INTEGER REFERENCES escolaridades(id),
    jornada_trabalho jornada_trabalho_enum,
    turno turno_enum,
    contato_duvidas VARCHAR(255),
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE cursos_categorias (
    curso_id INTEGER REFERENCES cursos(id) ON DELETE CASCADE,
    categoria_id INTEGER REFERENCES categorias(id) ON DELETE CASCADE,
    PRIMARY KEY (curso_id, categoria_id)
);

CREATE TABLE cursos_acessibilidades (
    curso_id INTEGER REFERENCES cursos(id) ON DELETE CASCADE,
    acessibilidade_id INTEGER REFERENCES acessibilidades(id) ON DELETE CASCADE,
    PRIMARY KEY (curso_id, acessibilidade_id)
);

-- Índices para melhorar performance de buscas comuns
CREATE INDEX idx_cursos_orgao ON cursos(orgao_id);
CREATE INDEX idx_cursos_instituicao ON cursos(instituicao_id);
CREATE INDEX idx_cursos_status ON cursos(status);
CREATE INDEX idx_cursos_data_inicio ON cursos(data_inicio);
CREATE INDEX idx_cursos_data_limite_inscricoes ON cursos(data_limite_inscricoes);

CREATE INDEX idx_empregos_orgao ON empregos(orgao_id);
CREATE INDEX idx_empregos_empresa ON empregos(empresa_id);
CREATE INDEX idx_empregos_escolaridade ON empregos(escolaridade_id);
CREATE INDEX idx_empregos_status ON empregos(status);
CREATE INDEX idx_empregos_tipo_contratacao ON empregos(tipo_contratacao);
CREATE INDEX idx_empregos_data_limite_candidatura ON empregos(data_limite_candidatura);
CREATE INDEX idx_empregos_localizacao ON empregos(latitude, longitude);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS cursos_acessibilidades;
DROP TABLE IF EXISTS cursos_categorias;
DROP TABLE IF EXISTS empregos;
DROP TABLE IF EXISTS cursos;
DROP TYPE IF EXISTS jornada_trabalho_enum;
DROP TYPE IF EXISTS status_emprego_enum;
DROP TYPE IF EXISTS tipo_contratacao_enum;
DROP TYPE IF EXISTS formato_aula_enum;
DROP TYPE IF EXISTS turno_enum;
DROP TYPE IF EXISTS status_curso_enum;
DROP TYPE IF EXISTS modalidade_enum;
DROP TABLE IF EXISTS acessibilidades;
DROP TABLE IF EXISTS categorias;
DROP TABLE IF EXISTS escolaridades;
DROP TABLE IF EXISTS empresas;
DROP TABLE IF EXISTS instituicoes_ensino;
DROP TABLE IF EXISTS orgaos;
-- +goose StatementEnd 