-- +goose Up
-- Atualizações das candidaturas para o cidadão (história PREF-574): o card
-- "Minhas candidaturas" do superapp mostra quantas candidaturas mudaram de
-- status ou de etapa desde a última vez que a pessoa abriu a lista. As
-- mudanças vêm do histórico (emp_candidatura_historico); aqui fica só o id do
-- último evento que o candidato viu. O id do histórico cresce na ordem de
-- gravação, então serve de cursor sem depender do relógio da aplicação nem do
-- banco. Coluna anulável (expand, sem reescrever a tabela): candidatura
-- existente começa sem evento visto, e o código anterior a esta migração segue
-- funcionando. A contagem filtra por CPF (idx_emp_candidaturas_cpf) e lê o
-- histórico pelo índice de candidatura.
ALTER TABLE emp_candidaturas
    ADD COLUMN IF NOT EXISTS ultimo_evento_visto BIGINT;

-- +goose Down
ALTER TABLE emp_candidaturas
    DROP COLUMN IF EXISTS ultimo_evento_visto;
