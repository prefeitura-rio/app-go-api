package services_test

import (
	"testing"

	"github.com/prefeitura-rio/app-go-api/internal/models"
)

// Note: Full integration testing of TypesenseService requires mocking the Typesense client
// which is complex. These tests focus on the parameter conversion logic and data structures.

func TestCursoSearchParameters_ToSearchParameters(t *testing.T) {
	t.Run("Basic search with query", func(t *testing.T) {
		params := &models.CursoSearchParameters{
			Q:       "programação",
			Page:    1,
			PerPage: 10,
		}

		searchParams := params.ToSearchParameters()

		if searchParams.Q != "programação" {
			t.Errorf("expected query 'programação', got %s", searchParams.Q)
		}

		if searchParams.QueryBy != "record.titulo,record.descricao,record.pre_requisitos" {
			t.Errorf("unexpected queryBy: %s", searchParams.QueryBy)
		}

		if searchParams.Page != 1 {
			t.Errorf("expected page 1, got %d", searchParams.Page)
		}

		if searchParams.PerPage != 10 {
			t.Errorf("expected perPage 10, got %d", searchParams.PerPage)
		}

		if !searchParams.Prefix {
			t.Error("expected Prefix to be true")
		}

		if searchParams.NumTypos != 2 {
			t.Errorf("expected NumTypos 2, got %d", searchParams.NumTypos)
		}
	})

	t.Run("With OrgaoID filter", func(t *testing.T) {
		params := &models.CursoSearchParameters{
			Q:       "curso",
			OrgaoID: "org-123",
		}

		searchParams := params.ToSearchParameters()

		if searchParams.FilterBy == "" {
			t.Error("expected filterBy to be set")
		}

		// Should contain action:=read and orgao_id filter
		expectedFilters := "action:=read && record.orgao_id:=org-123"
		if searchParams.FilterBy != expectedFilters {
			t.Errorf("expected filters '%s', got '%s'", expectedFilters, searchParams.FilterBy)
		}
	})

	t.Run("With multiple filters", func(t *testing.T) {
		params := &models.CursoSearchParameters{
			Q:             "curso",
			OrgaoID:       "org-123",
			InstituicaoID: 5,
			Status:        "ativo",
			Modalidade:    "presencial",
			Turno:         "manhã",
			FormatoAula:   "teórico",
		}

		searchParams := params.ToSearchParameters()

		// Should contain all filters
		if searchParams.FilterBy == "" {
			t.Error("expected filterBy to be set")
		}

		// Verify key filters are present
		filters := searchParams.FilterBy
		expectedSubstrings := []string{
			"action:=read",
			"record.orgao_id:=org-123",
			"record.instituicao_id:=5",
			"record.status:=ativo",
			"record.modalidade:=presencial",
			"record.turno:=manhã",
			"record.formato_aula:=teórico",
		}

		for _, substr := range expectedSubstrings {
			if !contains(filters, substr) {
				t.Errorf("expected filter to contain '%s', got '%s'", substr, filters)
			}
		}
	})

	t.Run("Default sort by created_at desc", func(t *testing.T) {
		params := &models.CursoSearchParameters{
			Q: "curso",
		}

		searchParams := params.ToSearchParameters()

		if searchParams.SortBy != "record.created_at:desc" {
			t.Errorf("expected default sortBy 'record.created_at:desc', got '%s'", searchParams.SortBy)
		}
	})

	t.Run("Custom sort", func(t *testing.T) {
		params := &models.CursoSearchParameters{
			Q:      "curso",
			SortBy: "record.titulo:asc",
		}

		searchParams := params.ToSearchParameters()

		if searchParams.SortBy != "record.titulo:asc" {
			t.Errorf("expected custom sortBy, got '%s'", searchParams.SortBy)
		}
	})

	t.Run("With include and exclude fields", func(t *testing.T) {
		params := &models.CursoSearchParameters{
			Q:             "curso",
			IncludeFields: "record.titulo,record.descricao",
			ExcludeFields: "record.metadata",
		}

		searchParams := params.ToSearchParameters()

		if searchParams.IncludeFields != "record.titulo,record.descricao" {
			t.Errorf("expected includeFields, got '%s'", searchParams.IncludeFields)
		}

		if searchParams.ExcludeFields != "record.metadata" {
			t.Errorf("expected excludeFields, got '%s'", searchParams.ExcludeFields)
		}
	})

	t.Run("Only action filter when no other filters", func(t *testing.T) {
		params := &models.CursoSearchParameters{
			Q: "test",
		}

		searchParams := params.ToSearchParameters()

		if searchParams.FilterBy != "action:=read" {
			t.Errorf("expected only action filter, got '%s'", searchParams.FilterBy)
		}
	})
}

