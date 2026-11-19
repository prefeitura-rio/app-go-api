-- +goose Up
-- Backfill do bloqueio PcD (DEV-656, decisão da DEV-444, histórias
-- PREF-631/PREF-632). A migração de opcao_pcd (DEV-407) converteu as vagas
-- antigas "exclusivo_pcd" em opcao_pcd = 'exclusiva', mas deixou de propósito
-- o bloqueio PcD desligado: nenhuma delas tinha a chave "pcd" em
-- criterios_bloqueio. Elas aparecem como exclusivas para PcD e não barram a
-- candidatura de quem não declarou deficiência. Esta migração liga a chave em
-- toda vaga exclusiva que ainda não a tem:
--   - criterios_bloqueio NULL (ou o JSON null) vira ["pcd"];
--   - lista sem "pcd" recebe "pcd" no fim, sem mexer nas outras chaves;
--   - lista que já tem "pcd", vaga aberta ou preferencial e valor que não é
--     lista não são tocados.
-- Só acrescenta (expand): o código anterior já lê a chave. Rodar de novo não
-- muda nada, porque o WHERE deixa de casar com a vaga que já tem "pcd".
UPDATE emp_vagas
SET criterios_bloqueio =
        CASE
            WHEN criterios_bloqueio IS NULL OR jsonb_typeof(criterios_bloqueio) = 'null'
                THEN '["pcd"]'::jsonb
            ELSE criterios_bloqueio || '["pcd"]'::jsonb
        END
WHERE opcao_pcd = 'exclusiva'
  AND (criterios_bloqueio IS NULL
       OR jsonb_typeof(criterios_bloqueio) = 'null'
       OR (jsonb_typeof(criterios_bloqueio) = 'array'
           AND NOT criterios_bloqueio @> '["pcd"]'::jsonb));

-- +goose Down
-- Correção de dados: não há o que desfazer. Não se sabe mais quais vagas
-- ganharam a chave aqui e quais a tiveram ligada no Admin depois, e tirá-la
-- desligaria o bloqueio de candidatura PcD dessas vagas.
