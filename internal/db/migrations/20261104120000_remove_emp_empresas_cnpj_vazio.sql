-- +goose Up
-- DEV-261: o POST /empregabilidade/empresas não validava o CNPJ e gravou
-- empresas com chave primária em branco. Nenhuma rota /empresas/:cnpj alcança
-- essa chave, então a API não consegue corrigi-las nem removê-las. Apaga só as
-- que nenhuma vaga referencia (emp_vagas.id_contratante); uma referenciada fica
-- para correção manual, sem quebrar a chave estrangeira.
DELETE FROM emp_empresas e
WHERE TRIM(e.cnpj) = ''
  AND NOT EXISTS (SELECT 1 FROM emp_vagas v WHERE v.id_contratante = e.cnpj);

-- +goose Down
-- Correção de dados: não há o que desfazer.
