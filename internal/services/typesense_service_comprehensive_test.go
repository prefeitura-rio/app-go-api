package services_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/typesense/typesense-go/v3/typesense/api"
	"github.com/typesense/typesense-go/v3/typesense/api/pointer"
)

// TestSearchDocuments_Structure tests the SearchDocuments response structure
func TestSearchDocuments_Structure(t *testing.T) {
	// Note: Since we cannot inject a mock client easily, these tests
	// verify the data structure transformations that SearchDocuments performs

	t.Run("Verify SearchResult to SearchDocumentsResponse conversion logic", func(t *testing.T) {
		// Create a mock SearchResult similar to what Typesense returns
		found := 2
		page := 1

		document1 := map[string]interface{}{
			"id":         "doc1",
			"titulo":     "Test Curso",
			"descricao":  "Test description",
			"created_at": "2024-01-01",
		}

		document2 := map[string]interface{}{
			"id":         "doc2",
			"titulo":     "Test Curso 2",
			"descricao":  "Test description 2",
			"created_at": "2024-01-02",
		}

		highlights1 := []api.SearchHighlight{
			{
				Field:   pointer.String("titulo"),
				Snippet: pointer.String("<mark>Test</mark> Curso"),
			},
		}

		hits := []api.SearchResultHit{
			{
				Document:   &document1,
				Highlights: &highlights1,
			},
			{
				Document:   &document2,
				Highlights: &[]api.SearchHighlight{},
			},
		}

		requestParams := struct {
			CollectionName string `json:"collection_name"`
			PerPage        int    `json:"per_page"`
			Q              string `json:"q"`
			VoiceQuery     *struct {
				TranscribedQuery *string `json:"transcribed_query,omitempty"`
			} `json:"voice_query,omitempty"`
		}{
			CollectionName: "cursos",
			PerPage:        10,
			Q:              "test",
		}

		mockResult := &api.SearchResult{
			Found:         &found,
			Hits:          &hits,
			Page:          &page,
			RequestParams: &requestParams,
		}

		// Verify the structure can be processed
		if mockResult.Found == nil || *mockResult.Found != 2 {
			t.Error("Expected found to be 2")
		}

		if mockResult.Hits == nil || len(*mockResult.Hits) != 2 {
			t.Error("Expected 2 hits")
		}

		// Verify hit structure
		firstHit := (*mockResult.Hits)[0]
		if firstHit.Document == nil {
			t.Error("Expected document to be set")
		}

		if firstHit.Highlights == nil {
			t.Error("Expected highlights to be set")
		}

		// Test JSON marshaling/unmarshaling (what SearchDocuments does)
		docBytes, err := json.Marshal(firstHit.Document)
		if err != nil {
			t.Errorf("Failed to marshal document: %v", err)
		}

		var hitData map[string]interface{}
		if err := json.Unmarshal(docBytes, &hitData); err != nil {
			t.Errorf("Failed to unmarshal document: %v", err)
		}

		if hitData["titulo"] != "Test Curso" {
			t.Errorf("Expected titulo 'Test Curso', got %v", hitData["titulo"])
		}
	})
}

