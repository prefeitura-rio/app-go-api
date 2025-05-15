#!/bin/bash
set -e

echo "Parando contêineres existentes..."
docker-compose down

echo "Reconstruindo e iniciando contêineres..."
docker-compose up -d --build

echo "Aguardando inicialização dos serviços..."
sleep 5

echo "Logs do serviço API:"
docker-compose logs api 