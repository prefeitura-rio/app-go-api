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
func TestNewCourseHandler(t *testing.T) {
	handler := v1.NewCourseHandler(nil, nil, nil)
	assert.NotNil(t, handler)
}

// Test WithCache method
func TestCourseHandler_WithCache(t *testing.T) {
	handler := v1.NewCourseHandler(nil, nil, nil)
	result := handler.WithCache(nil)
	assert.NotNil(t, result)
}

// Test Create endpoint with invalid JSON
func TestCourseHandler_Create_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewCourseHandler(nil, nil, nil)
	r.POST("/api/v1/courses", h.Create)

	body := []byte(`{invalid json}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Dados inválidos")
}

// Test CreateDraft endpoint with invalid JSON
func TestCourseHandler_CreateDraft_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewCourseHandler(nil, nil, nil)
	r.POST("/api/v1/courses/draft", h.CreateDraft)

	body := []byte(`{invalid}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses/draft", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test Update endpoint with invalid ID
func TestCourseHandler_Update_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewCourseHandler(nil, nil, nil)
	r.PUT("/api/v1/courses/:courseId", h.Update)

	body := []byte(`{"titulo": "Test"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/courses/invalid", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "ID do curso inválido")
}


// Test GetByID endpoint with invalid ID
func TestCourseHandler_GetByID_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewCourseHandler(nil, nil, nil)
	r.GET("/api/v1/courses/:courseId", h.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses/invalid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "ID do curso inválido")
}

// Test Delete endpoint with invalid ID
func TestCourseHandler_Delete_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewCourseHandler(nil, nil, nil)
	r.DELETE("/api/v1/courses/:courseId", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/courses/invalid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

