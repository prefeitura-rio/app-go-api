-- +goose Up
-- Classificação das etapas existentes (DEV-688, PREF-687). tipo_etapa passa a
-- ser um de analise_curriculo, entrevista, teste, proposta ou outro, e a API
-- infere o tipo do título quando a etapa nova chega sem ele. Esta migration
-- aplica a mesma regra às etapas existentes, com os padrões de
-- empregabilidade.RegrasTipoEtapaPorTitulo, na mesma ordem (a primeira que
-- casa vence; nenhuma, outro), casando no começo de uma palavra (\m):
--   - o título não muda: continua livre;
--   - tipo gravado que só difere da lista por espaços ou maiúsculas é
--     normalizado e mantido;
--   - sem tipo (todas as etapas anteriores ao DEV-532), tipo vazio ou código
--     livre fora da lista (aceito entre o DEV-532 e esta versão, e que a API
--     agora recusa ao reenviar a vaga) recebem o tipo do título;
--   - updated_at não muda: a etapa não foi editada por ninguém.
-- Só dados, sem DDL (expand): o código anterior aceita qualquer valor na coluna.
-- Rodar de novo não muda nada, porque o WHERE deixa de casar.
UPDATE emp_etapas
SET tipo_etapa = lower(btrim(tipo_etapa))
WHERE tipo_etapa IS NOT NULL
  AND tipo_etapa <> lower(btrim(tipo_etapa))
  AND lower(btrim(tipo_etapa)) IN ('analise_curriculo', 'entrevista', 'teste', 'proposta', 'outro');

UPDATE emp_etapas
SET tipo_etapa =
        CASE
            WHEN titulo ~* '\m(curr[iíÍ]cul|triagem)' THEN 'analise_curriculo'
            WHEN titulo ~* '\m(entrevista)' THEN 'entrevista'
            WHEN titulo ~* '\m(teste|prova|avalia|din[aâÂ]mica)' THEN 'teste'
            WHEN titulo ~* '\m(proposta|oferta|contrata|admiss)' THEN 'proposta'
            ELSE 'outro'
        END
WHERE tipo_etapa IS NULL
   OR tipo_etapa NOT IN ('analise_curriculo', 'entrevista', 'teste', 'proposta', 'outro');

-- +goose Down
-- Correção de dados: não há o que desfazer. Não se sabe mais quais etapas
-- ganharam o tipo aqui e quais o receberam pela API depois, e o código
-- anterior convive com a coluna preenchida.