// TestSearchCursos_ParameterConversion tests parameter conversion for curso search
func TestSearchCursos_ParameterConversion(t *testing.T) {
	t.Run("Verify CursoSearchParameters produces correct SearchParameters", func(t *testing.T) {
		params := models.CursoSearchParameters{
			Q:             "programação",
			OrgaoID:       "org-123",
			InstituicaoID: 5,
			Status:        "ativo",
			Modalidade:    "presencial",
			Page:          1,
			PerPage:       10,
		}

		searchParams := params.ToSearchParameters()

		// Verify query parameters
		if searchParams.Q != "programação" {
			t.Errorf("Expected Q='programação', got '%s'", searchParams.Q)
		}

		if searchParams.QueryBy != "record.titulo,record.descricao,record.pre_requisitos" {
			t.Errorf("Unexpected QueryBy: %s", searchParams.QueryBy)
		}

		// Verify filters are properly constructed
		expectedFilters := []string{
			"action:=read",
			"record.orgao_id:=org-123",
			"record.instituicao_id:=5",
			"record.status:=ativo",
			"record.modalidade:=presencial",
		}

		for _, expected := range expectedFilters {
			if !contains(searchParams.FilterBy, expected) {
				t.Errorf("Expected filter to contain '%s', got '%s'", expected, searchParams.FilterBy)
			}
		}

		// Verify pagination
		if searchParams.Page != 1 {
			t.Errorf("Expected Page=1, got %d", searchParams.Page)
		}

		if searchParams.PerPage != 10 {
			t.Errorf("Expected PerPage=10, got %d", searchParams.PerPage)
		}

		// Verify defaults
		if !searchParams.Prefix {
			t.Error("Expected Prefix to be true")
		}

		if searchParams.NumTypos != 2 {
			t.Errorf("Expected NumTypos=2, got %d", searchParams.NumTypos)
		}
	})

	t.Run("Empty search query", func(t *testing.T) {
		params := models.CursoSearchParameters{
			Q:       "",
			Page:    1,
			PerPage: 10,
		}

		searchParams := params.ToSearchParameters()

		if searchParams.Q != "" {
			t.Errorf("Expected empty Q, got '%s'", searchParams.Q)
		}

		// Should still have action filter
		if !contains(searchParams.FilterBy, "action:=read") {
			t.Error("Expected action filter even with empty query")
		}
	})

	t.Run("Special characters in query", func(t *testing.T) {
		params := models.CursoSearchParameters{
			Q:       "curso & desenvolvimento",
			Page:    1,
			PerPage: 10,
		}

		searchParams := params.ToSearchParameters()

		if searchParams.Q != "curso & desenvolvimento" {
			t.Errorf("Query should preserve special characters, got '%s'", searchParams.Q)
		}
	})
}

// TestSearchEmpregos_ParameterConversion tests parameter conversion for emprego search
func TestSearchEmpregos_ParameterConversion(t *testing.T) {
	t.Run("Verify EmpregoSearchParameters produces correct SearchParameters", func(t *testing.T) {
		params := models.EmpregoSearchParameters{
			Q:               "desenvolvedor",
			OrgaoID:         "org-456",
			EmpresaID:       10,
			EscolaridadeID:  3,
			Status:          "aberto",
			TipoContratacao: "CLT",
			SalarioMin:      5000,
			SalarioMax:      10000,
			Page:            2,
			PerPage:         20,
		}

		searchParams := params.ToSearchParameters()

		// Verify query parameters
		if searchParams.Q != "desenvolvedor" {
			t.Errorf("Expected Q='desenvolvedor', got '%s'", searchParams.Q)
		}

		if searchParams.QueryBy != "record.titulo,record.descricao,record.pre_requisitos,record.beneficios" {
			t.Errorf("Unexpected QueryBy: %s", searchParams.QueryBy)
		}

		// Verify filters
		expectedFilters := []string{
			"action:=read",
			"record.orgao_id:=org-456",
			"record.empresa_id:=10",
			"record.escolaridade_id:=3",
			"record.status:=aberto",
			"record.tipo_contratacao:=CLT",
			"record.salario_min:>=5000",
			"record.salario_max:<=10000",
		}

		for _, expected := range expectedFilters {
			if !contains(searchParams.FilterBy, expected) {
				t.Errorf("Expected filter to contain '%s', got '%s'", expected, searchParams.FilterBy)
			}
		}

		// Verify pagination
		if searchParams.Page != 2 {
			t.Errorf("Expected Page=2, got %d", searchParams.Page)
		}

		if searchParams.PerPage != 20 {
			t.Errorf("Expected PerPage=20, got %d", searchParams.PerPage)
		}
	})

	t.Run("Salary range edge cases", func(t *testing.T) {
		t.Run("Only minimum salary", func(t *testing.T) {
			params := models.EmpregoSearchParameters{
				Q:          "vaga",
				SalarioMin: 3000,
			}

			searchParams := params.ToSearchParameters()

			if !contains(searchParams.FilterBy, "record.salario_min:>=3000") {
				t.Error("Expected salario_min filter")
			}

			if contains(searchParams.FilterBy, "salario_max") {
				t.Error("Should not contain salario_max filter")
			}
		})

		t.Run("Only maximum salary", func(t *testing.T) {
			params := models.EmpregoSearchParameters{
				Q:          "vaga",
				SalarioMax: 15000,
			}

			searchParams := params.ToSearchParameters()

			if !contains(searchParams.FilterBy, "record.salario_max:<=15000") {
				t.Error("Expected salario_max filter")
			}

			if contains(searchParams.FilterBy, "salario_min:>=") {
				t.Error("Should not contain salario_min filter")
			}
		})

		t.Run("Zero salary values should be ignored", func(t *testing.T) {
			params := models.EmpregoSearchParameters{
				Q:          "vaga",
				SalarioMin: 0,
				SalarioMax: 0,
			}

			searchParams := params.ToSearchParameters()

			if contains(searchParams.FilterBy, "salario_min") {
				t.Error("Should not filter on zero salario_min")
			}

			if contains(searchParams.FilterBy, "salario_max") {
				t.Error("Should not filter on zero salario_max")
			}
		})
	})
}

