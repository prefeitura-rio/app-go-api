# Runbook operacional — app-go-api

Guia para quem está de plantão quando a API de cursos, vagas e MEI (`app-go-api`) dá problema em
staging ou produção: o que o serviço faz, onde olhar, o que cada alerta significa, como voltar uma
versão, reiniciar com segurança e reprocessar o que ficou pelo caminho.

Este documento não tem segredos nem credenciais. Os valores sensíveis ficam no Infisical (Secret
`go-secrets`) e nunca devem ser copiados para issue, chat ou log. Ao investigar, nunca leia o
conteúdo de um `Secret` do Kubernetes nem cole dados de cidadãos (CPF, nome, e-mail, telefone).

> Regra geral: **leitura para diagnóstico pode** (`kubectl get|describe|logs`, `gh run view`);
> **toda mudança de configuração vai por PR** no repositório que a define (este, para `k8s/`, ou o
> `prefeitura-rio/infra`, para Istio, Redis, Typesense, Infisical e SigNoz). Comandos de escrita
> no cluster citados abaixo são para a pessoa de plantão com permissão, e só em incidente.

## 1. Visão rápida

| Item | Valor |
|---|---|
| O que faz | API REST (Go/Gin) de cursos, inscrições, vagas de emprego, empregabilidade (currículos, vagas, candidaturas) e oportunidades MEI |
| Quem consome | `superapp` (cidadão, `pref.rio` / `staging.app.dados.rio`), ferramentas internas (`admin.pref.rio` / `admin.staging.app.dados.rio`, Portal Interno) e o `app-catalogo` |
| Entrada pública | `https://services.pref.rio/go/...` (produção) e `https://services.staging.app.dados.rio/go/...` (staging). O Istio tira o prefixo `/go` antes de chegar ao pod |
| Autenticação | Feita fora do serviço: Istio + `ext_authz` (`cerbos-authz`) + Keycloak. A API decodifica o JWT de `X-Auth-Request-Token` sem validar assinatura |
| Cluster / namespace | GKE `application` (`rj-superapp` em produção, `rj-superapp-staging` em staging), namespace `go` |
| Recursos (`k8s/<env>/resources.yaml`) | `Rollout/go`, `Service/go`, `Service/go-preview` (staging) ou `Service/go-canary` (produção), `ScaledObject/go-scaled-object` (KEDA), `PodDisruptionBudget/go-pdb`, `AnalysisTemplate/go-success-rate` (produção) |
| Entrega | Argo CD Application `go` (namespace `argocd`) sincroniza `k8s/staging` ou `k8s/prod` da `main`; a imagem vem do override `spec.source.kustomize.images` que os workflows gravam |

### De quem depende

| Dependência | Onde fica | O que acontece se cair |
|---|---|---|
| PostgreSQL (Cloud SQL, banco `go`) | `DB_*` em `go-secrets` | Pod não sobe (o `goose up` do entrypoint e o `gorm.Open` falham → `CrashLoopBackOff`). Com o pod no ar, as rotas respondem 5xx. O `/health` **não** testa o banco |
| Redis (`Deployment/go-redis`, 1 réplica, no namespace `go`) | `REDIS_*` | Cache vira miss (a API continua, mais lenta); fila de e-mails para; locks distribuídos dos workers falham e o ciclo é pulado; o lembrete D-1 de aula não envia (fail-closed para não duplicar) |
| Typesense (Helm release `typesense`, namespace `go`) | `TYPESENSE_*` | Rotas `/api/v1/typesense/*` (busca de cursos e vagas) falham; o resto da API segue |
| app-rmi | `RMI_BASE_URL` | Ver "RMI fora do ar" na seção 4.5: leitura segue (fail-open), escrita com escopo de secretaria é negada (fail-closed), sincronizações ficam com dado antigo |
| Heimdall (`HEIMDALL_BASE_URL`, timeout 3 s) | `go-secrets` | Todo usuário vira `USER`: administradores recebem 403 nas rotas de gestão. Não há log dessa falha; confirme pelo Heimdall |
| Keycloak (conta de serviço `KEYCLOAK_*`) | `go-secrets` | Sem token, o worker de cidadãos e o resolvedor de secretarias não chamam o RMI (efeito igual ao do RMI fora) |
| Data Relay (`DATA_RELAY_*`) | `go-secrets` | E-mails falham e vão para retentativa; esgotadas, para a fila morta (seção 4.6) |
| Cerbos (`CERBOS_*`, só se `CERBOS_ENABLED=true`) | `go-secrets` | Autorização fina de propostas MEI falha |
| Coletor OTLP do SigNoz | `TRACING_*` em `k8s/<env>/resources.yaml` | Só perde traces; a API segue |

