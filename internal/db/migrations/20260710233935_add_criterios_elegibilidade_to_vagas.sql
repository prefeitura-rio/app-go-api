-- +goose Up
-- +goose StatementBegin

ALTER TABLE emp_vagas ADD COLUMN idade_minima INTEGER;
ALTER TABLE emp_vagas ADD COLUMN idade_maxima INTEGER;
ALTER TABLE emp_vagas ADD COLUMN bairros_elegibilidade JSONB;
ALTER TABLE emp_vagas ADD COLUMN id_escolaridade_minima UUID REFERENCES emp_escolaridades(id);
ALTER TABLE emp_vagas ADD COLUMN areas_formacao_elegibilidade JSONB;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE emp_vagas DROP COLUMN IF EXISTS idade_minima;
ALTER TABLE emp_vagas DROP COLUMN IF EXISTS idade_maxima;
ALTER TABLE emp_vagas DROP COLUMN IF EXISTS bairros_elegibilidade;
ALTER TABLE emp_vagas DROP COLUMN IF EXISTS id_escolaridade_minima;
ALTER TABLE emp_vagas DROP COLUMN IF EXISTS areas_formacao_elegibilidade;

-- +goose StatementEnd
