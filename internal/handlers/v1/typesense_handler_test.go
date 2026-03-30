package v1_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	v1 "github.com/prefeitura-rio/app-go-api/internal/handlers/v1"
	"github.com/stretchr/testify/assert"
)

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