## 2. Onde olhar

### Saúde

- `GET /health` → `{"status":"ok"}`. É **raso**: responde 200 se o processo HTTP está de pé, sem
  testar banco, Redis nem RMI. Usado pelas probes de readiness e liveness e pelo Gatus.
- Público: `curl -s https://services.staging.app.dados.rio/go/health` (staging) ou
  `https://services.pref.rio/go/health` (produção).

### Métricas (`/metrics`)

- `GET /metrics` expõe o formato Prometheus (métricas padrão de Go/processo e as do cliente RMI).
  Pela entrada pública ele responde **403** (bloqueado pelo `ext_authz`); leia de dentro do cluster:
  `kubectl -n go port-forward svc/go 8080:80` e `curl -s localhost:8080/metrics | grep rmi_client_`.
- Não há scrape versionado dessas métricas neste repositório (sem `ServiceMonitor` nem anotação).
- Métricas do cliente RMI (label `client`: `api` para as requisições, `orgao_sync` para o worker):

| Métrica | Leitura |
|---|---|
| `rmi_client_circuit_breaker_state` | `0` fechado, `1` meio-aberto, `2` **aberto** (RMI considerado fora) |
| `rmi_client_circuit_breaker_transitions_total{from,to}` | Quantas vezes o circuito abriu/fechou |
| `rmi_client_circuit_breaker_rejections_total` | Chamadas recusadas sem ir ao RMI |
| `rmi_client_retries_total{reason}` | Novas tentativas (erro de rede, 502/503/504) |
| `rmi_client_secretaria_cache_total{result}` | `hit`/`miss` do cache de secretarias por CPF |

- Tráfego e erros por resposta vêm do Istio, no Prometheus
  `http://prometheus-server.istio-system.svc.cluster.local:9090` (o mesmo que a KEDA e o canário usam):

```promql
# requisições por segundo no workload (é a métrica que escala a KEDA)
sum(rate(istio_requests_total{destination_workload="go"}[1m]))
# taxa de 5xx por serviço (go = estável, go-canary = canário em produção)
sum(rate(istio_requests_total{destination_service_name="go",response_code=~"5.."}[5m]))
  / sum(rate(istio_requests_total{destination_service_name="go"}[5m]))
```

### Traces (SigNoz)

- OTLP/gRPC para `signoz-otel-collector.signoz.svc.cluster.local:4317` (detalhes em
  [`observability.md`](observability.md)). Instrumentados: Gin, GORM (sem valores de parâmetros) e Redis.
- Atenção ao nome do serviço, que difere por ambiente (`TRACING_SERVICE_NAME`):
  - staging: `service.name = app-go-api`
  - produção: `service.name = go`
- Consultas úteis no Traces Explorer: filtre por `service.name`, `hasError = true` e ordene por
  duração; para banco, `db.system = postgresql`; para Redis, `db.system = redis`.
- O chart `k8s-infra` do SigNoz coleta métricas e eventos do cluster, filtráveis por
  `k8s.namespace.name = go`.
- Não há dashboard nem alerta do SigNoz versionado para este serviço no `prefeitura-rio/infra`;
  também não há Grafana para ele. Se precisar de um, peça pela raia de infraestrutura.

### Logs

- Seletor: `-n go -l app=go` (contêiner `go`; o sidecar do Istio é `istio-proxy`).
- `kubectl -n go logs -l app=go -c go --since=15m --tail=200` (acrescente `--previous` num pod que reiniciou).
- Formato texto, sem JSON. Linha de acesso por requisição:
  `[GIN] <data> | <status> | <latência> | <ip> | <método> "<rota>"`. A rota sai como template
  (`/api/v1/enrollments/user/:cpf`) e a query string vem com valores `[omitido]` (sem dados pessoais).
