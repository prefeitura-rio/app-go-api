package docs_test

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// O docs/openapi-v3.json é gerado por `just openapi` a partir do
// docs/swagger.json e é o arquivo que o Heimdall lê para mapear as ações de
// cada rota. PR que regenera só o Swagger 2 deixa o OpenAPI 3 para trás (uma
// rota que falta nele fica sem mapeamento), então os dois precisam descrever
// as mesmas operações e os mesmos schemas.

var httpMethods = map[string]bool{
	"get": true, "put": true, "post": true, "delete": true,
	"options": true, "head": true, "patch": true,
}

func readJSONFile(t *testing.T, name string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(name)
	require.NoError(t, err)
	var spec map[string]any
	require.NoError(t, json.Unmarshal(raw, &spec), name)
	return spec
}

func objectOf(v any) map[string]any {
	m, _ := v.(map[string]any)
	return m
}

func stringOf(v any) string {
	s, _ := v.(string)
	return s
}

// operationSummary é o que o Swagger 2 e o OpenAPI 3 descrevem igual numa
// operação: textos, tags, parâmetros fora do corpo e respostas.
func operationSummary(op map[string]any) map[string]string {
	out := map[string]string{
		"summary":     stringOf(op["summary"]),
		"description": stringOf(op["description"]),
	}
	if tags, ok := op["tags"].([]any); ok {
		names := make([]string, 0, len(tags))
		for _, tag := range tags {
			names = append(names, stringOf(tag))
		}
		out["tags"] = strings.Join(names, ",")
	}
	params, _ := op["parameters"].([]any)
	for _, raw := range params {
		p := objectOf(raw)
		in := stringOf(p["in"])
		if in == "body" || in == "formData" {
			// viram requestBody no OpenAPI 3
			continue
		}
		out["param "+in+" "+stringOf(p["name"])] = stringOf(p["description"])
	}
	for code, raw := range objectOf(op["responses"]) {
		out["response "+code] = stringOf(objectOf(raw)["description"])
	}
	return out
}

func operations(spec map[string]any) map[string]map[string]string {
	ops := map[string]map[string]string{}
	for path, rawItem := range objectOf(spec["paths"]) {
		for method, rawOp := range objectOf(rawItem) {
			if !httpMethods[method] {
				continue
			}
			ops[strings.ToUpper(method)+" "+path] = operationSummary(objectOf(rawOp))
		}
	}
	return ops
}

// schemaSummary é a descrição do schema e, por propriedade, a descrição dela.
func schemaSummary(schema map[string]any) map[string]string {
	out := map[string]string{"description": stringOf(schema["description"])}
	for name, raw := range objectOf(schema["properties"]) {
		out["property "+name] = stringOf(objectOf(raw)["description"])
	}
	if enum, ok := schema["enum"].([]any); ok {
		out["enum"] = fmt.Sprint(enum)
	}
	return out
}

func schemas(defs map[string]any) map[string]map[string]string {
	out := map[string]map[string]string{}
	for name, raw := range defs {
		out[name] = schemaSummary(objectOf(raw))
	}
	return out
}

// diffSummaries lista o que difere entre dois conjuntos de resumos: item que
// só existe de um lado e chave cujo valor muda.
func diffSummaries(kind string, swagger, openapi map[string]map[string]string) []string {
	var diffs []string
	for name, want := range swagger {
		got, ok := openapi[name]
		if !ok {
			diffs = append(diffs, fmt.Sprintf("%s %s ausente no openapi-v3.json", kind, name))
			continue
		}
		for key, value := range want {
			gotValue, has := got[key]
			switch {
			case !has:
				diffs = append(diffs, fmt.Sprintf("%s %s: %s ausente no openapi-v3.json", kind, name, key))
			case gotValue != value:
				diffs = append(diffs, fmt.Sprintf("%s %s: %s diverge do swagger.json", kind, name, key))
			}
		}
		for key := range got {
			if _, has := want[key]; !has {
				diffs = append(diffs, fmt.Sprintf("%s %s: %s só existe no openapi-v3.json", kind, name, key))
			}
		}
	}
	for name := range openapi {
		if _, ok := swagger[name]; !ok {
			diffs = append(diffs, fmt.Sprintf("%s %s só existe no openapi-v3.json", kind, name))
		}
	}
	sort.Strings(diffs)
	return diffs
}

func openAPIDivergences(swagger, openapi map[string]any) []string {
	diffs := diffSummaries("operação", operations(swagger), operations(openapi))
	diffs = append(diffs, diffSummaries("schema",
		schemas(objectOf(swagger["definitions"])),
		schemas(objectOf(objectOf(openapi["components"])["schemas"])))...)
	return diffs
}

