-- +goose Up
-- +goose StatementBegin

ALTER TABLE emp_escolaridades ADD COLUMN ordem INTEGER;
ALTER TABLE emp_niveis_idioma ADD COLUMN ordem INTEGER;

UPDATE emp_escolaridades SET ordem = 1 WHERE descricao = 'Fundamental incompleto';
UPDATE emp_escolaridades SET ordem = 2 WHERE descricao = 'Fundamental completo';
UPDATE emp_escolaridades SET ordem = 3 WHERE descricao = 'Médio incompleto';
UPDATE emp_escolaridades SET ordem = 4 WHERE descricao = 'Médio completo';
UPDATE emp_escolaridades SET ordem = 5 WHERE descricao = 'Técnico';
UPDATE emp_escolaridades SET ordem = 6 WHERE descricao = 'Superior incompleto';
UPDATE emp_escolaridades SET ordem = 7 WHERE descricao = 'Superior completo';
UPDATE emp_escolaridades SET ordem = 8 WHERE descricao = 'Pós-graduação/MBA';
UPDATE emp_escolaridades SET ordem = 9 WHERE descricao = 'Mestrado';
UPDATE emp_escolaridades SET ordem = 10 WHERE descricao = 'Doutorado';

UPDATE emp_niveis_idioma SET ordem = 1 WHERE descricao = 'Básico';
UPDATE emp_niveis_idioma SET ordem = 2 WHERE descricao = 'Intermediário';
UPDATE emp_niveis_idioma SET ordem = 3 WHERE descricao = 'Avançado';
UPDATE emp_niveis_idioma SET ordem = 4 WHERE descricao = 'Nativo/Fluente';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE emp_escolaridades DROP COLUMN IF EXISTS ordem;
ALTER TABLE emp_niveis_idioma DROP COLUMN IF EXISTS ordem;

-- +goose StatementEnd