- Prefixos para filtrar (`grep`):

| Prefixo | Componente |
|---|---|
| `[RMIClient] circuit breaker` | Mudanças de estado do circuito do RMI |
| `[SecretariaMiddleware]` / `[SecretariaResolver]` | Falha ao resolver secretarias (fail-open/fail-closed) |
| `[OrgaoSyncWorker]` | Sincronização de órgãos com o RMI |
| `[CitizenSyncWorker]` | Sincronização de dados de cidadãos com o RMI |
| `[EmailWorker]` / `[EmailNotificationService]` | Fila e envio de e-mails |
| `[ClassReminderWorker]` | Lembrete D-1 de aula |
| `[Router]` | Workers que subiram ou foram desligados na inicialização |
| `Job <id> failed` / `Erro ao processar job` | Importação de inscrições em massa |

- Início do pod: `Executando migrações...` / `Migrações concluídas.` (Goose), `Servidor iniciado em`,
  e, com `APP_DEBUG=true`, o bloco `=== Configurações Carregadas ===` com `Ambiente:` e
  `RBAC Enabled:`.

### Estado do deploy

```bash
kubectl -n argocd get application go -o jsonpath='{.status.sync.status} {.status.health.status}{"\n"}'
kubectl -n argocd get application go -o jsonpath='{.spec.source.kustomize.images}{"\n"}'  # imagem em uso
kubectl -n go get rollout go                       # fase e réplicas
kubectl argo rollouts get rollout go -n go         # com o plugin: revisões, passos, análise
kubectl -n go get pods -l app=go
kubectl -n go get events --field-selector type=Warning --sort-by=.lastTimestamp
kubectl -n go get scaledobject go-scaled-object
kubectl -n go get analysisrun --sort-by=.metadata.creationTimestamp   # produção (canário)
gh run list --workflow deploy-staging.yaml --limit 5
gh run list --workflow release.yaml --limit 5
```

## 3. Alertas: o que significam e o primeiro passo

Os sinais que existem hoje (versionados neste repositório ou no `prefeitura-rio/infra`):

| Alerta / sinal | Onde chega | Significa | Primeiro passo |
|---|---|---|---|
| Gatus **"GO API"** (a cada 5 min: status 200, resposta < 1 s e `status == ok` em `/go/health`) | Discord (Gatus) | O processo não responde pela entrada pública, ou está lento. Como `/health` é raso, é problema de pod, Istio ou gateway, não de banco | `kubectl -n go get pods -l app=go` e eventos `Warning`; se os pods estão `Ready`, olhe o VirtualService `go-virtual-service` e o gateway; se não, os logs `--previous` |
| **"❌ Staging Deployment Failed"** | Discord (`deploy-staging.yaml`) | O smoke test (`/health` no `go-preview`) falhou ou o rollout não promoveu; o workflow abortou o rollout e o serviço ativo segue na versão anterior | `gh run view <run> --log-failed`; logs dos pods novos (`--previous`); erro comum: migration que falhou no `goose up` do entrypoint |
| **"🛑 Production Release Aborted"** | Discord (`release.yaml`) | A análise `go-success-rate` reprovou o canário (sucesso < 95% ou 5xx > 5%, até 5 falhas); o Argo Rollouts voltou o tráfego para a versão estável | Confirme `kubectl -n go get rollout go` (fase `Degraded`/abortado) e o `AnalysisRun` mais recente; leia os logs dos pods canário; corrija e publique nova release |
| **"❌ Production Release Failed"** (`Degraded`) ou **"⚠️ Unknown"** | Discord (`release.yaml`) | O rollout não ficou saudável no prazo do workflow (45 min) ou o estado não pôde ser lido | `kubectl argo rollouts get rollout go -n go`; se o canário está preso, siga a seção 4.1 |
| Circuito do RMI aberto (`rmi_client_circuit_breaker_state == 2`, log `[RMIClient] circuit breaker "api": closed -> open`) | Só métrica/log (sem alerta configurado) | 5 falhas seguidas no RMI (padrão `RMI_CIRCUIT_BREAKER_FAILURE_THRESHOLD`); chamadas são recusadas por 30 s antes de testar de novo | Seção 4.5 |
| KEDA em fallback (`ScaledObject` com condição `Fallback`) | Só no recurso | A consulta ao Prometheus falhou 3 vezes; a escala deixa de seguir a carga e vai para o `fallback.replicas` (limitado pelo mínimo do HPA) | Verifique o `prometheus-server` do `istio-system`; não escale à mão (seção 4.2) |
| Fila morta de e-mail crescendo (`email:queue:failed`) | Só no Redis/log | E-mails que esgotaram as 5 tentativas ou causaram panic | Seção 4.6 |

