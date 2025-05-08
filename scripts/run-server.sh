#!/bin/bash

echo "Atualizando documentação Swagger..."

$HOME/go/bin/swag init -g cmd/server/main.go

if [ $? -eq 0 ]; then
    echo "Documentação atualizada com sucesso! Iniciando servidor..."
    go run cmd/server/main.go
else
    echo "Erro ao atualizar documentação Swagger. Verifique os erros acima."
    exit 1
fi 