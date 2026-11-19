package docs_test

import (
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// httpMethods são as chaves de um path item que descrevem operações; o resto
// (parameters, summary…) não é operação.
var httpMethods = map[string]bool{
	"get": true, "put": true, "post": true, "delete": true,
	"options": true, "head": true, "patch": true, "trace": true,
}

// operations lista "MÉTODO path" de cada operação do spec.
func operations(t *testing.T, s specDoc) []string {
	t.Helper()
	var ops []string
	for path, rawItem := range asMap(t, s.raw["paths"], "paths") {
		for method := range asMap(t, rawItem, path) {
			if httpMethods[method] {
				ops = append(ops, strings.ToUpper(method)+" "+path)
			}
		}
	}
	sort.Strings(ops)
	return ops
}

// O Heimdall cria as ações de autorização a partir do openapi-v3.json publicado
// (scripts/manage_openapi_mappings.py): uma operação que está no swagger.json mas
// não no openapi-v3.json fica sem ação e o ext_authz a recusa com
// "no action mapping found for endpoint" (DEV-566). Regenere com `just openapi`.
func TestOpenAPIV3_CoversEverySwaggerOperation(t *testing.T) {
	swagger := operations(t, loadSwaggerJSON(t))
	openapi := operations(t, loadOpenAPIV3(t))

	inOpenAPI := make(map[string]bool, len(openapi))
	for _, op := range openapi {
		inOpenAPI[op] = true
	}
	inSwagger := make(map[string]bool, len(swagger))
	for _, op := range swagger {
		inSwagger[op] = true
	}

	var missing, extra []string
	for _, op := range swagger {
		if !inOpenAPI[op] {
			missing = append(missing, op)
		}
	}
	for _, op := range openapi {
		if !inSwagger[op] {
			extra = append(extra, op)
		}
	}

	assert.Empty(t, missing, "operações do swagger.json ausentes do openapi-v3.json (rode `just openapi`)")
	assert.Empty(t, extra, "operações do openapi-v3.json que não existem no swagger.json (rode `just openapi`)")
}

// Cada operação publicada exige bearerAuth: é o que o pós-processamento de
// `just openapi` (e do build-container.yaml) aplica.
func TestOpenAPIV3_EveryOperationDeclaresBearerAuth(t *testing.T) {
	s := loadOpenAPIV3(t)
	var without []string
	for path, rawItem := range asMap(t, s.raw["paths"], "paths") {
		for method, rawOp := range asMap(t, rawItem, path) {
			if !httpMethods[method] {
				continue
			}
			op := asMap(t, rawOp, method+" "+path)
			sec, _ := op["security"].([]any)
			found := false
			for _, raw := range sec {
				if req, ok := raw.(map[string]any); ok {
					if _, has := req["bearerAuth"]; has {
						found = true
					}
				}
			}
			if !found {
				without = append(without, strings.ToUpper(method)+" "+path)
			}
		}
	}
	sort.Strings(without)
	assert.Empty(t, without, "operações do openapi-v3.json sem bearerAuth (rode `just openapi`)")
}
