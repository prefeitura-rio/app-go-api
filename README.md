# go-rio-api

API Go para a Prefeitura do Rio de Janeiro.

## Configuração do Ambiente

A aplicação utiliza o [Viper](https://github.com/spf13/viper) para gerenciar configurações através de variáveis de ambiente. Você pode definir essas variáveis diretamente no sistema ou criar um arquivo `.env` na raiz do projeto.

### Variáveis de Ambiente

Copie o exemplo abaixo para criar seu arquivo `.env`:

```
# Configurações do Servidor
APP_ENV=development
PORT=8080
DEBUG=true
API_PREFIX=/api/v1
LOG_LEVEL=debug

# Configurações do Banco de Dados
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=app_db
DB_SSL_MODE=disable
DB_TIMEZONE=UTC

# Configurações de JWT
JWT_SECRET=your_jwt_secret_key
JWT_EXPIRATION=24h
```

### Executando a Aplicação

Para executar a aplicação:

```bash
go run cmd/server/main.go
```

Ou construa e execute o binário:

```bash
go build -o app cmd/server/main.go
./app
```

### Docker

Para executar com Docker:

```bash
docker-compose up -d
```