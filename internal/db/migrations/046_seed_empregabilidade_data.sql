-- +goose Up
-- +goose StatementBegin

-- Seed Regimes de Contratação
INSERT INTO emp_regimes_contratacao (descricao) VALUES
    ('CLT'),
    ('Estágio'),
    ('Aprendiz'),
    ('Freelance'),
    ('Temporário'),
    ('PJ'),
    ('Voluntário');

-- Seed Modelos de Trabalho
INSERT INTO emp_modelos_trabalho (descricao) VALUES
    ('Remoto'),
    ('Presencial'),
    ('Híbrido');

-- Seed Tipos de PCD
INSERT INTO emp_tipos_pcd (descricao) VALUES
    ('Física'),
    ('Visual'),
    ('Auditiva'),
    ('Intelectual'),
    ('Múltipla');

-- Seed Idiomas
INSERT INTO emp_idiomas (descricao) VALUES
    ('Português'),
    ('Inglês'),
    ('Espanhol'),
    ('Italiano'),
    ('Alemão'),
    ('Francês'),
    ('Outro');

-- Seed Níveis de Idioma
INSERT INTO emp_niveis_idioma (descricao) VALUES
    ('Básico'),
    ('Intermediário'),
    ('Avançado'),
    ('Nativo/Fluente');

-- Seed Escolaridades
INSERT INTO emp_escolaridades (descricao) VALUES
    ('Fundamental incompleto'),
    ('Fundamental completo'),
    ('Médio incompleto'),
    ('Médio completo'),
    ('Técnico'),
    ('Superior incompleto'),
    ('Superior completo'),
    ('Pós-graduação/MBA'),
    ('Mestrado'),
    ('Doutorado');

-- Seed Tipos de Conquista
INSERT INTO emp_tipos_conquista (descricao) VALUES
    ('Certificado'),
    ('Curso'),
    ('Reconhecimento'),
    ('Trabalho voluntário');

-- Seed Situações Atuais
INSERT INTO emp_situacoes_atual (descricao) VALUES
    ('Buscando primeiro emprego'),
    ('Empregado(a)'),
    ('Desempregado(a)'),
    ('Autônomo(a)'),
    ('Estudante'),
    ('Aposentado(a)');

-- Seed Disponibilidades
INSERT INTO emp_disponibilidades (descricao) VALUES
    ('Imediato'),
    ('Em até 15 dias'),
    ('Em até 30 dias'),
    ('Em até 60 dias'),
    ('Acima de 60 dias');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DELETE FROM emp_disponibilidades;
DELETE FROM emp_situacoes_atual;
DELETE FROM emp_tipos_conquista;
DELETE FROM emp_escolaridades;
DELETE FROM emp_niveis_idioma;
DELETE FROM emp_idiomas;
DELETE FROM emp_tipos_pcd;
DELETE FROM emp_modelos_trabalho;
DELETE FROM emp_regimes_contratacao;

-- +goose StatementEnd