Não existem alertas do SigNoz nem do Prometheus/Alertmanager para este serviço. Se um alerta novo
for criado, acrescente-o nesta tabela.

## 4. Rotina de incidente

1. **Delimite:** staging ou produção? Desde quando? Coincide com um deploy
   (`gh run list --workflow deploy-staging.yaml` / `release.yaml`) ou com mudança no Infisical
   (o Rollout tem `secrets.infisical.com/auto-reload: "true"`: alterar `go-secrets` reinicia os pods)?
2. **Olhe na ordem:** pods e eventos → logs do contêiner `go` → taxa de 5xx no Istio → traces com
   erro no SigNoz → dependências (seção 1).
3. **Se começou com um deploy, volte a versão primeiro** (seção 4.1) e investigue depois.
4. **Registre** o que fez (comandos, horários, versão) na issue do incidente, sem dados pessoais.

### 4.1 Voltar uma versão (rollback)

Ninguém altera o banco num rollback: as migrations são cumulativas e o código anterior precisa
continuar compatível com o schema novo (expand/contract). Veja a seção 4.3 antes de qualquer
`goose down`.

**Produção (canário com Argo Rollouts):**

- *Durante o canário* (passos `10% → 2m → 20% → 3m → 40% → 5m → 60% → 5m → 80% → 5m → 100%`, análise
  a partir do passo 2): se a análise reprovar, o próprio Argo Rollouts aborta e devolve 100% do
  tráfego ao estável. Para abortar à mão: `kubectl argo rollouts abort go -n go`.
- *Depois de concluído* (100%): o caminho seguro é **republicar a GitHub Release anterior**, que
  refaz a imagem daquela tag, grava `:stable` e roda um canário novo com o código antigo:

  ```bash
  gh release list --limit 5                          # versões recentes
  gh run list --workflow release.yaml --limit 10     # ache o run da release anterior
  gh run rerun <run-id-da-release-anterior>          # reexecuta o deploy daquela tag
  gh run watch <run-id-da-release-anterior>
  ```

  O GitHub só reexecuta runs de até 30 dias. Para uma versão mais antiga, publique uma release nova
  a partir do commit da tag anterior (por exemplo `gh release create v<X.Y.Z+1> --target <sha-da-tag-anterior>`),
  o que dispara o mesmo `release.yaml` com aquele código.

  Alternativa mais rápida, se a imagem anterior estiver identificada pelo digest: aplicar no
  Application `go` o mesmo patch que o workflow faz, com o digest anterior
  (`kubectl patch application go -n argocd --type=json -p='[{"op":"replace","path":"/spec/source/kustomize/images/0","value":"ghcr.io/prefeitura-rio/app-go-api:stable@sha256:<digest-anterior>"}]'`).
- **Não use `kubectl argo rollouts undo` sozinho:** o Application `go` tem `selfHeal: true` e volta
  o Rollout para a imagem do override; mude o override (release anterior ou patch acima).
- Este serviço usa **Argo Rollouts**, não Flagger.

**Staging (blue-green):**

- Se o deploy falhou, o workflow já abortou e o serviço ativo continua na versão anterior.
- Para desfazer uma mudança já promovida, reverta o commit na `main` por PR
  (`git revert <sha>`); o push na `main` dispara `deploy-staging.yaml` de novo.