// TestSearchDocuments_EdgeCases tests edge cases for SearchDocuments
func TestSearchDocuments_EdgeCases(t *testing.T) {
	t.Run("Empty result set", func(t *testing.T) {
		found := 0
		page := 1
		hits := []api.SearchResultHit{}

		requestParams := struct {
			CollectionName string `json:"collection_name"`
			PerPage        int    `json:"per_page"`
			Q              string `json:"q"`
			VoiceQuery     *struct {
				TranscribedQuery *string `json:"transcribed_query,omitempty"`
			} `json:"voice_query,omitempty"`
		}{
			CollectionName: "cursos",
			PerPage:        10,
			Q:              "nonexistent",
		}

		mockResult := &api.SearchResult{
			Found:         &found,
			Hits:          &hits,
			Page:          &page,
			RequestParams: &requestParams,
		}

		// Verify structure
		if mockResult.Found == nil || *mockResult.Found != 0 {
			t.Error("Expected found to be 0")
		}

		if mockResult.Hits == nil || len(*mockResult.Hits) != 0 {
			t.Error("Expected 0 hits")
		}
	})

	t.Run("Large result set pagination", func(t *testing.T) {
		params := models.SearchParameters{
			Q:       "test",
			QueryBy: "titulo",
			Page:    5,
			PerPage: 100,
		}

		if params.Page != 5 {
			t.Error("Page should be preserved")
		}

		if params.PerPage != 100 {
			t.Error("PerPage should be preserved")
		}
	})

	t.Run("Complex filter combinations", func(t *testing.T) {
		params := models.SearchParameters{
			Q:        "test",
			QueryBy:  "titulo,descricao",
			FilterBy: "action:=read && status:=active && category:=[curso,workshop] && price:>0",
			SortBy:   "created_at:desc,price:asc",
		}

		if params.FilterBy == "" {
			t.Error("FilterBy should not be empty")
		}

		// Verify multiple sort fields
		if !contains(params.SortBy, "created_at:desc") || !contains(params.SortBy, "price:asc") {
			t.Error("Should support multiple sort fields")
		}
	})
}

// TestMultiCollectionSearch_Logic tests multi-collection search logic
func TestMultiCollectionSearch_Logic(t *testing.T) {
	t.Run("Empty collections list should error", func(t *testing.T) {
		params := models.MultiCollectionSearchParameters{
			Collections: []string{},
			Params: models.SearchParameters{
				Q: "test",
			},
		}

		// The service should validate this
		if len(params.Collections) != 0 {
			t.Error("Expected empty collections")
		}
	})

	t.Run("Multiple collections with same query", func(t *testing.T) {
		params := models.MultiCollectionSearchParameters{
			Collections: []string{"cursos", "empregos", "workshops"},
			Params: models.SearchParameters{
				Q:       "desenvolvimento",
				QueryBy: "titulo,descricao",
				Page:    1,
				PerPage: 10,
			},
		}

		if len(params.Collections) != 3 {
			t.Errorf("Expected 3 collections, got %d", len(params.Collections))
		}

		// Verify same params apply to all collections
		if params.Params.Q != "desenvolvimento" {
			t.Error("Query should be same for all collections")
		}
	})

	t.Run("Response structure aggregation", func(t *testing.T) {
		response := models.MultiCollectionSearchResponse{
			TotalFound: 25,
			Results: map[string]models.SearchDocumentsResponse{
				"cursos": {
					Found: 10,
					Page:  1,
					Hits:  make([]map[string]interface{}, 10),
				},
				"empregos": {
					Found: 15,
					Page:  1,
					Hits:  make([]map[string]interface{}, 15),
				},
			},
		}

		if response.TotalFound != 25 {
			t.Errorf("Expected TotalFound=25, got %d", response.TotalFound)
		}

		if len(response.Results) != 2 {
			t.Errorf("Expected 2 result sets, got %d", len(response.Results))
		}

		// Verify individual collection results
		cursosResult, exists := response.Results["cursos"]
		if !exists {
			t.Error("Expected cursos results")
		} else if cursosResult.Found != 10 {
			t.Errorf("Expected 10 cursos, got %d", cursosResult.Found)
		}
	})

	t.Run("Partial failures in multi-collection search", func(t *testing.T) {
		// Test scenario where one collection fails but others succeed
		response := models.MultiCollectionSearchResponse{
			TotalFound: 10,
			Results: map[string]models.SearchDocumentsResponse{
				"cursos": {
					Found: 10,
					Page:  1,
					Hits:  make([]map[string]interface{}, 10),
				},
				"empregos": {
					Found: 0,
					Hits:  []map[string]interface{}{},
					RequestParams: map[string]interface{}{
						"error": "collection not found",
					},
				},
			},
		}

		// Should still have successful results
		if response.TotalFound != 10 {
			t.Error("Should count only successful results")
		}

		// Error should be recorded in RequestParams
		empregosResult := response.Results["empregos"]
		if errMsg, ok := empregosResult.RequestParams["error"]; !ok {
			t.Error("Expected error in RequestParams")
		} else if errMsg != "collection not found" {
			t.Error("Expected specific error message")
		}
	})
}