func TestOpenAPIV3_InSyncWithSwaggerJSON(t *testing.T) {
	swagger := readJSONFile(t, "swagger.json")
	openapi := readJSONFile(t, "openapi-v3.json")

	require.True(t, strings.HasPrefix(stringOf(openapi["openapi"]), "3."), "openapi-v3.json não é OpenAPI 3")
	require.NotEmpty(t, operations(swagger), "swagger.json sem operações")

	diffs := openAPIDivergences(swagger, openapi)
	assert.Empty(t, diffs, "docs/openapi-v3.json desatualizado em relação ao docs/swagger.json: rode `just openapi` e commite o resultado")
}

// As operações do OpenAPI 3 publicado exigem o bearerAuth que o Istio cobra,
// como o pós-processamento de `just openapi` grava.
func TestOpenAPIV3_OperationsRequireBearerAuth(t *testing.T) {
	openapi := readJSONFile(t, "openapi-v3.json")

	scheme := objectOf(objectOf(objectOf(openapi["components"])["securitySchemes"])["bearerAuth"])
	assert.Equal(t, "http", scheme["type"])
	assert.Equal(t, "bearer", scheme["scheme"])

	var missing []string
	for path, rawItem := range objectOf(openapi["paths"]) {
		for method, rawOp := range objectOf(rawItem) {
			if !httpMethods[method] {
				continue
			}
			security, _ := objectOf(rawOp)["security"].([]any)
			if len(security) != 1 || objectOf(security[0])["bearerAuth"] == nil {
				missing = append(missing, strings.ToUpper(method)+" "+path)
			}
		}
	}
	sort.Strings(missing)
	assert.Empty(t, missing, "operações sem security bearerAuth: rode `just openapi`")
}

// O comparador acusa cada tipo de divergência que já aconteceu: rota nova,
// schema novo, campo novo e descrição alterada.
func TestOpenAPIDivergences_DetectsDrift(t *testing.T) {
	swagger := map[string]any{
		"paths": map[string]any{
			"/a": map[string]any{"get": map[string]any{
				"description": "nova",
				"parameters": []any{
					map[string]any{"in": "path", "name": "cpf"},
					map[string]any{"in": "body", "name": "body"},
				},
				"responses": map[string]any{"200": map[string]any{"description": "OK"}},
			}},
			"/b": map[string]any{"put": map[string]any{}},
		},
		"definitions": map[string]any{
			"Vaga": map[string]any{"properties": map[string]any{
				"titulo":     map[string]any{"type": "string"},
				"tipo_etapa": map[string]any{"type": "string"},
			}},
			"Nova": map[string]any{},
		},
	}
	openapi := map[string]any{
		"paths": map[string]any{
			"/a": map[string]any{"get": map[string]any{
				"description": "antiga",
				"parameters":  []any{map[string]any{"in": "path", "name": "cpf"}},
				"responses":   map[string]any{"200": map[string]any{"description": "OK"}},
			}},
		},
		"components": map[string]any{"schemas": map[string]any{
			"Vaga": map[string]any{"properties": map[string]any{
				"titulo": map[string]any{"type": "string"},
			}},
		}},
	}

	assert.Equal(t, []string{
		"operação GET /a: description diverge do swagger.json",
		"operação PUT /b ausente no openapi-v3.json",
		"schema Nova ausente no openapi-v3.json",
		"schema Vaga: property tipo_etapa ausente no openapi-v3.json",
	}, openAPIDivergences(swagger, openapi))

	assert.Empty(t, openAPIDivergences(swagger, swagger2openapiLike(swagger)))
}

// swagger2openapiLike faz a parte da conversão que o comparador enxerga:
// definitions viram components.schemas e o parâmetro body sai da lista.
func swagger2openapiLike(swagger map[string]any) map[string]any {
	paths := map[string]any{}
	for path, rawItem := range objectOf(swagger["paths"]) {
		item := map[string]any{}
		for method, rawOp := range objectOf(rawItem) {
			op := map[string]any{}
			for k, v := range objectOf(rawOp) {
				op[k] = v
			}
			if params, ok := op["parameters"].([]any); ok {
				var kept []any
				for _, p := range params {
					if stringOf(objectOf(p)["in"]) != "body" {
						kept = append(kept, p)
					}
				}
				op["parameters"] = kept
			}
			item[method] = op
		}
		paths[path] = item
	}
	return map[string]any{
		"paths":      paths,
		"components": map[string]any{"schemas": swagger["definitions"]},
	}
}
