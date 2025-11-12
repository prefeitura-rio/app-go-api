#!/bin/bash

# Script temporário para inserir categorias via API
# Usa o port-forward ativo em localhost:8080

API_URL="http://localhost:8080/api/v1/categorias"

# Array com as 15 categorias solicitadas pelo Patrick
categorias=(
  "Tecnologia"
  "Marketing"
  "Finanças"
  "Gestão"
  "Games"
  "Nutrição"
  "Educação"
  "Gastronomia"
  "Saúde"
  "Veterinária"
  "Carreira"
  "Estética"
  "Sustentabilidade"
  "Artes"
  "Cultura",
  "Astronomia",
)

echo "========================================="
echo "Inserindo categorias via API..."
echo "========================================="
echo ""

success_count=0
error_count=0

for categoria in "${categorias[@]}"; do
  echo -n "Inserindo '$categoria'... "

  response=$(curl -s -X POST "$API_URL" \
    -H "Content-Type: application/json" \
    -d "{\"nome\": \"$categoria\"}" \
    -w "\n%{http_code}")

  http_code=$(echo "$response" | tail -n1)
  body=$(echo "$response" | sed '$d')

  if [ "$http_code" = "201" ] || [ "$http_code" = "200" ]; then
    id=$(echo "$body" | jq -r '.data.id // .id // "N/A"')
    echo "✅ OK (ID: $id)"
    ((success_count++))
  else
    echo "❌ ERRO (HTTP $http_code)"
    echo "   Resposta: $body"
    ((error_count++))
  fi
done

echo ""
echo "========================================="
echo "Resumo:"
echo "  ✅ Sucesso: $success_count"
echo "  ❌ Erros:   $error_count"
echo "========================================="
echo ""

# Listar todas as categorias criadas
echo "Categorias no sistema:"
curl -s "$API_URL?pageSize=20" | jq -r '.data[] | "\(.id) - \(.nome)"' | sort -n

exit 0
