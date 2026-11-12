-- +goose Up
-- Habilidades e competências cadastradas pela secretaria em cada vaga (DEV-528):
-- uma ou mais áreas de atuação, os conhecimentos de cada área e os
-- comportamentos/atitudes, todos ligados ao id_vaga. Segue o padrão das tabelas
-- de habilidades do currículo (catálogos emp_areas_atuacao, emp_habilidades,
-- area_atuacao_habilidade e emp_comportamento_atitudes). Só cria tabelas novas.
--
-- A chave primária de cada tabela começa por id_vaga, então ela também é o
-- índice em id_vaga (leitura por vaga e cascata da exclusão da vaga).

-- 1. Áreas de atuação da vaga: cada área uma vez por vaga.
CREATE TABLE IF NOT EXISTS emp_vagas_areas_atuacao (
    id_vaga UUID NOT NULL REFERENCES emp_vagas(id) ON DELETE CASCADE,
    id_area_atuacao BIGINT NOT NULL REFERENCES emp_areas_atuacao(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT pk_emp_vagas_areas_atuacao PRIMARY KEY (id_vaga, id_area_atuacao)
);

CREATE INDEX IF NOT EXISTS idx_emp_vagas_areas_atuacao_area
    ON emp_vagas_areas_atuacao (id_area_atuacao);

DROP TRIGGER IF EXISTS update_emp_vagas_areas_atuacao_updated_at ON emp_vagas_areas_atuacao;
CREATE TRIGGER update_emp_vagas_areas_atuacao_updated_at
    BEFORE UPDATE ON emp_vagas_areas_atuacao
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 2. Conhecimentos de cada área da vaga: cada conhecimento uma vez por área da
-- vaga. A FK composta (id_vaga, id_area_atuacao) exige a área já registrada na
-- vaga e prende o conhecimento à área em que foi marcado; tirar a área da vaga
-- tira os conhecimentos dela. A FK (id_habilidade, id_area_atuacao) exige que o
-- conhecimento pertença à área no catálogo, como no currículo.
CREATE TABLE IF NOT EXISTS emp_vagas_conhecimentos (
    id_vaga UUID NOT NULL,
    id_area_atuacao BIGINT NOT NULL,
    id_habilidade BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT pk_emp_vagas_conhecimentos PRIMARY KEY (id_vaga, id_area_atuacao, id_habilidade),
    CONSTRAINT fk_emp_vagas_conhecimentos_vaga_area
        FOREIGN KEY (id_vaga, id_area_atuacao)
        REFERENCES emp_vagas_areas_atuacao (id_vaga, id_area_atuacao) ON DELETE CASCADE,
    CONSTRAINT fk_emp_vagas_conhecimentos_area_habilidade
        FOREIGN KEY (id_habilidade, id_area_atuacao)
        REFERENCES area_atuacao_habilidade (id_habilidade, id_area_atuacao) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_emp_vagas_conhecimentos_habilidade_area
    ON emp_vagas_conhecimentos (id_habilidade, id_area_atuacao);

DROP TRIGGER IF EXISTS update_emp_vagas_conhecimentos_updated_at ON emp_vagas_conhecimentos;
CREATE TRIGGER update_emp_vagas_conhecimentos_updated_at
    BEFORE UPDATE ON emp_vagas_conhecimentos
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 3. Comportamentos e atitudes da vaga: cada um uma vez por vaga.
CREATE TABLE IF NOT EXISTS emp_vagas_comportamento_atitudes (
    id_vaga UUID NOT NULL REFERENCES emp_vagas(id) ON DELETE CASCADE,
    id_comportamento_atitudes BIGINT NOT NULL REFERENCES emp_comportamento_atitudes(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT pk_emp_vagas_comportamento_atitudes PRIMARY KEY (id_vaga, id_comportamento_atitudes)
);

CREATE INDEX IF NOT EXISTS idx_emp_vagas_comportamento_atitudes_comportamento
    ON emp_vagas_comportamento_atitudes (id_comportamento_atitudes);

DROP TRIGGER IF EXISTS update_emp_vagas_comportamento_atitudes_updated_at ON emp_vagas_comportamento_atitudes;
CREATE TRIGGER update_emp_vagas_comportamento_atitudes_updated_at
    BEFORE UPDATE ON emp_vagas_comportamento_atitudes
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- +goose Down
DROP TABLE IF EXISTS emp_vagas_conhecimentos;
DROP TABLE IF EXISTS emp_vagas_comportamento_atitudes;
DROP TABLE IF EXISTS emp_vagas_areas_atuacao;
