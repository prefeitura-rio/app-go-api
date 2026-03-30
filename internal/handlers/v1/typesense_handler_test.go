package v1_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	v1 "github.com/prefeitura-rio/app-go-api/internal/handlers/v1"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/stretchr/testify/assert"
)

// mockTypesenseService is a mock implementation of the TypesenseService for testing
type mockTypesenseService struct {
	searchCursosFunc          func(ctx context.Context, params models.CursoSearchParameters) (*models.SearchDocumentsResponse, error)
	searchEmpregosFunc        func(ctx context.Context, params models.EmpregoSearchParameters) (*models.SearchDocumentsResponse, error)
	searchDocumentsFunc       func(ctx context.Context, collection string, params models.SearchParameters) (*models.SearchDocumentsResponse, error)
	searchMultiCollectionFunc func(ctx context.Context, params models.MultiCollectionSearchParameters) (*models.MultiCollectionSearchResponse, error)
}

func (m *mockTypesenseService) SearchCursos(ctx context.Context, params models.CursoSearchParameters) (*models.SearchDocumentsResponse, error) {
	if m.searchCursosFunc != nil {
		return m.searchCursosFunc(ctx, params)
	}
	return &models.SearchDocumentsResponse{}, nil
}

func (m *mockTypesenseService) SearchEmpregos(ctx context.Context, params models.EmpregoSearchParameters) (*models.SearchDocumentsResponse, error) {
	if m.searchEmpregosFunc != nil {
		return m.searchEmpregosFunc(ctx, params)
	}
	return &models.SearchDocumentsResponse{}, nil
}

func (m *mockTypesenseService) SearchDocuments(ctx context.Context, collection string, params models.SearchParameters) (*models.SearchDocumentsResponse, error) {
	if m.searchDocumentsFunc != nil {
		return m.searchDocumentsFunc(ctx, collection, params)
	}
	return &models.SearchDocumentsResponse{}, nil
}

func (m *mockTypesenseService) SearchMultiCollection(ctx context.Context, params models.MultiCollectionSearchParameters) (*models.MultiCollectionSearchResponse, error) {
	if m.searchMultiCollectionFunc != nil {
		return m.searchMultiCollectionFunc(ctx, params)
	}
	return &models.MultiCollectionSearchResponse{}, nil
}

