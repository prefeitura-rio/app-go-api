# API Go - Plataforma de Cursos e Empregos

Uma API RESTful desenvolvida em Go para gerenciar cursos e oportunidades de emprego.

## Estrutura do Banco de Dados

```
erDiagram
    CURSO {
        int      id PK
        string   titulo
        string   descricao
        int      orgao_id FK
        int      instituicao_id FK
        enum     modalidade
        string   local_realizacao
        datetime data_inicio
        datetime data_termino
        datetime data_limite_inscricoes
        int      numero_vagas
        int      carga_horaria
        string   pre_requisitos
        boolean  certificacao_oferecida
        enum     status
        enum     turno
        enum     formato_aula
        string   link_inscricao
        string   contato_duvidas
    }

    EMPREGO {
        int      id PK
        string   titulo
        string   descricao
        int      orgao_id FK
        int      empresa_id FK
        enum     tipo_contratacao
        int      numero_vagas
        float    latitude
        float    longitude
        int      salario_min
        int      salario_max
        string   beneficios
        string   pre_requisitos
        datetime data_inicio_prevista
        datetime data_limite_candidatura
        enum     status
        int      escolaridade_id FK
        enum     jornada_trabalho
        enum     turno
        string   contato_duvidas
    }
```

## Requisitos

- Go 1.22+
- PostgreSQL 16+
- Redis 7+
- Docker e Docker Compose (opcional)
- Just command runner (`brew install just` ou `cargo install just`)

## Configuração e Instalação

### Setup Rápido

1. Clone o repositório:
```bash
git clone https://github.com/prefeitura-rio/app-go-api.git
cd app-go-api
```

2. Configure o ambiente de desenvolvimento:
```bash
just setup
```

3. Configure as variáveis de ambiente:
```bash
cp .env.example .env
# Edite o arquivo .env com suas configurações
```

4. Execute as migrações:
```bash
just migrate-up
```

5. Execute a aplicação:
```bash
just dev  # Com hot reload
# ou
just run  # Sem hot reload
```

### Com Docker

```bash
docker-compose up -d
```

## Estrutura do Projeto

```
app-go-api
  cmd/
    server/           # Ponto de entrada da aplicação
  docs/               # Documentação
  internal/
    config/           # Configurações da aplicação
    db/
      migrations/     # Migrações do banco de dados
    handlers/
      v1/             # Handlers da API v1
    middlewares/      # Middlewares da aplicação
    models/           # Modelos de dados
    repository/       # Camada de acesso a dados
    router/           # Definição de rotas
    services/         # Camada de serviços
  migrations/         # Scripts de migração
  scripts/            # Scripts utilitários
```

## Endpoints da API

### Cursos

- `GET /api/v1/cursos` - Listar cursos
- `POST /api/v1/cursos` - Criar curso
- `GET /api/v1/cursos/:id` - Obter curso por ID
- `PUT /api/v1/cursos/:id` - Atualizar curso
- `DELETE /api/v1/cursos/:id` - Excluir curso

### Empregos

- `GET /api/v1/empregos` - Listar empregos
- `POST /api/v1/empregos` - Criar emprego
- `GET /api/v1/empregos/:id` - Obter emprego por ID
- `PUT /api/v1/empregos/:id` - Atualizar emprego
- `DELETE /api/v1/empregos/:id` - Excluir emprego

## Desenvolvimento

### Comandos Disponíveis