- O preview fica no `Service/go-preview` até a promoção (`autoPromotionEnabled: false`; quem
  promove é o workflow depois do smoke test).

### 4.2 Escalar e reiniciar com segurança

- **Réplicas são da KEDA** (`ScaledObject/go-scaled-object`: mínimo 3, máximo 20, escala por
  `istio_requests_total` acima de 100 req/s; ativação em 65). Não use `kubectl scale`: a KEDA
  desfaz. Para mudar mínimo/máximo, abra PR em `k8s/<env>/resources.yaml`.
- **Reiniciar os pods** (por exemplo, depois de trocar um valor no Infisical sem auto-reload):
  `kubectl argo rollouts restart go -n go`. Os pods são trocados aos poucos e o
  `PodDisruptionBudget` mantém ao menos 1 disponível.
- Antes de reiniciar, confira se há importação em massa rodando (seção 4.4): o processamento vive
  na memória do pod e é perdido.
- Cada pod que sobe roda `goose up` (seção 4.3) antes do servidor. Pod preso no início com erro de
  migration não fica `Ready`; não reinicie em laço, leia o log.
- No encerramento (SIGTERM), o servidor para de aceitar conexões e espera as requisições em curso
  por até `SERVER_REQUEST_TIMEOUT` (padrão 30 s); os workers são cancelados.

### 4.3 Migrations (Goose)

- Arquivos SQL em `internal/db/migrations/` (`YYYYMMDDHHMMSS_nome.sql`, `-- +goose Up/Down`).
- Em staging e produção, o `scripts/docker-entrypoint.sh` roda
  `goose -dir internal/db/migrations postgres "<dsn>" up` a cada início de pod quando
  `RUN_MIGRATIONS=true` (padrão da imagem). Com a mesma variável, o processo também roda o
  `AutoMigrate` do GORM para as tabelas básicas.
- Migration que falha derruba o pod (`set -e`): ele não fica `Ready`, o preview de staging não passa
  no smoke test e o canário de produção não avança. A versão anterior continua servindo.
- Diagnóstico (só leitura): `goose -dir internal/db/migrations postgres "$DATABASE_URL" status`
  com uma conexão de leitura; a tabela de controle é `goose_db_version`.
- **Reverter** (`just migrate-down` / `goose ... down`) desfaz só a última migration e pode apagar
  dados: é decisão de uma pessoa responsável pelo banco, com backup do Cloud SQL conferido. O caminho
  normal é corrigir para a frente com uma migration nova.

### 4.4 Importação de inscrições em massa (jobs)

- Entrada: `POST /api/v1/courses/{courseId}/enrollments/import` (CSV ou XLSX até 10 MB) → `202` com
  `job_id`. Acompanhamento: `GET /api/v1/jobs/{jobId}/status`.
- O arquivo é salvo num temporário **do próprio pod** e processado numa goroutine **do mesmo pod**.
  Status na tabela `jobs`: `pending` → `processing` → `completed` | `failed`, com `progress`,
  `success_count`, `error_count`, erros por linha em `errors` e o resumo em `result`.
- Se o pod morrer no meio (deploy, reinício, OOM, escala para baixo), o arquivo se perde e não há
  retomada. O job em execução renova `updated_at` a cada `JOB_HEARTBEAT_INTERVAL` (1 min); a varredura
  de jobs interrompidos (`[JobReaper]` nos logs, na inicialização de cada pod e a cada
  `JOB_REAPER_INTERVAL`, 2 min) marca como `failed` os jobs `pending`/`processing` sem renovação há mais
  de `JOB_STALE_AFTER` (10 min), com `completed_at` e o motivo "Processamento interrompido…" em `errors`
  (`line` 0).
- **Reprocessar:** peça a quem importou para reenviar o mesmo arquivo. As linhas já importadas voltam
  como erro "CPF já inscrito neste curso" (contadas em `duplicate_count`) e não duplicam inscrição.
- Para achar jobs presos que a varredura ainda não encerrou (consulta de leitura, sem dados pessoais):

  ```sql
  SELECT id, status, progress, total_records, success_count, error_count, updated_at
  FROM jobs WHERE status = 'processing' AND updated_at < now() - interval '30 minutes';
  ```

  Marcar um desses como `failed` é escrita em produção: só com decisão de uma pessoa responsável.
