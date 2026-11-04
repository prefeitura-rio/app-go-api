-- +goose Up
-- +goose StatementBegin

-- Cascata só nos dependentes. Apagar uma área não apaga o conhecimento, e o contrário também não.
DO $$
DECLARE r record;
BEGIN
  FOR r IN
    SELECT c.conname AS name, t.relname AS table_name
    FROM pg_constraint c
    JOIN pg_class t ON t.oid = c.conrelid
    JOIN pg_namespace n ON n.oid = t.relnamespace
    WHERE c.contype = 'f'
      AND n.nspname = current_schema()
      AND t.relname IN (
        'area_atuacao_habilidade',
        'emp_curriculo_area_atuacao_habilidade',
        'emp_curriculo_comportamento_atitudes'
      )
  LOOP
    EXECUTE format('ALTER TABLE %I DROP CONSTRAINT %I', r.table_name, r.name);
  END LOOP;
END $$;

ALTER TABLE area_atuacao_habilidade
  ADD CONSTRAINT fk_aah_habilidade
  FOREIGN KEY (id_habilidade) REFERENCES emp_habilidades(id) ON DELETE CASCADE;

ALTER TABLE area_atuacao_habilidade
  ADD CONSTRAINT fk_aah_area
  FOREIGN KEY (id_area_atuacao) REFERENCES emp_areas_atuacao(id) ON DELETE CASCADE;

ALTER TABLE emp_curriculo_area_atuacao_habilidade
  ADD CONSTRAINT fk_curriculo_aah
  FOREIGN KEY (id_area_atuacao_habilidade) REFERENCES area_atuacao_habilidade(id) ON DELETE CASCADE;

ALTER TABLE emp_curriculo_comportamento_atitudes
  ADD CONSTRAINT fk_curriculo_comportamento
  FOREIGN KEY (id_comportamento_atitudes) REFERENCES emp_comportamento_atitudes(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_aah_id_area_atuacao
  ON area_atuacao_habilidade (id_area_atuacao);

CREATE TABLE IF NOT EXISTS emp_curriculo_area_atuacao (
    cpf CHAR(11) PRIMARY KEY,
    id_area_atuacao BIGINT NOT NULL REFERENCES emp_areas_atuacao(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

DROP TRIGGER IF EXISTS update_emp_curriculo_area_atuacao_updated_at ON emp_curriculo_area_atuacao;
CREATE TRIGGER update_emp_curriculo_area_atuacao_updated_at
    BEFORE UPDATE ON emp_curriculo_area_atuacao
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Quem já marcou pares de mais de uma área fica com a área que tem mais vínculos.
-- Empate fica com o menor id_area_atuacao. Vínculos das outras áreas saem.
WITH ranked AS (
    SELECT
        c.cpf,
        aah.id_area_atuacao,
        ROW_NUMBER() OVER (
            PARTITION BY c.cpf
            ORDER BY COUNT(*) DESC, aah.id_area_atuacao ASC
        ) AS rn
    FROM emp_curriculo_area_atuacao_habilidade c
    JOIN area_atuacao_habilidade aah ON aah.id = c.id_area_atuacao_habilidade
    GROUP BY c.cpf, aah.id_area_atuacao
),
chosen AS (
    SELECT cpf, id_area_atuacao FROM ranked WHERE rn = 1
),
removed AS (
    DELETE FROM emp_curriculo_area_atuacao_habilidade c
    USING area_atuacao_habilidade aah, chosen ch
    WHERE c.id_area_atuacao_habilidade = aah.id
      AND c.cpf = ch.cpf
      AND aah.id_area_atuacao <> ch.id_area_atuacao
    RETURNING c.cpf
)
INSERT INTO emp_curriculo_area_atuacao (cpf, id_area_atuacao)
SELECT cpf, id_area_atuacao FROM chosen
ON CONFLICT (cpf) DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS update_emp_curriculo_area_atuacao_updated_at ON emp_curriculo_area_atuacao;
DROP TABLE IF EXISTS emp_curriculo_area_atuacao;
DROP INDEX IF EXISTS idx_aah_id_area_atuacao;

ALTER TABLE area_atuacao_habilidade DROP CONSTRAINT IF EXISTS fk_aah_habilidade;
ALTER TABLE area_atuacao_habilidade DROP CONSTRAINT IF EXISTS fk_aah_area;
ALTER TABLE emp_curriculo_area_atuacao_habilidade DROP CONSTRAINT IF EXISTS fk_curriculo_aah;
ALTER TABLE emp_curriculo_comportamento_atitudes DROP CONSTRAINT IF EXISTS fk_curriculo_comportamento;

ALTER TABLE area_atuacao_habilidade
  ADD CONSTRAINT fk_aah_habilidade
  FOREIGN KEY (id_habilidade) REFERENCES emp_habilidades(id) ON DELETE RESTRICT;

ALTER TABLE area_atuacao_habilidade
  ADD CONSTRAINT fk_aah_area
  FOREIGN KEY (id_area_atuacao) REFERENCES emp_areas_atuacao(id) ON DELETE RESTRICT;

ALTER TABLE emp_curriculo_area_atuacao_habilidade
  ADD CONSTRAINT fk_curriculo_aah
  FOREIGN KEY (id_area_atuacao_habilidade) REFERENCES area_atuacao_habilidade(id) ON DELETE RESTRICT;

ALTER TABLE emp_curriculo_comportamento_atitudes
  ADD CONSTRAINT fk_curriculo_comportamento
  FOREIGN KEY (id_comportamento_atitudes) REFERENCES emp_comportamento_atitudes(id) ON DELETE RESTRICT;

-- +goose StatementEnd