// Test SearchCursos endpoint with invalid JSON
func TestTypesenseHandler_SearchCursos_InvalidJSON(t *testing.T) {
	// Note: Creating TypesenseHandler requires env vars, so test with nil handler
	// We're testing input validation which happens before service call
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Create a minimal handler setup that will still test validation
	h := &v1.TypesenseHandler{}
	r.POST("/api/v1/typesense/search-cursos", h.SearchCursos)

	body := []byte(`{invalid json}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/typesense/search-cursos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Parâmetros de busca inválidos")
}

// Test SearchCursos endpoint with missing query term
func TestTypesenseHandler_SearchCursos_MissingQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	h := &v1.TypesenseHandler{}
	r.POST("/api/v1/typesense/search-cursos", h.SearchCursos)

	body := []byte(`{"q": ""}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/typesense/search-cursos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Termo de busca (q) é obrigatório")
}

// Test SearchEmpregos endpoint with invalid JSON
func TestTypesenseHandler_SearchEmpregos_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	h := &v1.TypesenseHandler{}
	r.POST("/api/v1/typesense/search-empregos", h.SearchEmpregos)

	body := []byte(`{invalid}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/typesense/search-empregos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Parâmetros de busca inválidos")
}

// Test SearchEmpregos endpoint with missing query term
func TestTypesenseHandler_SearchEmpregos_MissingQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	h := &v1.TypesenseHandler{}
	r.POST("/api/v1/typesense/search-empregos", h.SearchEmpregos)

	body := []byte(`{"q": ""}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/typesense/search-empregos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Termo de busca (q) é obrigatório")
}

// Test SearchMultiCollection endpoint with invalid JSON
func TestTypesenseHandler_SearchMultiCollection_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	h := &v1.TypesenseHandler{}
	r.POST("/api/v1/typesense/multi-search", h.SearchMultiCollection)

	body := []byte(`{invalid}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/typesense/multi-search", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Parâmetros de busca inválidos")
}

// Test SearchMultiCollection with empty collections array
func TestTypesenseHandler_SearchMultiCollection_EmptyCollections(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	h := &v1.TypesenseHandler{}
	r.POST("/api/v1/typesense/multi-search", h.SearchMultiCollection)

	body := []byte(`{"collections": [], "params": {"q": "test", "query_by": "title"}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/typesense/multi-search", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "É necessário especificar pelo menos uma coleção")
}

// Test SearchMultiCollection with missing query term
func TestTypesenseHandler_SearchMultiCollection_MissingQ(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	h := &v1.TypesenseHandler{}
	r.POST("/api/v1/typesense/multi-search", h.SearchMultiCollection)

	body := []byte(`{"collections": ["cursos"], "params": {"q": "", "query_by": "title"}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/typesense/multi-search", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Termo de busca (q) e campos para busca (query_by) são obrigatórios")
}

// Test SearchMultiCollection with missing query_by
func TestTypesenseHandler_SearchMultiCollection_MissingQueryBy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	h := &v1.TypesenseHandler{}
	r.POST("/api/v1/typesense/multi-search", h.SearchMultiCollection)

	body := []byte(`{"collections": ["cursos"], "params": {"q": "test", "query_by": ""}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/typesense/multi-search", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Termo de busca (q) e campos para busca (query_by) são obrigatórios")
}

// Test SearchDocuments with invalid JSON
func TestTypesenseHandler_SearchDocuments_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	h := &v1.TypesenseHandler{}
	r.POST("/api/v1/typesense/collections/:collection/documents/search", h.SearchDocuments)

	body := []byte(`{invalid}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/typesense/collections/cursos/documents/search", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Parâmetros de busca inválidos")
}

// Test SearchDocuments with missing collection name
func TestTypesenseHandler_SearchDocuments_EmptyCollection(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	h := &v1.TypesenseHandler{}
	r.POST("/api/v1/typesense/collections/:collection/documents/search", func(c *gin.Context) {
		// Force empty collection param
		c.Params = gin.Params{{Key: "collection", Value: ""}}
		h.SearchDocuments(c)
	})

	body := []byte(`{"q": "test", "query_by": "title"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/typesense/collections/test/documents/search", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Nome da coleção é obrigatório")
}


// Test empty request body
func TestTypesenseHandler_SearchCursos_EmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	h := &v1.TypesenseHandler{}
	r.POST("/api/v1/typesense/search-cursos", h.SearchCursos)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/typesense/search-cursos", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test SearchCursos with service error
func TestTypesenseHandler_SearchCursos_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockSvc := &mockTypesenseService{
		searchCursosFunc: func(ctx context.Context, params models.CursoSearchParameters) (*models.SearchDocumentsResponse, error) {
			return nil, errors.New("typesense connection error")
		},
	}

	h := v1.NewTypesenseHandlerWithService(mockSvc)
	r.POST("/api/v1/typesense/search-cursos", h.SearchCursos)

	body := []byte(`{"q": "engenharia"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/typesense/search-cursos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Erro ao buscar cursos")
}

// Test SearchCursos with valid parameters and success
func TestTypesenseHandler_SearchCursos_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockSvc := &mockTypesenseService{
		searchCursosFunc: func(ctx context.Context, params models.CursoSearchParameters) (*models.SearchDocumentsResponse, error) {
			return &models.SearchDocumentsResponse{
				Found: 10,
				Hits:  []map[string]interface{}{{"id": "1", "title": "Curso de Engenharia"}},
			}, nil
		},
	}

	h := v1.NewTypesenseHandlerWithService(mockSvc)
	r.POST("/api/v1/typesense/search-cursos", h.SearchCursos)

	body := []byte(`{"q": "engenharia", "page": 1, "per_page": 10}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/typesense/search-cursos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Curso de Engenharia")
}

// Test SearchCursos with filters
func TestTypesenseHandler_SearchCursos_WithFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	var capturedParams models.CursoSearchParameters
	mockSvc := &mockTypesenseService{
		searchCursosFunc: func(ctx context.Context, params models.CursoSearchParameters) (*models.SearchDocumentsResponse, error) {
			capturedParams = params
			return &models.SearchDocumentsResponse{Found: 5}, nil
		},
	}

	h := v1.NewTypesenseHandlerWithService(mockSvc)
	r.POST("/api/v1/typesense/search-cursos", h.SearchCursos)

	body := []byte(`{
		"q": "engenharia",
		"orgao_id": "ORG123",
		"instituicao_id": 42,
		"status": "ativo",
		"modalidade": "presencial",
		"turno": "noite",
		"formato_aula": "hibrido"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/typesense/search-cursos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "engenharia", capturedParams.Q)
	assert.Equal(t, "ORG123", capturedParams.OrgaoID)
	assert.Equal(t, 42, capturedParams.InstituicaoID)
	assert.Equal(t, "ativo", capturedParams.Status)
	assert.Equal(t, "presencial", capturedParams.Modalidade)
	assert.Equal(t, "noite", capturedParams.Turno)
	assert.Equal(t, "hibrido", capturedParams.FormatoAula)
}

// Test SearchCursos with empty result set
func TestTypesenseHandler_SearchCursos_EmptyResults(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockSvc := &mockTypesenseService{
		searchCursosFunc: func(ctx context.Context, params models.CursoSearchParameters) (*models.SearchDocumentsResponse, error) {
			return &models.SearchDocumentsResponse{
				Found: 0,
				Hits:  []map[string]interface{}{},
			}, nil
		},
	}

	h := v1.NewTypesenseHandlerWithService(mockSvc)
	r.POST("/api/v1/typesense/search-cursos", h.SearchCursos)

	body := []byte(`{"q": "nonexistent"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/typesense/search-cursos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"found":0`)
}

// Test SearchEmpregos with service error
func TestTypesenseHandler_SearchEmpregos_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockSvc := &mockTypesenseService{
		searchEmpregosFunc: func(ctx context.Context, params models.EmpregoSearchParameters) (*models.SearchDocumentsResponse, error) {
			return nil, errors.New("database timeout")
		},
	}

	h := v1.NewTypesenseHandlerWithService(mockSvc)
	r.POST("/api/v1/typesense/search-empregos", h.SearchEmpregos)

	body := []byte(`{"q": "desenvolvedor"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/typesense/search-empregos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Erro ao buscar empregos")
}

// Test SearchEmpregos with success
func TestTypesenseHandler_SearchEmpregos_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockSvc := &mockTypesenseService{
		searchEmpregosFunc: func(ctx context.Context, params models.EmpregoSearchParameters) (*models.SearchDocumentsResponse, error) {
			return &models.SearchDocumentsResponse{
				Found: 20,
				Hits:  []map[string]interface{}{{"id": "1", "title": "Desenvolvedor Go"}},
			}, nil
		},
	}

	h := v1.NewTypesenseHandlerWithService(mockSvc)
	r.POST("/api/v1/typesense/search-empregos", h.SearchEmpregos)

	body := []byte(`{"q": "desenvolvedor", "page": 2, "per_page": 20}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/typesense/search-empregos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Desenvolvedor Go")
}

// Test SearchEmpregos with filters
func TestTypesenseHandler_SearchEmpregos_WithFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	var capturedParams models.EmpregoSearchParameters
	mockSvc := &mockTypesenseService{
		searchEmpregosFunc: func(ctx context.Context, params models.EmpregoSearchParameters) (*models.SearchDocumentsResponse, error) {
			capturedParams = params
			return &models.SearchDocumentsResponse{Found: 3}, nil
		},
	}

	h := v1.NewTypesenseHandlerWithService(mockSvc)
	r.POST("/api/v1/typesense/search-empregos", h.SearchEmpregos)

	body := []byte(`{
		"q": "desenvolvedor",
		"orgao_id": "ORG456",
		"empresa_id": 99,
		"escolaridade_id": 3,
		"status": "aberto",
		"tipo_contratacao": "CLT",
		"jornada_trabalho": "40h",
		"turno": "comercial",
		"salario_min": 5000,
		"salario_max": 10000
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/typesense/search-empregos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "desenvolvedor", capturedParams.Q)
	assert.Equal(t, "ORG456", capturedParams.OrgaoID)
	assert.Equal(t, 99, capturedParams.EmpresaID)
	assert.Equal(t, 3, capturedParams.EscolaridadeID)
	assert.Equal(t, "aberto", capturedParams.Status)
	assert.Equal(t, "CLT", capturedParams.TipoContratacao)
	assert.Equal(t, "40h", capturedParams.JornadaTrabalho)
	assert.Equal(t, "comercial", capturedParams.Turno)
	assert.Equal(t, 5000, capturedParams.SalarioMin)
	assert.Equal(t, 10000, capturedParams.SalarioMax)
}

// Test SearchEmpregos with empty results
func TestTypesenseHandler_SearchEmpregos_EmptyResults(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockSvc := &mockTypesenseService{
		searchEmpregosFunc: func(ctx context.Context, params models.EmpregoSearchParameters) (*models.SearchDocumentsResponse, error) {
			return &models.SearchDocumentsResponse{
				Found: 0,
				Hits:  []map[string]interface{}{},
			}, nil
		},
	}

	h := v1.NewTypesenseHandlerWithService(mockSvc)
	r.POST("/api/v1/typesense/search-empregos", h.SearchEmpregos)

	body := []byte(`{"q": "impossible job"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/typesense/search-empregos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"found":0`)
}

// Test SearchDocuments with service error
func TestTypesenseHandler_SearchDocuments_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockSvc := &mockTypesenseService{
		searchDocumentsFunc: func(ctx context.Context, collection string, params models.SearchParameters) (*models.SearchDocumentsResponse, error) {
			return nil, errors.New("collection not found")
		},
	}

	h := v1.NewTypesenseHandlerWithService(mockSvc)
	r.POST("/api/v1/typesense/collections/:collection/documents/search", h.SearchDocuments)

	body := []byte(`{"q": "test", "query_by": "title"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/typesense/collections/mycollection/documents/search", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Erro ao buscar documentos")
}

// Test SearchDocuments with success
func TestTypesenseHandler_SearchDocuments_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	var capturedCollection string
	var capturedParams models.SearchParameters
	mockSvc := &mockTypesenseService{
		searchDocumentsFunc: func(ctx context.Context, collection string, params models.SearchParameters) (*models.SearchDocumentsResponse, error) {
			capturedCollection = collection
			capturedParams = params
			return &models.SearchDocumentsResponse{
				Found: 15,
				Hits:  []map[string]interface{}{{"id": "doc1", "title": "Document 1"}},
			}, nil
		},
	}

	h := v1.NewTypesenseHandlerWithService(mockSvc)
	r.POST("/api/v1/typesense/collections/:collection/documents/search", h.SearchDocuments)

	body := []byte(`{"q": "test", "query_by": "title,description", "page": 1, "per_page": 25}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/typesense/collections/documents/documents/search", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "documents", capturedCollection)
	assert.Equal(t, "test", capturedParams.Q)
	assert.Equal(t, "title,description", capturedParams.QueryBy)
	assert.Equal(t, 1, capturedParams.Page)
	assert.Equal(t, 25, capturedParams.PerPage)
}

// Test SearchDocuments with all optional parameters
func TestTypesenseHandler_SearchDocuments_WithAllParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	var capturedParams models.SearchParameters
	mockSvc := &mockTypesenseService{
		searchDocumentsFunc: func(ctx context.Context, collection string, params models.SearchParameters) (*models.SearchDocumentsResponse, error) {
			capturedParams = params
			return &models.SearchDocumentsResponse{Found: 5}, nil
		},
	}

	h := v1.NewTypesenseHandlerWithService(mockSvc)
	r.POST("/api/v1/typesense/collections/:collection/documents/search", h.SearchDocuments)

	body := []byte(`{
		"q": "search term",
		"query_by": "title,body",
		"filter_by": "status:=active",
		"sort_by": "created_at:desc",
		"page": 2,
		"per_page": 50,
		"facet_by": "category",
		"max_facet_values": 10,
		"facet_query": "tech",
		"include_fields": "id,title",
		"exclude_fields": "internal",
		"highlight_fields": "title",
		"highlight_full_fields": "body",
		"prefix": true,
		"num_typos": 1
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/typesense/collections/articles/documents/search", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "search term", capturedParams.Q)
	assert.Equal(t, "title,body", capturedParams.QueryBy)
	assert.Equal(t, "status:=active", capturedParams.FilterBy)
	assert.Equal(t, "created_at:desc", capturedParams.SortBy)
	assert.Equal(t, 2, capturedParams.Page)
	assert.Equal(t, 50, capturedParams.PerPage)
	assert.Equal(t, "category", capturedParams.FacetBy)
	assert.Equal(t, 10, capturedParams.MaxFacetValues)
	assert.Equal(t, "tech", capturedParams.FacetQuery)
	assert.Equal(t, "id,title", capturedParams.IncludeFields)
	assert.Equal(t, "internal", capturedParams.ExcludeFields)
	assert.Equal(t, "title", capturedParams.HighlightFields)
	assert.Equal(t, "body", capturedParams.HighlightFullFields)
	assert.True(t, capturedParams.Prefix)
	assert.Equal(t, 1, capturedParams.NumTypos)
}

// Test SearchDocuments with empty results
func TestTypesenseHandler_SearchDocuments_EmptyResults(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockSvc := &mockTypesenseService{
		searchDocumentsFunc: func(ctx context.Context, collection string, params models.SearchParameters) (*models.SearchDocumentsResponse, error) {
			return &models.SearchDocumentsResponse{
				Found: 0,
				Hits:  []map[string]interface{}{},
			}, nil
		},
	}

	h := v1.NewTypesenseHandlerWithService(mockSvc)
	r.POST("/api/v1/typesense/collections/:collection/documents/search", h.SearchDocuments)

	body := []byte(`{"q": "nonexistent", "query_by": "title"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/typesense/collections/test/documents/search", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"found":0`)
}

// Test SearchMultiCollection with service error
func TestTypesenseHandler_SearchMultiCollection_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockSvc := &mockTypesenseService{
		searchMultiCollectionFunc: func(ctx context.Context, params models.MultiCollectionSearchParameters) (*models.MultiCollectionSearchResponse, error) {
			return nil, errors.New("multi-search failed")
		},
	}

	h := v1.NewTypesenseHandlerWithService(mockSvc)
	r.POST("/api/v1/typesense/multi-search", h.SearchMultiCollection)

	body := []byte(`{"collections": ["cursos", "empregos"], "params": {"q": "test", "query_by": "title"}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/typesense/multi-search", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Erro ao realizar busca em múltiplas coleções")
}

// Test SearchMultiCollection with success
func TestTypesenseHandler_SearchMultiCollection_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockSvc := &mockTypesenseService{
		searchMultiCollectionFunc: func(ctx context.Context, params models.MultiCollectionSearchParameters) (*models.MultiCollectionSearchResponse, error) {
			return &models.MultiCollectionSearchResponse{
				TotalFound: 8,
				Results: map[string]models.SearchDocumentsResponse{
					"cursos":   {Found: 5},
					"empregos": {Found: 3},
				},
			}, nil
		},
	}

	h := v1.NewTypesenseHandlerWithService(mockSvc)
	r.POST("/api/v1/typesense/multi-search", h.SearchMultiCollection)

	body := []byte(`{"collections": ["cursos", "empregos"], "params": {"q": "engenharia", "query_by": "title"}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/typesense/multi-search", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