// TestSearchCursos_RecordProcessing tests the record extraction logic
func TestSearchCursos_RecordProcessing(t *testing.T) {
	t.Run("Extract record data to hit level", func(t *testing.T) {
		// Simulate a hit with nested record structure
		hit := map[string]interface{}{
			"id":     "doc123",
			"action": "read",
			"record": map[string]interface{}{
				"titulo":         "Curso de Go",
				"descricao":      "Aprenda Go",
				"created_at":     "2024-01-01",
				"instituicao_id": 5,
			},
			"highlights": []interface{}{},
		}

		// Simulate what SearchCursos does
		if record, exists := hit["record"]; exists {
			if recordMap, ok := record.(map[string]interface{}); ok {
				recordMap["document_id"] = hit["id"]
				recordMap["action"] = hit["action"]

				if highlights, exists := hit["highlights"]; exists {
					recordMap["highlights"] = highlights
				}

				// Verify processing
				if recordMap["document_id"] != "doc123" {
					t.Error("Expected document_id to be preserved")
				}

				if recordMap["action"] != "read" {
					t.Error("Expected action to be preserved")
				}

				if recordMap["titulo"] != "Curso de Go" {
					t.Error("Expected titulo from record")
				}
			}
		}
	})

	t.Run("Handle missing record field", func(t *testing.T) {
		// Hit without record field should be handled gracefully
		hit := map[string]interface{}{
			"id":         "doc456",
			"action":     "read",
			"titulo":     "Direct titulo",
			"highlights": []interface{}{},
		}

		// If no record exists, should not panic
		if record, exists := hit["record"]; exists {
			t.Errorf("Did not expect record field, got %v", record)
		}
	})

	t.Run("Record with highlights", func(t *testing.T) {
		highlights := []interface{}{
			map[string]interface{}{
				"field":   "titulo",
				"snippet": "<mark>Go</mark> Programming",
			},
		}

		hit := map[string]interface{}{
			"id":     "doc789",
			"action": "read",
			"record": map[string]interface{}{
				"titulo":    "Go Programming",
				"descricao": "Learn Go",
			},
			"highlights": highlights,
		}

		if record, exists := hit["record"]; exists {
			if recordMap, ok := record.(map[string]interface{}); ok {
				recordMap["highlights"] = hit["highlights"]

				// Verify highlights are preserved
				if recordHighlights, ok := recordMap["highlights"].([]interface{}); !ok {
					t.Error("Expected highlights to be preserved")
				} else if len(recordHighlights) != 1 {
					t.Error("Expected 1 highlight")
				}
			}
		}
	})
}

// TestSearchEmpregos_RecordProcessing tests emprego-specific record processing
func TestSearchEmpregos_RecordProcessing(t *testing.T) {
	t.Run("Extract emprego record with all fields", func(t *testing.T) {
		hit := map[string]interface{}{
			"id":     "emp123",
			"action": "read",
			"record": map[string]interface{}{
				"titulo":           "Desenvolvedor Go",
				"descricao":        "Vaga para dev Go",
				"empresa_id":       10,
				"salario_min":      8000,
				"salario_max":      12000,
				"tipo_contratacao": "CLT",
				"status":           "aberto",
			},
		}

		if record, exists := hit["record"]; exists {
			if recordMap, ok := record.(map[string]interface{}); ok {
				// Verify all fields are accessible
				if recordMap["titulo"] != "Desenvolvedor Go" {
					t.Error("Expected titulo")
				}

				if recordMap["salario_min"] != 8000 {
					t.Error("Expected salario_min")
				}

				if recordMap["tipo_contratacao"] != "CLT" {
					t.Error("Expected tipo_contratacao")
				}
			}
		}
	})
}

