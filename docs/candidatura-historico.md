# Histórico de etapas e status das candidaturas

`emp_candidatura_historico` guarda um evento por transição de cada candidatura de
empregabilidade (DEV-532 / PREF-687). A API grava o evento **na mesma transação** que muda
`emp_candidaturas`: se o evento falha, a mudança é desfeita.

## Regras

- **Só de inclusão.** Triggers recusam `UPDATE`, `DELETE` e `TRUNCATE` na tabela. Não há
  chave estrangeira: a exclusão de uma vaga (que apaga candidaturas e etapas em cascata) não
  apaga o histórico.
- **Sem dado pessoal.** Só ids (`id_candidatura`, `id_vaga`, etapas), códigos e datas; nada de
  CPF, nome ou e-mail. `motivo` é um código (`^[a-z0-9_]+$`), nunca texto livre.
- **Um evento por mudança efetiva.** Pedir o status ou a etapa que a candidatura já tem não gera
  evento. Operações em lote geram um evento por candidatura que mudou.
- **Sem migração retroativa.** Vale a partir da entrada em produção. O estado "anterior" é lido
  da linha travada (`SELECT ... FOR UPDATE`) na transação: o primeiro evento de uma candidatura
  antiga traz o estado de origem em `status_anterior` / `id_etapa_anterior`.

## Colunas

| Coluna | Conteúdo |
|---|---|
| `id` | `BIGSERIAL`, crescente: ordem de gravação e cursor de sincronização |
| `id_candidatura`, `id_vaga` | a candidatura e a vaga dela |
| `tipo_evento` | ver tabela abaixo |
| `ator` | `candidato`, `operador` ou `automacao` |
| `status_anterior`, `status_novo` | status antes e depois (iguais em eventos de etapa; `status_anterior` nulo na criação) |
| `id_etapa_anterior`, `id_etapa_nova` | etapa antes e depois (iguais em eventos de status: a etapa em que a candidatura estava) |
| `motivo` | código da automação (`vaga_congelada`, `vaga_descongelada`, `vaga_descontinuada`, `vaga_reativada`, `etapa_excluida`) ou nulo |
| `created_at` | quando a transição aconteceu (`TIMESTAMPTZ`) |

| `tipo_evento` | Quando | Ator |
|---|---|---|
| `candidatura_criada` | candidatura nova (app, cadastro pela secretaria, associação pelo banco de currículos) | `candidato` (app) ou `operador` |
| `mudanca_etapa` | ida para etapa de ordem maior, ou a primeira etapa | `operador` |
| `volta_etapa` | ida para etapa de ordem menor | `operador` |
| `etapa_removida` | a etapa em que a candidatura estava foi excluída da vaga; ela fica sem etapa | `automacao` |
| `candidatura_aprovada` | status → `aprovada` | `operador` |
| `candidatura_reprovada` | status → `reprovada`: manual (`operador`) ou automática (`automacao`, com `motivo`) | `operador` / `automacao` |
| `candidatura_cancelada` | exclusão da candidatura (desistência ou remoção) | `candidato` (a própria pessoa) ou `operador` |
| `status_alterado` | outro status: reabertura, vaga congelada/descongelada/descontinuada/reativada | `operador` / `automacao` |

`emp_etapas.tipo_etapa` (opcional, até 64 caracteres) classifica a etapa com um código definido
pela gestão, para comparar o funil entre vagas com etapas de nomes diferentes. Numa atualização,
ausente ou vazio mantém o valor gravado.

## Consultas

As três consultas abaixo são exercitadas pelo teste de integração
`internal/repository/empregabilidade/candidatura_historico_integration_test.go`. Para agrupar por
tipo de etapa em vez de etapa, junte `emp_etapas` por `id_etapa_nova` e agrupe por `tipo_etapa`.

Funil por etapa (candidaturas distintas que chegaram a cada etapa da vaga):

```sql
SELECT h.id_etapa_nova AS id_etapa, COUNT(DISTINCT h.id_candidatura) AS candidaturas
FROM emp_candidatura_historico h
WHERE h.id_vaga = :vaga
  AND h.id_etapa_nova IS NOT NULL
  AND h.tipo_evento IN ('candidatura_criada', 'mudanca_etapa', 'volta_etapa')
GROUP BY h.id_etapa_nova;
```

Tempo médio em cada etapa (de entrar na etapa até a próxima mudança de etapa, a decisão ou o
cancelamento; passagens ainda abertas ficam de fora):

```sql
WITH eventos AS (
    SELECT tipo_evento, id_etapa_nova, created_at,
           LEAD(created_at) OVER (PARTITION BY id_candidatura ORDER BY created_at, id) AS saiu_em
    FROM emp_candidatura_historico
    WHERE id_vaga = :vaga
      AND tipo_evento IN ('candidatura_criada', 'mudanca_etapa', 'volta_etapa', 'etapa_removida',
                          'candidatura_aprovada', 'candidatura_reprovada', 'candidatura_cancelada')
)
SELECT id_etapa_nova AS id_etapa,
       EXTRACT(EPOCH FROM AVG(saiu_em - created_at)) AS segundos_medios,
       COUNT(*) AS passagens
FROM eventos
WHERE tipo_evento IN ('candidatura_criada', 'mudanca_etapa', 'volta_etapa')
  AND id_etapa_nova IS NOT NULL
  AND saiu_em IS NOT NULL
GROUP BY id_etapa_nova;
```

Reprovados por etapa e motivo:

```sql
SELECT id_etapa_nova AS id_etapa, ator, COALESCE(motivo, 'sem_motivo') AS motivo, COUNT(*) AS reprovacoes
FROM emp_candidatura_historico
WHERE id_vaga = :vaga AND tipo_evento = 'candidatura_reprovada'
GROUP BY id_etapa_nova, ator, COALESCE(motivo, 'sem_motivo');
```

## Sincronização para o BigQuery (`brutos_go`)

O banco `go` (Cloud SQL do `superapp`) chega ao BigQuery pelo Airbyte do cluster do `superapp`
(namespace `airbyte`), com o usuário `airbyte_cdc` e a publicação `airbyte_publication` (slot
`airbyte_slot`). A conexão é configurada na interface do Airbyte, fora deste repositório; a
migration não mexe na publicação porque o usuário `go` não é dono dela. Depois do deploy da
migration, quem administra o Airbyte e o banco:

1. Confere se a publicação já cobre a tabela nova
   (`SELECT * FROM pg_publication_tables WHERE pubname = 'airbyte_publication' AND tablename = 'emp_candidatura_historico';`).
   Se ela não for `FOR ALL TABLES`, o dono da publicação inclui a tabela:
   `ALTER PUBLICATION airbyte_publication ADD TABLE public.emp_candidatura_historico;`
2. Garante a leitura ao Airbyte: `GRANT SELECT ON public.emp_candidatura_historico TO airbyte_cdc;`
3. Na conexão do banco `go` para o destino BigQuery `brutos_go`, atualiza o schema da fonte
   (*Refresh source schema*), ativa o stream `emp_candidatura_historico` com modo
   **Incremental | Append** pela replicação lógica (CDC, chave primária `id`) e aceita a coluna
   nova `tipo_etapa` do stream `emp_etapas`. Sem CDC, use **Full Refresh | Overwrite**: um cursor
   em `id` ou `created_at` pode pular eventos, porque o valor é atribuído antes do commit e as
   transações terminam fora de ordem.
4. Roda uma sincronização e confere a tabela `brutos_go.emp_candidatura_historico` no BigQuery.

Como a tabela é só de inclusão, o CDC em modo Append entrega cada evento uma vez, na ordem de commit.