```bash
# Ver todos os comandos disponíveis
just --list

# Configurar ambiente de desenvolvimento
just setup

# Qualidade de código
just fmt           # Formatar código
just lint          # Verificar linting
just lint-fix      # Corrigir problemas de linting automaticamente
just ci            # Executar todas as verificações (fmt + lint + test)

# Testes
just test                    # Executar todos os testes
just test-coverage           # Gerar relatório de cobertura
just test-pkg <package>      # Testar pacote específico
just test-integration        # Executar testes de integração
just test-e2e [url]          # Executar testes E2E (padrão: http://localhost:8080)

# Build e execução
just build         # Compilar binário
just run           # Executar aplicação
just dev           # Executar com hot reload

# Docker
just docker-build           # Build da imagem Docker
just docker-run             # Executar container

# Database
just migrate-up             # Aplicar migrações
just migrate-down           # Reverter migrações
just migrate-create <name>  # Criar nova migração

# Utilitários
just clean         # Limpar artefatos de build
just deps-update   # Atualizar dependências
just security-scan # Scan de segurança
```

### Workflow de Desenvolvimento

1. Crie uma feature branch a partir de `staging`:
```bash
git checkout staging
git pull
git checkout -b feat/minha-feature
```

2. Faça suas alterações e verifique a qualidade:
```bash
just fmt
just lint
just test
```

3. Ou execute todas as verificações de uma vez:
```bash
just ci
```

4. Commit e push:
```bash
git add .
git commit -m "feat: minha nova feature"
git push -u origin feat/minha-feature
```

5. Crie um Pull Request para `staging`

## CI/CD Pipeline

### Branch Strategy

```
feature → staging → main → production
```

### Pull Request Workflow

Todos os PRs para `staging` passam por quality gates automáticos:

1. **Linting** - golangci-lint com timeout de 10 minutos
2. **Dockerfile Lint** - Hadolint para validar Dockerfile
3. **Security Scan** - Trivy para vulnerabilidades CRITICAL/HIGH
4. **Unit Tests** - Testes com race detection
5. **Coverage Check** - Cobertura não pode diminuir >0.1%

**Requisitos para merge:**
- ✅ Todos os checks devem passar
- ✅ Código revisado e aprovado
- ✅ Branch atualizado com staging

### Deployment Process

#### Staging Environment
- **Trigger**: Push para branch `staging`
- **Workflow**: `build-container.yaml`
- **Processo**:
  1. Build da imagem Docker
  2. Push para GHCR com tag `latest`
  3. Gera documentação OpenAPI v3

#### Staging Deployment (Blue-Green)
- **Trigger**: Push para branch `main`
- **Workflow**: `deploy-staging.yaml`
- **Processo**:
  1. Build e push da imagem com digest
  2. Update do ArgoCD Application
  3. Blue-green deployment com preview
  4. Smoke tests (health + swagger)
  5. Auto-promote se tests passarem
  6. Auto-rollback se falharem
  7. Notificação no Discord

#### Production Deployment (Canary)
- **Trigger**: GitHub Release
- **Workflow**: `release.yaml`
- **Processo**:
  1. Build da imagem com versão (v1.2.3)
  2. Update do ArgoCD Application
  3. Canary deployment progressivo
  4. Análise de métricas automática
  5. Auto-rollback se falhar
  6. Notificação no Discord

### Quality Standards

- ✅ Linting deve passar sem warnings
- ✅ Todos os testes devem passar com race detection
- ✅ Cobertura de código não pode diminuir
- ✅ Sem vulnerabilidades CRITICAL ou HIGH
- ✅ Dockerfile deve seguir best practices

### Troubleshooting CI/CD

#### Linting Failures
```bash
just lint-fix  # Corrigir automaticamente
just lint      # Verificar issues restantes
```

#### Coverage Decrease
```bash
just test-coverage  # Gerar relatório HTML
# Abrir coverage/coverage.html
# Adicionar testes para código não coberto
```

#### Security Vulnerabilities
```bash
just security-scan          # Executar Trivy localmente
just deps-update            # Atualizar dependências
```

#### Test Failures
```bash
just test           # Executar testes localmente
just test-pkg <pkg> # Testar pacote específico
# Verificar logs de erro
# Corrigir race conditions (testes usam -race flag)
```

## Licença

Este projeto está licenciado sob a [Licença MIT](LICENSE).

<!-- Fix #80 -->
