-- +goose Up
-- Classificação das etapas existentes (DEV-688, PREF-687). tipo_etapa passa a
-- ser um de analise_curriculo, entrevista, teste, proposta ou outro, e a API
-- infere o tipo do título quando a etapa nova chega sem ele. Esta migration
-- aplica a mesma regra às etapas que ficaram sem tipo (todas as anteriores ao
-- DEV-532), com os padrões de empregabilidade.RegrasTipoEtapaPorTitulo, na
-- mesma ordem (a primeira que casa vence; nenhuma, outro):
--   - o título não muda: continua livre;
--   - tipo gravado que só difere da lista por espaços ou maiúsculas é
--     normalizado; qualquer outro tipo já gravado não é tocado;
--   - updated_at não muda: a etapa não foi editada por ninguém.
-- Só preenche (expand): o código anterior aceita qualquer valor na coluna.
-- Rodar de novo não muda nada, porque o WHERE deixa de casar.
UPDATE emp_etapas
SET tipo_etapa = lower(btrim(tipo_etapa))
WHERE tipo_etapa IS NOT NULL
  AND tipo_etapa <> lower(btrim(tipo_etapa))
  AND lower(btrim(tipo_etapa)) IN ('analise_curriculo', 'entrevista', 'teste', 'proposta', 'outro');

UPDATE emp_etapas
SET tipo_etapa =
        CASE
            WHEN titulo ~* 'curr[iíÍ]cul|triagem' THEN 'analise_curriculo'
            WHEN titulo ~* 'entrevista' THEN 'entrevista'
            WHEN titulo ~* 'teste|prova|avalia|din[aâÂ]mica' THEN 'teste'
            WHEN titulo ~* 'proposta|oferta|contrata|admiss' THEN 'proposta'
            ELSE 'outro'
        END
WHERE tipo_etapa IS NULL OR btrim(tipo_etapa) = '';

-- +goose Down
-- Correção de dados: não há o que desfazer. Não se sabe mais quais etapas
-- ganharam o tipo aqui e quais o receberam pela API depois, e o código
-- anterior convive com a coluna preenchida.
