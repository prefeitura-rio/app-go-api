-- +goose Up
-- Opção Jovem Aprendiz da seção Idade da vaga (DEV-428, história PREF-634):
-- faixa fixa de 14 a 24 anos, sem número livre. A opção Maiores de 18 continua
-- sendo a coluna idade_minima. O bloqueio de candidatura é a chave
-- "jovem_aprendiz" de criterios_bloqueio. Só acrescenta uma coluna com padrão
-- FALSE (expand, sem reescrever a tabela): toda vaga existente nasce com
-- Jovem Aprendiz desligado e o código anterior a esta migração segue
-- funcionando, porque o INSERT dele cai no padrão.
ALTER TABLE emp_vagas
    ADD COLUMN IF NOT EXISTS jovem_aprendiz BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE emp_vagas
    DROP COLUMN IF EXISTS jovem_aprendiz;
