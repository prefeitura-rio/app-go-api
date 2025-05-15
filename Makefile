.PHONY: swagger run help migrate-up migrate-down

help: ## Exibe ajuda
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

swagger: ## Atualiza a documentação Swagger
	./scripts/update-swagger.sh

run: swagger ## Atualiza o Swagger e inicia o servidor
	go run cmd/server/main.go

dev: ## Inicia o servidor com hot reload
	air

build: ## Compila o projeto
	go build -o bin/server cmd/server/main.go

test: ## Executa os testes
	go test -v ./...

migrate-up: ## Executa as migrações do banco de dados
	goose -dir internal/db/migrations postgres "user=${DB_USER} password=${DB_PASSWORD} dbname=${DB_NAME} host=${DB_HOST} port=${DB_PORT} sslmode=disable" up

migrate-down: ## Reverte todas as migrações do banco de dados
	goose -dir internal/db/migrations postgres "user=${DB_USER} password=${DB_PASSWORD} dbname=${DB_NAME} host=${DB_HOST} port=${DB_PORT} sslmode=disable" down

clean: ## Remove binários e arquivos temporários
	rm -rf bin
	go clean 