func TestEmpregoSearchParameters_ToSearchParameters(t *testing.T) {
	t.Run("Basic search with query", func(t *testing.T) {
		params := &models.EmpregoSearchParameters{
			Q:       "desenvolvedor",
			Page:    1,
			PerPage: 20,
		}

		searchParams := params.ToSearchParameters()

		if searchParams.Q != "desenvolvedor" {
			t.Errorf("expected query 'desenvolvedor', got %s", searchParams.Q)
		}

		if searchParams.QueryBy != "record.titulo,record.descricao,record.pre_requisitos,record.beneficios" {
			t.Errorf("unexpected queryBy: %s", searchParams.QueryBy)
		}

		if searchParams.Page != 1 {
			t.Errorf("expected page 1, got %d", searchParams.Page)
		}

		if searchParams.PerPage != 20 {
			t.Errorf("expected perPage 20, got %d", searchParams.PerPage)
		}
	})

	t.Run("With OrgaoID and EmpresaID filters", func(t *testing.T) {
		params := &models.EmpregoSearchParameters{
			Q:         "vaga",
			OrgaoID:   "org-456",
			EmpresaID: 10,
		}

		searchParams := params.ToSearchParameters()

		filters := searchParams.FilterBy
		expectedSubstrings := []string{
			"action:=read",
			"record.orgao_id:=org-456",
			"record.empresa_id:=10",
		}

		for _, substr := range expectedSubstrings {
			if !contains(filters, substr) {
				t.Errorf("expected filter to contain '%s', got '%s'", substr, filters)
			}
		}
	})

	t.Run("With employment filters", func(t *testing.T) {
		params := &models.EmpregoSearchParameters{
			Q:               "analista",
			Status:          "aberto",
			TipoContratacao: "CLT",
			JornadaTrabalho: "integral",
			Turno:           "comercial",
			EscolaridadeID:  3,
		}

		searchParams := params.ToSearchParameters()

		filters := searchParams.FilterBy
		expectedSubstrings := []string{
			"action:=read",
			"record.status:=aberto",
			"record.tipo_contratacao:=CLT",
			"record.jornada_trabalho:=integral",
			"record.turno:=comercial",
			"record.escolaridade_id:=3",
		}

		for _, substr := range expectedSubstrings {
			if !contains(filters, substr) {
				t.Errorf("expected filter to contain '%s', got '%s'", substr, filters)
			}
		}
	})

	t.Run("With salary range filters", func(t *testing.T) {
		params := &models.EmpregoSearchParameters{
			Q:          "engenheiro",
			SalarioMin: 5000,
			SalarioMax: 10000,
		}

		searchParams := params.ToSearchParameters()

		filters := searchParams.FilterBy
		expectedSubstrings := []string{
			"action:=read",
			"record.salario_min:>=5000",
			"record.salario_max:<=10000",
		}

		for _, substr := range expectedSubstrings {
			if !contains(filters, substr) {
				t.Errorf("expected filter to contain '%s', got '%s'", substr, filters)
			}
		}
	})

	t.Run("Default sort by created_at desc", func(t *testing.T) {
		params := &models.EmpregoSearchParameters{
			Q: "vaga",
		}

		searchParams := params.ToSearchParameters()

		if searchParams.SortBy != "record.created_at:desc" {
			t.Errorf("expected default sortBy, got '%s'", searchParams.SortBy)
		}
	})

	t.Run("Custom sort", func(t *testing.T) {
		params := &models.EmpregoSearchParameters{
			Q:      "vaga",
			SortBy: "record.salario_min:desc",
		}

		searchParams := params.ToSearchParameters()

		if searchParams.SortBy != "record.salario_min:desc" {
			t.Errorf("expected custom sortBy, got '%s'", searchParams.SortBy)
		}
	})

	t.Run("With include and exclude fields", func(t *testing.T) {
		params := &models.EmpregoSearchParameters{
			Q:             "gerente",
			IncludeFields: "record.titulo,record.salario_min",
			ExcludeFields: "record.internal_notes",
		}

		searchParams := params.ToSearchParameters()

		if searchParams.IncludeFields != "record.titulo,record.salario_min" {
			t.Errorf("expected includeFields, got '%s'", searchParams.IncludeFields)
		}

		if searchParams.ExcludeFields != "record.internal_notes" {
			t.Errorf("expected excludeFields, got '%s'", searchParams.ExcludeFields)
		}
	})

	t.Run("Comprehensive filters", func(t *testing.T) {
		params := &models.EmpregoSearchParameters{
			Q:               "desenvolvedor",
			OrgaoID:         "org-789",
			EmpresaID:       15,
			EscolaridadeID:  4,
			Status:          "ativo",
			TipoContratacao: "PJ",
			JornadaTrabalho: "parcial",
			Turno:           "noite",
			SalarioMin:      3000,
			SalarioMax:      8000,
		}

		searchParams := params.ToSearchParameters()

		filters := searchParams.FilterBy
		expectedSubstrings := []string{
			"action:=read",
			"record.orgao_id:=org-789",
			"record.empresa_id:=15",
			"record.escolaridade_id:=4",
			"record.status:=ativo",
			"record.tipo_contratacao:=PJ",
			"record.jornada_trabalho:=parcial",
			"record.turno:=noite",
			"record.salario_min:>=3000",
			"record.salario_max:<=8000",
		}

		for _, substr := range expectedSubstrings {
			if !contains(filters, substr) {
				t.Errorf("expected filter to contain '%s', got '%s'", substr, filters)
			}
		}
	})
}

