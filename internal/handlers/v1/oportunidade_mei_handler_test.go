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

// Test basic handler initialization
func TestNewOportunidadeMEIHandler(t *testing.T) {
	handler := v1.NewOportunidadeMEIHandler(nil)
	assert.NotNil(t, handler)
}

// Test invalid JSON handling for Create
func TestOportunidadeMEIHandler_Create_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewOportunidadeMEIHandler(nil)
	r.POST("/api/v1/oportunidades-mei", h.Create)

	body := []byte(`{invalid json}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/oportunidades-mei", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Dados inválidos")
}

// Test invalid JSON handling for CreateDraft
func TestOportunidadeMEIHandler_CreateDraft_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewOportunidadeMEIHandler(nil)
	r.POST("/api/v1/oportunidades-mei/draft", h.CreateDraft)

	body := []byte(`{invalid}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/oportunidades-mei/draft", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test invalid ID parameter for GetByID
func TestOportunidadeMEIHandler_GetByID_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewOportunidadeMEIHandler(nil)
	r.GET("/api/v1/oportunidades-mei/:id", h.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/oportunidades-mei/invalid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "ID inválido")
}

// Test invalid ID parameter for Update
func TestOportunidadeMEIHandler_Update_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewOportunidadeMEIHandler(nil)
	r.PUT("/api/v1/oportunidades-mei/:id", h.Update)

	body := []byte(`{"titulo": "Test"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/oportunidades-mei/invalid", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test invalid JSON for Update
func TestOportunidadeMEIHandler_Update_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewOportunidadeMEIHandler(nil)
	r.PUT("/api/v1/oportunidades-mei/:id", h.Update)

	body := []byte(`{invalid}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/oportunidades-mei/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test invalid ID parameter for Delete
func TestOportunidadeMEIHandler_Delete_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewOportunidadeMEIHandler(nil)
	r.DELETE("/api/v1/oportunidades-mei/:id", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/oportunidades-mei/invalid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test invalid ID parameter for Publish
func TestOportunidadeMEIHandler_Publish_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewOportunidadeMEIHandler(nil)
	r.PUT("/api/v1/oportunidades-mei/:id/publish", h.Publish)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/oportunidades-mei/invalid/publish", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test invalid status filter for List
func TestOportunidadeMEIHandler_List_InvalidStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewOportunidadeMEIHandler(nil)
	r.GET("/api/v1/oportunidades-mei", h.List)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/oportunidades-mei?status=invalid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Status inválido")
}

