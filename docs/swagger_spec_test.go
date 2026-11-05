package docs_test

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"testing"

	"github.com/prefeitura-rio/app-go-api/docs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// nonBodyParamsWithExample lista os parâmetros de operação que não são `in: body`
// e têm `example`: o Swagger 2.0 não aceita `example` nesses parâmetros, e o
// spec inválido quebra geradores que validam o schema (o Orval do portal-interno).
func nonBodyParamsWithExample(t *testing.T, spec map[string]any) []string {
	t.Helper()

	paths, ok := spec["paths"].(map[string]any)
	require.True(t, ok, "spec sem paths")

	var found []string
	for path, rawItem := range paths {
		item, ok := rawItem.(map[string]any)
		require.True(t, ok, "path %s inválido", path)
		for method, rawOp := range item {
			var params []any
			if method == "parameters" {
				params, _ = rawOp.([]any)
			} else if op, ok := rawOp.(map[string]any); ok {
				params, _ = op["parameters"].([]any)
			}
			for _, rawParam := range params {
				param, ok := rawParam.(map[string]any)
				if !ok {
					continue
				}
				if _, has := param["example"]; has && param["in"] != "body" {
					found = append(found, fmt.Sprintf("%s %s %v (in: %v)", method, path, param["name"], param["in"]))
				}
			}
		}
	}
	sort.Strings(found)
	return found
}

func TestSwaggerSpec_NonBodyParams_HaveNoExample(t *testing.T) {
	loaders := map[string]func(t *testing.T) map[string]any{
		"docs.go (spec servido pela API)": func(t *testing.T) map[string]any {
			var spec map[string]any
			require.NoError(t, json.Unmarshal([]byte(docs.SwaggerInfo.ReadDoc()), &spec))
			return spec
		},
		"swagger.json": func(t *testing.T) map[string]any {
			raw, err := os.ReadFile("swagger.json")
			require.NoError(t, err)
			var spec map[string]any
			require.NoError(t, json.Unmarshal(raw, &spec))
			return spec
		},
		"swagger.yaml": func(t *testing.T) map[string]any {
			raw, err := os.ReadFile("swagger.yaml")
			require.NoError(t, err)
			var spec map[string]any
			require.NoError(t, yaml.Unmarshal(raw, &spec))
			return spec
		},
	}

	for name, load := range loaders {
		t.Run(name, func(t *testing.T) {
			spec := load(t)
			assert.Equal(t, "2.0", spec["swagger"])
			assert.Empty(t, nonBodyParamsWithExample(t, spec),
				"parâmetros não-body com example (inválido no Swagger 2.0; use a descrição)")
		})
	}
}
