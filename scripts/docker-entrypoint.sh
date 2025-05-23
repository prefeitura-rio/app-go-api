#!/bin/sh
set -e

echo "Iniciando aplicação..."

# Verificar se o arquivo .env existe
if [ -f "$(pwd)/.env" ]; then
    echo "Arquivo .env encontrado, carregando..."
    # Carregar variáveis do arquivo .env
    set -a
    . "$(pwd)/.env"
    set +a
    echo "Arquivo .env carregado com sucesso."
fi

# Verificar e definir variáveis de ambiente com valores padrão se estiverem ausentes
export DB_HOST="${DB_HOST:-db}"
export DB_PORT="${DB_PORT:-5432}"
export DB_USER="${DB_USER:-postgres}"
export DB_PASSWORD="${DB_PASSWORD:-postgres}"
export DB_NAME="${DB_NAME:-app_go_api}"
export DB_SSL_MODE="${DB_SSL_MODE:-disable}"
export DB_TIMEZONE="${DB_TIMEZONE:-UTC}"
export SERVER_HOST="${SERVER_HOST:-0.0.0.0}"
export SERVER_PORT="${SERVER_PORT:-8081}"
export APP_ENV="${APP_ENV:-development}"

# Imprimir valores para debug
echo "Variáveis configuradas:"
echo "DB_HOST=$DB_HOST"
echo "DB_PORT=$DB_PORT"
echo "DB_USER=$DB_USER"
echo "DB_NAME=$DB_NAME"
echo "DB_SSL_MODE=$DB_SSL_MODE"
echo "SERVER_HOST=$SERVER_HOST"
echo "SERVER_PORT=$SERVER_PORT"
echo "APP_ENV=$APP_ENV"

# Verificar e validar conexão com o banco de dados antes de prosseguir
if [ "$RUN_MIGRATIONS" = "true" ]; then
    echo "Executando migrações..."
    goose -dir internal/db/migrations postgres "user=${DB_USER} password=${DB_PASSWORD} dbname=${DB_NAME} host=${DB_HOST} port=${DB_PORT} sslmode=${DB_SSL_MODE:-disable}" up
    echo "Migrações concluídas."
fi

# Documentação Swagger já foi gerada durante o build
echo "Documentação Swagger já gerada durante o build..."

# Executar o comando fornecido (normalmente ./api)
echo "Executando servidor: $@"
exec "$@"