- Os erros por linha guardam a identificação da linha sem dados pessoais; não exporte a tabela.

### 4.5 RMI fora do ar (fail-open)

- O cliente RMI tenta de novo até 2 vezes em erro de rede e 502/503/504 (backoff de 100 ms a 1 s;
  timeout não é repetido) e abre o circuito depois de `RMI_CIRCUIT_BREAKER_FAILURE_THRESHOLD`
  falhas seguidas (padrão 5), por `RMI_CIRCUIT_BREAKER_OPEN_TIMEOUT` (padrão 30 s).
- O resolvedor de secretarias (CPF → órgãos, só com `RBAC_ENABLED=true` e Keycloak configurado) usa
  cache em memória por `RMI_SECRETARIA_CACHE_TTL` (padrão 2 min, limitado entre 1 e 5 min).
- Comportamento com o RMI fora:
  - **Leitura (GET/HEAD/OPTIONS) é fail-open:** segue sem escopo de secretaria; o cidadão continua
    vendo as listas públicas de cursos e vagas. Log `[SecretariaMiddleware] ... (fail-open)`.
  - **Escrita é fail-closed:** o escopo vira vazio e escrita restrita por secretaria é negada (403).
    Log `[SecretariaMiddleware] ... (fail-closed)`.
  - **Worker de cidadãos:** devolve o snapshot antigo quando o RMI falha
    (`[CitizenSyncWorker] RMI failed, returning stale snapshot ...`).
  - **Worker de órgãos:** os órgãos que falham ficam marcados como falha e mantêm o último snapshot;
    voltam a ser tentados depois de `ORGAO_SYNC_FAILED_RETRY_THRESHOLD` (padrão 1 h).
  - Consulta de CNPJ/CNAE (MEI, empresas) falha para o usuário.
- Primeiro passo: confirme o RMI (saúde e logs no namespace dele) e acompanhe
  `rmi_client_circuit_breaker_state`. Não reinicie o app-go-api por isso: ele se recupera sozinho
  quando o RMI volta (o circuito testa a cada 30 s).

### 4.6 Workers em segundo plano

| Worker | Liga/desliga | Intervalo | Concorrência entre pods |
|---|---|---|---|
| Sincronização de órgãos (`[OrgaoSyncWorker]`) | `ORGAO_SYNC_ENABLED` (padrão `true`) | `ORGAO_SYNC_INTERVAL` (15 min) | Lock no Redis `orgao_sync:lock` (TTL = intervalo); um pod por ciclo |
| Sincronização de cidadãos (`[CitizenSyncWorker]`) | `CITIZEN_SYNC_ENABLED` (padrão `true`) e Keycloak configurado | `CITIZEN_SYNC_INTERVAL` (15 min) | Sem lock: cada pod roda o seu ciclo |
| Fila de e-mails (`[EmailWorker]`) | Sempre ligado | Contínuo (2 consumidores por pod) | Fila no Redis |
| Lembrete D-1 de aula (`[ClassReminderWorker]`) | `CLASS_REMINDER_ENABLED` (padrão `true`) | `CLASS_REMINDER_INTERVAL` (1 h) | Lock `email:class_reminder:lock` e chave por envio `email:temporal:*` |

- Fila de e-mails no Redis: `email:queue` (pendentes), `email:queue:delayed` (retentativa agendada,
  até 5 tentativas) e `email:queue:failed` (fila morta). Para medir sem ler o conteúdo, que tem dados
  pessoais: `LLEN email:queue`, `ZCARD email:queue:delayed`, `LLEN email:queue:failed`.
- A fila morta **não é reprocessada sozinha**. Devolver itens para `email:queue` reenvia e-mails a
  cidadãos: só com decisão de uma pessoa, depois de corrigida a causa (Data Relay, template).
- Para desligar um worker, mude a variável no Infisical (`go-secrets`); os pods reiniciam sozinhos.

## 5. Feature flags e variáveis críticas

