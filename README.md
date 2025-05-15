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

- Go 1.18+
- PostgreSQL
- Docker e Docker Compose (opcional)

## Configuração e Instalação

1. Clone o repositório:
```bash
git clone https://github.com/prefeitura-rio/app-go-api.git
cd app-go-api
```

2. Configure as variáveis de ambiente:
```bash
cp .env.example .env
# Edite o arquivo .env com suas configurações
```

3. Execute com Docker:
```bash
docker-compose up -d
```

Ou execute localmente:
```bash
go mod tidy
go run cmd/server/main.go
```

4. Execute as migrações:
```bash
make migrate-up
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

## Licença

Este projeto está licenciado sob a [Licença MIT](LICENSE).