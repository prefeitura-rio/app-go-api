package docs_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/prefeitura-rio/app-go-api/docs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// specDoc abstrai a diferença entre o Swagger 2.0 (definitions, schema direto
// na resposta) e o OpenAPI 3 publicado (components.schemas, content por media type).
type specDoc struct {
	raw     map[string]any
	openapi bool
}

func loadSwaggerJSON(t *testing.T) specDoc {
	t.Helper()
	raw, err := os.ReadFile("swagger.json")
	require.NoError(t, err)
	var spec map[string]any
	require.NoError(t, json.Unmarshal(raw, &spec))
	return specDoc{raw: spec}
}

func loadDocsGo(t *testing.T) specDoc {
	t.Helper()
	var spec map[string]any
	require.NoError(t, json.Unmarshal([]byte(docs.SwaggerInfo.ReadDoc()), &spec))
	return specDoc{raw: spec}
}

func loadOpenAPIV3(t *testing.T) specDoc {
	t.Helper()
	raw, err := os.ReadFile("openapi-v3.json")
	require.NoError(t, err)
	var spec map[string]any
	require.NoError(t, json.Unmarshal(raw, &spec))
	return specDoc{raw: spec, openapi: true}
}

func asMap(t *testing.T, v any, what string) map[string]any {
	t.Helper()
	m, ok := v.(map[string]any)
	require.True(t, ok, "%s: esperado objeto, veio %T", what, v)
	return m
}

// definition devolve o schema nomeado (models.Curso, empregabilidade.Vaga…).
func (s specDoc) definition(t *testing.T, name string) map[string]any {
	t.Helper()
	var defs any
	if s.openapi {
		defs = asMap(t, s.raw["components"], "components")["schemas"]
	} else {
		defs = s.raw["definitions"]
	}
	def, ok := asMap(t, defs, "definitions")[name]
	require.True(t, ok, "schema %s ausente", name)
	return asMap(t, def, name)
}

// resolve segue $ref e allOf de um único elemento (como o swag descreve um
// campo com descrição ou extensão) até o schema nomeado.
func (s specDoc) resolve(t *testing.T, schema map[string]any) map[string]any {
	t.Helper()
	if ref, ok := schema["$ref"].(string); ok {
		return s.definition(t, ref[strings.LastIndex(ref, "/")+1:])
	}
	if allOf, ok := schema["allOf"].([]any); ok && len(allOf) == 1 {
		return s.resolve(t, asMap(t, allOf[0], "allOf[0]"))
	}
	return schema
}

// refName devolve o nome do schema referenciado por $ref (direto ou via allOf).
func refName(t *testing.T, schema map[string]any) string {
	t.Helper()
	if ref, ok := schema["$ref"].(string); ok {
		return ref[strings.LastIndex(ref, "/")+1:]
	}
	if allOf, ok := schema["allOf"].([]any); ok && len(allOf) == 1 {
		return refName(t, asMap(t, allOf[0], "allOf[0]"))
	}
	t.Fatalf("schema sem $ref: %v", schema)
	return ""
}

// response200 devolve o schema da resposta 200 de uma operação GET.
func (s specDoc) response200(t *testing.T, path string) map[string]any {
	t.Helper()
	item := asMap(t, asMap(t, s.raw["paths"], "paths")[path], path)
	op := asMap(t, item["get"], path+" get")
	resp := asMap(t, asMap(t, op["responses"], "responses")["200"], path+" 200")
	if s.openapi {
		content := asMap(t, resp["content"], path+" content")
		return asMap(t, asMap(t, content["application/json"], "application/json")["schema"], "schema")
	}
	return asMap(t, resp["schema"], path+" schema")
}

func property(t *testing.T, schema map[string]any, name string) map[string]any {
	t.Helper()
	props := asMap(t, schema["properties"], "properties")
	p, ok := props[name]
	require.True(t, ok, "propriedade %s ausente", name)
	return asMap(t, p, name)
}

func isNullable(schema map[string]any) bool {
	return schema["x-nullable"] == true || schema["nullable"] == true
}

func specLoaders() map[string]func(t *testing.T) specDoc {
	return map[string]func(t *testing.T) specDoc{
		"docs.go (spec servido pela API)": loadDocsGo,
		"swagger.json":                    loadSwaggerJSON,
		"openapi-v3.json (publicado)":     loadOpenAPIV3,
	}
}

