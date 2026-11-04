package router

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prefeitura-rio/app-go-api/docs"
	v1 "github.com/prefeitura-rio/app-go-api/internal/handlers/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// O Heimdall cria o mapeamento endpoint→ação a partir do OpenAPI gerado pelo
// Swagger; rota registrada sem @Router no spec é negada pelo ext_authz para
// qualquer cliente (DEV-228). Toda rota de typesense precisa estar no spec.
func TestTypesenseRoutes_AllDocumentedInSwagger(t *testing.T) {
	engine := gin.New()
	registerTypesenseRoutes(engine.Group("/api/v1"), v1.NewTypesenseHandlerWithService(nil))

	var spec struct {
		Paths map[string]map[string]json.RawMessage `json:"paths"`
	}
	require.NoError(t, json.Unmarshal([]byte(docs.SwaggerInfo.ReadDoc()), &spec))

	routes := engine.Routes()
	require.Len(t, routes, 4, "rotas de typesense registradas")

	for _, route := range routes {
		path := swaggerPath(route.Path)
		method := strings.ToLower(route.Method)
		operations, ok := spec.Paths[path]
		if assert.Truef(t, ok, "%s %s não está no Swagger (docs/docs.go)", route.Method, path) {
			assert.Containsf(t, operations, method, "%s %s não está no Swagger (docs/docs.go)", route.Method, path)
		}
	}
}

func TestTypesenseRoutes_CursosEEmpregosNoSwagger(t *testing.T) {
	var spec struct {
		Paths map[string]map[string]json.RawMessage `json:"paths"`
	}
	require.NoError(t, json.Unmarshal([]byte(docs.SwaggerInfo.ReadDoc()), &spec))

	for _, path := range []string{"/api/v1/typesense/cursos/search", "/api/v1/typesense/empregos/search"} {
		operations, ok := spec.Paths[path]
		require.Truef(t, ok, "%s ausente do Swagger", path)
		require.Containsf(t, operations, "post", "POST %s ausente do Swagger", path)

		var op struct {
			Summary    string                     `json:"summary"`
			Tags       []string                   `json:"tags"`
			Parameters []map[string]any           `json:"parameters"`
			Responses  map[string]json.RawMessage `json:"responses"`
		}
		require.NoError(t, json.Unmarshal(operations["post"], &op))
		assert.NotEmpty(t, op.Summary, "%s sem @Summary", path)
		assert.Equal(t, []string{"search"}, op.Tags, "%s deve ficar na tag search", path)
		assert.Len(t, op.Parameters, 1, "%s deve documentar o corpo da busca", path)
		for _, code := range []string{"200", "400", "500"} {
			assert.Containsf(t, op.Responses, code, "%s sem resposta %s", path, code)
		}
	}
}

// swaggerPath converte "/a/:id" (Gin) em "/a/{id}" (Swagger).
func swaggerPath(ginPath string) string {
	segments := strings.Split(ginPath, "/")
	for i, segment := range segments {
		if strings.HasPrefix(segment, ":") {
			segments[i] = "{" + strings.TrimPrefix(segment, ":") + "}"
		}
	}
	return strings.Join(segments, "/")
}