func TestSearchDocumentsResponse_Structure(t *testing.T) {
	t.Run("Create response with hits", func(t *testing.T) {
		response := &models.SearchDocumentsResponse{
			Found: 5,
			Page:  1,
			Hits: []map[string]interface{}{
				{
					"id":      "doc1",
					"titulo":  "Curso de Go",
					"action":  "read",
					"highlights": []interface{}{},
				},
				{
					"id":      "doc2",
					"titulo":  "Curso de Python",
					"action":  "read",
					"highlights": []interface{}{},
				},
			},
			RequestParams: map[string]interface{}{
				"q":        "curso",
				"query_by": "titulo,descricao",
			},
		}

		if response.Found != 5 {
			t.Errorf("expected Found=5, got %d", response.Found)
		}

		if len(response.Hits) != 2 {
			t.Errorf("expected 2 hits, got %d", len(response.Hits))
		}

		if response.Hits[0]["titulo"] != "Curso de Go" {
			t.Errorf("unexpected hit titulo: %v", response.Hits[0]["titulo"])
		}
	})

	t.Run("Empty response", func(t *testing.T) {
		response := &models.SearchDocumentsResponse{
			Found:         0,
			Page:          1,
			Hits:          []map[string]interface{}{},
			RequestParams: map[string]interface{}{},
		}

		if response.Found != 0 {
			t.Errorf("expected Found=0, got %d", response.Found)
		}

		if len(response.Hits) != 0 {
			t.Errorf("expected 0 hits, got %d", len(response.Hits))
		}
	})
}

func TestMultiCollectionSearchResponse_Structure(t *testing.T) {
	t.Run("Create multi-collection response", func(t *testing.T) {
		response := &models.MultiCollectionSearchResponse{
			TotalFound: 10,
			Results: map[string]models.SearchDocumentsResponse{
				"cursos": {
					Found: 5,
					Page:  1,
					Hits: []map[string]interface{}{
						{"id": "c1", "titulo": "Curso 1"},
					},
				},
				"empregos": {
					Found: 5,
					Page:  1,
					Hits: []map[string]interface{}{
						{"id": "e1", "titulo": "Emprego 1"},
					},
				},
			},
		}

		if response.TotalFound != 10 {
			t.Errorf("expected TotalFound=10, got %d", response.TotalFound)
		}

		if len(response.Results) != 2 {
			t.Errorf("expected 2 collections, got %d", len(response.Results))
		}

		cursosResult, exists := response.Results["cursos"]
		if !exists {
			t.Error("expected cursos results")
		} else if cursosResult.Found != 5 {
			t.Errorf("expected 5 cursos, got %d", cursosResult.Found)
		}

		empregosResult, exists := response.Results["empregos"]
		if !exists {
			t.Error("expected empregos results")
		} else if empregosResult.Found != 5 {
			t.Errorf("expected 5 empregos, got %d", empregosResult.Found)
		}
	})
}

func TestMultiCollectionSearchParameters_Structure(t *testing.T) {
	t.Run("Create multi-collection search params", func(t *testing.T) {
		params := &models.MultiCollectionSearchParameters{
			Collections: []string{"cursos", "empregos"},
			Params: models.SearchParameters{
				Q:       "desenvolvimento",
				QueryBy: "titulo,descricao",
				Page:    1,
				PerPage: 10,
			},
		}

		if len(params.Collections) != 2 {
			t.Errorf("expected 2 collections, got %d", len(params.Collections))
		}

		if params.Params.Q != "desenvolvimento" {
			t.Errorf("expected query 'desenvolvimento', got %s", params.Params.Q)
		}
	})

	t.Run("Empty collections list", func(t *testing.T) {
		params := &models.MultiCollectionSearchParameters{
			Collections: []string{},
			Params: models.SearchParameters{
				Q: "test",
			},
		}

		if len(params.Collections) != 0 {
			t.Errorf("expected 0 collections, got %d", len(params.Collections))
		}
	})
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