// Os valores que a API devolve em formato_aula e turno (inclusive o vazio, de
// cursos sem o campo) precisam estar no enum, senão a resposta real é inválida
// pelo próprio contrato.
func TestSpec_CursoEnums_IncludeRealValues(t *testing.T) {
	for name, load := range specLoaders() {
		t.Run(name, func(t *testing.T) {
			s := load(t)
			formato := s.definition(t, "models.FormatoAula")["enum"]
			assert.ElementsMatch(t, []any{"GRAVADO", "AO_VIVO", "PRESENCIAL", ""}, formato)

			turno := s.definition(t, "models.Turno")["enum"]
			assert.Contains(t, turno, "")
			assert.Contains(t, turno, "LIVRE")
		})
	}
}

// A resposta de curso (listagens e detalhe) traz os relacionamentos que
// transformCursoToResponse devolve, e os campos que a API envia como null são
// anuláveis.
func TestSpec_CursoResponse_DescribesRelationsAndNulls(t *testing.T) {
	for name, load := range specLoaders() {
		t.Run(name, func(t *testing.T) {
			s := load(t)
			curso := s.definition(t, "models.CursoResponse")

			assert.Equal(t, "models.Categoria", refName(t, asMap(t, property(t, curso, "categorias")["items"], "categorias.items")))
			assert.Equal(t, "models.Acessibilidade", refName(t, asMap(t, property(t, curso, "acessibilidades")["items"], "acessibilidades.items")))
			assert.Equal(t, "models.CustomField", refName(t, asMap(t, property(t, curso, "custom_fields")["items"], "custom_fields.items")))
			assert.Equal(t, "models.LocationClass", refName(t, asMap(t, property(t, curso, "locations")["items"], "locations.items")))
			assert.Equal(t, "models.InstituicaoEnsino", refName(t, property(t, curso, "instituicao")))
			assert.Equal(t, "models.RemoteClass", refName(t, property(t, curso, "remote_class")))
			assert.Equal(t, "models.FormatoAula", refName(t, property(t, curso, "formato_aula")))

			for _, field := range []string{
				"categorias", "acessibilidades", "custom_fields", "locations", "instituicao", "remote_class",
				"enrollment_start_date", "enrollment_end_date", "data_inicio", "data_termino",
				"data_limite_inscricoes", "instituicao_id", "is_external_partner", "is_visible",
				"auto_approve_enrollments",
			} {
				assert.True(t, isNullable(property(t, curso, field)), "%s deveria ser anulável", field)
			}
		})
	}
}

// As listagens públicas descrevem o envelope e o schema dos itens, e o detalhe
// do curso o envelope { success, data }.
func TestSpec_PublicListings_DescribeEnvelope(t *testing.T) {
	for name, load := range specLoaders() {
		t.Run(name, func(t *testing.T) {
			s := load(t)

			for _, path := range []string{"/api/public/courses", "/api/v1/courses"} {
				list := s.resolve(t, s.response200(t, path))
				assert.Equal(t, "boolean", property(t, list, "success")["type"], path)
				data := s.resolve(t, property(t, list, "data"))
				courses := property(t, data, "courses")
				assert.Equal(t, "array", courses["type"], path)
				assert.Equal(t, "models.CursoResponse", refName(t, asMap(t, courses["items"], "courses.items")), path)
				pagination := s.resolve(t, property(t, data, "pagination"))
				for _, f := range []string{"page", "limit", "total", "total_pages"} {
					assert.Equal(t, "integer", property(t, pagination, f)["type"], "%s pagination.%s", path, f)
				}
			}

			for _, path := range []string{"/api/public/courses/{courseId}", "/api/v1/courses/{courseId}"} {
				detail := s.resolve(t, s.response200(t, path))
				assert.Equal(t, "boolean", property(t, detail, "success")["type"], path)
				assert.Equal(t, "models.CursoResponse", refName(t, property(t, detail, "data")), path)
			}

			listings := map[string]string{
				"/api/public/empregabilidade/vagas": "empregabilidade.Vaga",
				"/api/public/oportunidades-mei":     "models.OportunidadeMEI",
			}
			for path, item := range listings {
				list := s.resolve(t, s.response200(t, path))
				data := property(t, list, "data")
				assert.Equal(t, "array", data["type"], path)
				assert.True(t, isNullable(data), "%s data deveria ser anulável (lista vazia sai como null)", path)
				assert.Equal(t, item, refName(t, asMap(t, data["items"], "data.items")), path)
				meta := s.resolve(t, property(t, list, "meta"))
				for _, f := range []string{"page", "page_size", "total"} {
					assert.Equal(t, "integer", property(t, meta, f)["type"], "%s meta.%s", path, f)
				}
			}
		})
	}
}
