package v1

import (
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// documentedJSONKeys lista as chaves JSON que o swag publica para um struct:
// campos de structs embutidos sem tag json entram achatados, e campos com
// swaggerignore:"true" ou json:"-" ficam de fora.
func documentedJSONKeys(t reflect.Type) []string {
	var keys []string
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Tag.Get("swaggerignore") == "true" {
			continue
		}
		tag := f.Tag.Get("json")
		name := strings.Split(tag, ",")[0]
		if f.Anonymous && name == "" {
			keys = append(keys, documentedJSONKeys(f.Type)...)
			continue
		}
		if name == "-" || !f.IsExported() {
			continue
		}
		if name == "" {
			name = f.Name
		}
		keys = append(keys, name)
	}
	return keys
}

// O schema publicado da resposta de curso (models.CursoResponse) precisa ter
// exatamente as chaves que transformCursoToResponse devolve nas listagens: um
// campo novo na resposta sem o campo no schema (ou o contrário) quebra os
// consumidores que validam contra o OpenAPI (catalogo, Orval).
func TestTransformCursoToResponse_KeysMatchCursoResponseSchema(t *testing.T) {
	raw, err := json.Marshal(transformCursoToResponse(&models.Curso{ID: 1}))
	require.NoError(t, err)
	var got map[string]any
	require.NoError(t, json.Unmarshal(raw, &got))

	responseKeys := make([]string, 0, len(got))
	for k := range got {
		responseKeys = append(responseKeys, k)
	}
	sort.Strings(responseKeys)

	schemaKeys := documentedJSONKeys(reflect.TypeOf(models.CursoResponse{}))
	sort.Strings(schemaKeys)

	assert.Equal(t, responseKeys, schemaKeys)
}

// O detalhe do curso (GET /courses/{id}) serializa models.Curso direto; as
// chaves que ele pode devolver (fora enrollments, que não é carregado) também
// são as do schema models.CursoResponse.
func TestCursoDetail_KeysMatchCursoResponseSchema(t *testing.T) {
	isTrue := true
	instituicaoID := 1
	curso := models.Curso{
		ID:                     1,
		OrgaoID:                "orgao",
		InstituicaoID:          &instituicaoID,
		FormacaoLink:           "https://exemplo.invalid",
		Facilitator:            "f",
		Objectives:             "o",
		ExpectedResults:        "e",
		ProgramContent:         "p",
		Methodology:            "m",
		ResourcesUsed:          "r",
		MaterialUsed:           "m",
		TeachingMaterial:       "t",
		IsExternalPartner:      &isTrue,
		ExternalPartnerName:    "n",
		ExternalPartnerURL:     "https://exemplo.invalid",
		ExternalPartnerLogoURL: "https://exemplo.invalid/logo.png",
		ExternalPartnerContact: "c",
		Accessibility:          "a",
		IsVisible:              &isTrue,
		AutoApproveEnrollments: &isTrue,
		Instituicao:            &models.InstituicaoEnsino{ID: 1, Nome: "i"},
		Categorias:             []models.Categoria{{ID: 1, Nome: "c"}},
		Acessibilidades:        []models.Acessibilidade{{ID: 1, Nome: "a"}},
		LocationClasses:        []models.LocationClass{{Address: "a"}},
	}
	raw, err := json.Marshal(curso)
	require.NoError(t, err)
	var got map[string]any
	require.NoError(t, json.Unmarshal(raw, &got))

	schemaKeys := map[string]bool{}
	for _, k := range documentedJSONKeys(reflect.TypeOf(models.CursoResponse{})) {
		schemaKeys[k] = true
	}
	for k := range got {
		assert.True(t, schemaKeys[k], "chave %q do detalhe do curso ausente de models.CursoResponse", k)
	}
}