Lista completa e padrões em [`.env.example`](../.env.example) e `internal/config/config.go`. Onde ficam:
as variáveis de `env:` em `k8s/<env>/resources.yaml` (mudança por PR neste repositório) e o restante
no Secret `go-secrets`, sincronizado do Infisical pelo operador (alteração reinicia os pods).

| Variável | Padrão | Efeito operacional |
|---|---|---|
| `APP_ENV` | `development` | Comparado literalmente. **Staging usa `development` e produção usa `prod`** (não `staging`/`production`) |
| `APP_DEBUG` | `true` | Imprime o bloco de configuração no início (útil para conferir `Ambiente` e `RBAC Enabled`) |
| `RBAC_ENABLED` | `true` | Liga os middlewares de autorização de cursos e vagas e o resolvedor de secretarias. Desligado, eles viram no-op |
| `CERBOS_ENABLED` | `false` | Liga a autorização de propostas MEI pelo Cerbos (`CERBOS_ENDPOINT`, `CERBOS_TIMEOUT_SECONDS`) |
| `PROPOSTA_MEI_*_PERMISSIONS` | vazio | Ações do Cerbos que permitem ler, alterar e apagar propostas MEI de outra pessoa; vazias, ou com o Cerbos desligado, só o dono acessa |
| `RUN_MIGRATIONS` | `true` na imagem | Roda `goose up` e o `AutoMigrate` a cada início de pod |
| `ORGAO_SYNC_ENABLED` / `CITIZEN_SYNC_ENABLED` / `CLASS_REMINDER_ENABLED` | `true` | Ligam os workers da seção 4.6 |
| `HEIMDALL_BASE_URL` | vazio | Resolução de papéis. Vazio: os papéis vêm do JWT (só para desenvolvimento local) |
| `RMI_BASE_URL`, `RMI_SECRETARIA_CACHE_TTL`, `RMI_CIRCUIT_BREAKER_*` | ver seção 4.5 | Endereço e resiliência do cliente RMI |
| `KEYCLOAK_URL`, `KEYCLOAK_REALM`, `KEYCLOAK_CLIENT_ID`, `KEYCLOAK_CLIENT_SECRET` | vazio | Conta de serviço para chamar o RMI. Sem ela, o worker de cidadãos e o resolvedor de secretarias ficam desligados |
| `SERVER_REQUEST_TIMEOUT` | `30` (s) | Prazo por requisição e janela de drenagem no encerramento |
| `DB_MAX_OPEN_CONNS` / `DB_MAX_IDLE_CONNS` | `25` / `10` | Pool por pod (multiplique pelo número de réplicas, até 20, ao avaliar o limite do Cloud SQL) |
| `CACHE_*_TTL` | `30m`, `1h`, `5m`, `1h` | TTL dos caches de pessoa jurídica, dados de referência, cursos e contatos |
| `TRACING_ENABLED`, `TRACING_ENDPOINT`, `TRACING_SERVICE_NAME` | ver seção 2 | Traces no SigNoz (nome do serviço muda por ambiente) |
| `CORS_ALLOWED_ORIGINS` | em `k8s/<env>/resources.yaml` | Origens aceitas; front novo que recebe erro de CORS precisa entrar aqui por PR |
| `APP_TIMEZONE` | `America/Sao_Paulo` | Fuso do processo; muda a serialização das datas da API |

## 6. Referências

- Manifests: `k8s/staging/resources.yaml`, `k8s/prod/resources.yaml`
- Workflows: `.github/workflows/deploy-staging.yaml` (staging, push na `main`) e
  `.github/workflows/release.yaml` (produção, GitHub Release publicada)
- Observabilidade: [`docs/observability.md`](observability.md)
- Infraestrutura do namespace `go` (Redis, Typesense, Infisical, VirtualService, Gatus):
  `prefeitura-rio/infra`, projeto `superapp/` (`modules/k8s/go.tf`, `modules/k8s/services.tf`,
  `modules/k8s/yamls/gatus/values.yaml`, `modules/k8s/yamls/argocd/applications/go.yaml`)
- Convenções de desenvolvimento e ambiente: [`AGENTS.md`](../AGENTS.md) e [`README.md`](../README.md)
