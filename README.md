# go-rio-api

API Go para a Prefeitura do Rio de Janeiro.

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

### Hot Reload (Desenvolvimento)

Para desenvolvimento com hot reload, primeiro instale o Air:

```bash
go install github.com/air-verse/air@latest
```

Em seguida, inicie o servidor com:

```bash
make dev
```

Ou diretamente:

```bash
air
```

O Air irá monitorar alterações nos arquivos e recompilar automaticamente quando detectar mudanças.

### Docker

Para executar com Docker:

```bash
docker-compose up -d
```