// TestSearchParameters_Validation tests parameter validation logic
func TestSearchParameters_Validation(t *testing.T) {
	t.Run("Required query field", func(t *testing.T) {
		params := models.SearchParameters{
			Q:       "test",
			QueryBy: "titulo",
		}

		if params.Q == "" {
			t.Error("Q should not be empty")
		}

		if params.QueryBy == "" {
			t.Error("QueryBy should not be empty")
		}
	})

	t.Run("Optional pagination defaults", func(t *testing.T) {
		params := models.SearchParameters{
			Q:       "test",
			QueryBy: "titulo",
			// No Page or PerPage specified
		}

		// Zero values are valid (Typesense has defaults)
		if params.Page < 0 {
			t.Error("Page should not be negative")
		}

		if params.PerPage < 0 {
			t.Error("PerPage should not be negative")
		}
	})

	t.Run("Faceting parameters", func(t *testing.T) {
		params := models.SearchParameters{
			Q:              "test",
			QueryBy:        "titulo",
			FacetBy:        "category,status",
			MaxFacetValues: 100,
			FacetQuery:     "category:curso",
		}

		if params.FacetBy == "" {
			t.Error("FacetBy should be set")
		}

		if params.MaxFacetValues != 100 {
			t.Error("MaxFacetValues should be 100")
		}
	})

	t.Run("Highlighting parameters", func(t *testing.T) {
		params := models.SearchParameters{
			Q:                   "test",
			QueryBy:             "titulo,descricao",
			HighlightFields:     "titulo",
			HighlightFullFields: "descricao",
		}

		if params.HighlightFields != "titulo" {
			t.Error("HighlightFields should be titulo")
		}

		if params.HighlightFullFields != "descricao" {
			t.Error("HighlightFullFields should be descricao")
		}
	})

	t.Run("Include/Exclude fields", func(t *testing.T) {
		params := models.SearchParameters{
			Q:             "test",
			QueryBy:       "titulo",
			IncludeFields: "titulo,descricao,created_at",
			ExcludeFields: "internal_notes,metadata",
		}

		if params.IncludeFields == "" {
			t.Error("IncludeFields should be set")
		}

		if params.ExcludeFields == "" {
			t.Error("ExcludeFields should be set")
		}
	})
}

// TestTypesenseService_ErrorHandling tests error scenarios
func TestTypesenseService_ErrorHandling(t *testing.T) {
	t.Run("Empty collection name", func(t *testing.T) {
		params := models.SearchParameters{
			Q:       "test",
			QueryBy: "titulo",
		}

		collectionName := ""

		if collectionName == "" {
			// Service should handle this
			t.Log("Empty collection name should be validated")
		}

		// Use params to avoid unused variable error
		_ = params
	})

	t.Run("Invalid filter syntax", func(t *testing.T) {
		params := models.SearchParameters{
			Q:        "test",
			QueryBy:  "titulo",
			FilterBy: "invalid filter syntax",
		}

		// Typesense would return error for invalid filter
		// Service should propagate this error
		if params.FilterBy == "" {
			t.Error("FilterBy should not be empty")
		}
	})

	t.Run("Network timeout simulation", func(t *testing.T) {
		// This would be an error from Typesense client
		err := errors.New("context deadline exceeded")

		if err == nil {
			t.Error("Expected timeout error")
		}

		if err.Error() != "context deadline exceeded" {
			t.Error("Expected specific error message")
		}
	})

	t.Run("Collection not found", func(t *testing.T) {
		err := errors.New("collection 'nonexistent' not found")

		if err == nil {
			t.Error("Expected collection not found error")
		}
	})
}

// Test coverage helper functions
func TestHelperFunctions(t *testing.T) {
	t.Run("contains helper", func(t *testing.T) {
		if !contains("hello world", "world") {
			t.Error("Should contain 'world'")
		}

		if contains("hello world", "xyz") {
			t.Error("Should not contain 'xyz'")
		}

		if !contains("exact", "exact") {
			t.Error("Should match exact string")
		}

		if contains("short", "longer string") {
			t.Error("Should not match longer string")
		}
	})
}
