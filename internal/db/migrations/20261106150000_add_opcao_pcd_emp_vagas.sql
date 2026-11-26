-- +goose Up
-- Opção PcD da vaga (DEV-407, história PREF-632): aberta a todos, preferencial
-- ou exclusiva para pessoas com deficiência. Substitui o campo descontinuado
-- acessibilidade_pcd, que fica (expand/contract): a API continua a preenchê-lo
-- a partir de opcao_pcd para quem ainda o lê, e o código anterior a esta
-- migração segue funcionando com a coluna nova no seu valor padrão.
ALTER TABLE emp_vagas
    ADD COLUMN IF NOT EXISTS opcao_pcd VARCHAR(20) NOT NULL DEFAULT 'aberta';

ALTER TABLE emp_vagas
    DROP CONSTRAINT IF EXISTS chk_emp_vagas_opcao_pcd;
ALTER TABLE emp_vagas
    ADD CONSTRAINT chk_emp_vagas_opcao_pcd
    CHECK (opcao_pcd IN ('aberta', 'preferencial', 'exclusiva'));

-- De/para das vagas antigas (o mesmo de OpcaoPCDDeAcessibilidade no código):
-- "Vaga PcD" desmarcada (acessibilidade_pcd nula) fica aberta, o padrão;
-- marcada + preferencial vira preferencial; marcada + exclusiva vira exclusiva;
-- marcada sem modalidade (para_pcd, herdado do antigo vaga_pcd = true) vira
-- preferencial, que não bloqueia ninguém. Só altera vagas ainda em aberta, então
-- rodar de novo não muda nada.
--
-- A exclusiva migrada fica com o bloqueio PcD desligado: a chave "pcd" de
-- criterios_bloqueio só passa a existir com esta versão, então nenhuma vaga
-- antiga a tem, e esta migração não a liga.
UPDATE emp_vagas
SET opcao_pcd = CASE acessibilidade_pcd
                    WHEN 'exclusivo_pcd' THEN 'exclusiva'
                    ELSE 'preferencial'
                END
WHERE acessibilidade_pcd IS NOT NULL
  AND opcao_pcd = 'aberta';

-- +goose Down
ALTER TABLE emp_vagas
    DROP CONSTRAINT IF EXISTS chk_emp_vagas_opcao_pcd;
ALTER TABLE emp_vagas
    DROP COLUMN IF EXISTS opcao_pcd;
