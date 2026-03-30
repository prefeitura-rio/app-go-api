package v1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	v1 "github.com/prefeitura-rio/app-go-api/internal/handlers/v1"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockOportunidadeMEIService is a mock implementation of OportunidadeMEIServiceInterface
type MockOportunidadeMEIService struct {
	mock.Mock
}

func (m *MockOportunidadeMEIService) Create(ctx context.Context, oportunidade *models.OportunidadeMEI, isDraft bool) (int, error) {
	args := m.Called(ctx, oportunidade, isDraft)
	return args.Int(0), args.Error(1)
}

func (m *MockOportunidadeMEIService) GetByID(ctx context.Context, id int) (*models.OportunidadeMEI, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.OportunidadeMEI), args.Error(1)
}

func (m *MockOportunidadeMEIService) Update(ctx context.Context, oportunidade *models.OportunidadeMEI) error {
	args := m.Called(ctx, oportunidade)
	return args.Error(0)
}

func (m *MockOportunidadeMEIService) Publish(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOportunidadeMEIService) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOportunidadeMEIService) List(ctx context.Context, filters map[string]interface{}, titulo string, page, pageSize int) ([]*models.OportunidadeMEI, int, error) {
	args := m.Called(ctx, filters, titulo, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.OportunidadeMEI), args.Int(1), args.Error(2)
}

func (m *MockOportunidadeMEIService) ListByStatus(ctx context.Context, status models.StatusOportunidadeMEI, page, pageSize int) ([]*models.OportunidadeMEI, int, error) {
	args := m.Called(ctx, status, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.OportunidadeMEI), args.Int(1), args.Error(2)
}

func (m *MockOportunidadeMEIService) ListDrafts(ctx context.Context, orgaoID, titulo string, page, pageSize int) ([]*models.OportunidadeMEI, int, error) {
	args := m.Called(ctx, orgaoID, titulo, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.OportunidadeMEI), args.Int(1), args.Error(2)
}

func (m *MockOportunidadeMEIService) ListActive(ctx context.Context, page, pageSize int) ([]*models.OportunidadeMEI, int, error) {
	args := m.Called(ctx, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.OportunidadeMEI), args.Int(1), args.Error(2)
}

func (m *MockOportunidadeMEIService) ListByOrgao(ctx context.Context, orgaoID string, page, pageSize int) ([]*models.OportunidadeMEI, int, error) {
	args := m.Called(ctx, orgaoID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.OportunidadeMEI), args.Int(1), args.Error(2)
}

func (m *MockOportunidadeMEIService) UpdateExpiredOpportunities(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

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

// ========== Comprehensive Tests with Mocks ==========

// Test Create - Success
func TestOportunidadeMEIHandler_Create_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockOportunidadeMEIService)
	handler := v1.NewOportunidadeMEIHandler(mockService)

	r := gin.New()
	r.POST("/api/v1/oportunidades-mei", handler.Create)

	futureDate := time.Now().Add(30 * 24 * time.Hour)
	oportunidade := models.OportunidadeMEI{
		Titulo:           "Oportunidade Test",
		DescricaoServico: "Test Description",
		OrgaoID:          "org123",
		DataExpiracao:    &futureDate,
	}

	mockService.On("Create", mock.Anything, mock.Anything, false).Return(1, nil)

	body, _ := json.Marshal(oportunidade)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/oportunidades-mei", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.OportunidadeMEI
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 1, response.ID)

	mockService.AssertExpectations(t)
}

// Test Create - Service Error
func TestOportunidadeMEIHandler_Create_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockOportunidadeMEIService)
	handler := v1.NewOportunidadeMEIHandler(mockService)

	r := gin.New()
	r.POST("/api/v1/oportunidades-mei", handler.Create)

	oportunidade := models.OportunidadeMEI{
		Titulo: "Test",
	}

	mockService.On("Create", mock.Anything, mock.Anything, false).Return(0, errors.New("validation error"))

	body, _ := json.Marshal(oportunidade)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/oportunidades-mei", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "validation error")

	mockService.AssertExpectations(t)
}

// Test CreateDraft - Success
func TestOportunidadeMEIHandler_CreateDraft_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockOportunidadeMEIService)
	handler := v1.NewOportunidadeMEIHandler(mockService)

	r := gin.New()
	r.POST("/api/v1/oportunidades-mei/draft", handler.CreateDraft)

	oportunidade := models.OportunidadeMEI{
		Titulo:    "Draft Oportunidade",
		DescricaoServico: "Draft Description",
	}

	mockService.On("Create", mock.Anything, mock.Anything, true).Return(2, nil)

	body, _ := json.Marshal(oportunidade)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/oportunidades-mei/draft", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.OportunidadeMEI
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 2, response.ID)

	mockService.AssertExpectations(t)
}

// Test GetByID - Success
func TestOportunidadeMEIHandler_GetByID_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockOportunidadeMEIService)
	handler := v1.NewOportunidadeMEIHandler(mockService)

	r := gin.New()
	r.GET("/api/v1/oportunidades-mei/:id", handler.GetByID)

	oportunidade := &models.OportunidadeMEI{
		ID:        1,
		Titulo:    "Test Oportunidade",
		DescricaoServico: "Test Description",
	}

	mockService.On("GetByID", mock.Anything, 1).Return(oportunidade, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/oportunidades-mei/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.OportunidadeMEI
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Test Oportunidade", response.Titulo)

	mockService.AssertExpectations(t)
}

// Test GetByID - Not Found
func TestOportunidadeMEIHandler_GetByID_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockOportunidadeMEIService)
	handler := v1.NewOportunidadeMEIHandler(mockService)

	r := gin.New()
	r.GET("/api/v1/oportunidades-mei/:id", handler.GetByID)

	mockService.On("GetByID", mock.Anything, 999).Return(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/oportunidades-mei/999", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Oportunidade não encontrada")

	mockService.AssertExpectations(t)
}

// Test GetByID - Service Error
func TestOportunidadeMEIHandler_GetByID_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockOportunidadeMEIService)
	handler := v1.NewOportunidadeMEIHandler(mockService)

	r := gin.New()
	r.GET("/api/v1/oportunidades-mei/:id", handler.GetByID)

	mockService.On("GetByID", mock.Anything, 1).Return(nil, errors.New("database error"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/oportunidades-mei/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockService.AssertExpectations(t)
}

// Test Update - Success
func TestOportunidadeMEIHandler_Update_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockOportunidadeMEIService)
	handler := v1.NewOportunidadeMEIHandler(mockService)

	r := gin.New()
	r.PUT("/api/v1/oportunidades-mei/:id", handler.Update)

	oportunidade := models.OportunidadeMEI{
		Titulo:           "Updated Title",
		DescricaoServico: "Updated Description",
	}

	mockService.On("Update", mock.Anything, mock.MatchedBy(func(o *models.OportunidadeMEI) bool {
		return o.ID == 1 && o.Titulo == "Updated Title"
	})).Return(nil)

	body, _ := json.Marshal(oportunidade)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/oportunidades-mei/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

// Test Update - Service Error
func TestOportunidadeMEIHandler_Update_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockOportunidadeMEIService)
	handler := v1.NewOportunidadeMEIHandler(mockService)

	r := gin.New()
	r.PUT("/api/v1/oportunidades-mei/:id", handler.Update)

	oportunidade := models.OportunidadeMEI{
		Titulo: "Test",
	}

	mockService.On("Update", mock.Anything, mock.Anything).Return(errors.New("update error"))

	body, _ := json.Marshal(oportunidade)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/oportunidades-mei/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	mockService.AssertExpectations(t)
}

// Test Delete - Success
func TestOportunidadeMEIHandler_Delete_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockOportunidadeMEIService)
	handler := v1.NewOportunidadeMEIHandler(mockService)

	r := gin.New()
	r.DELETE("/api/v1/oportunidades-mei/:id", handler.Delete)

	mockService.On("Delete", mock.Anything, 1).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/oportunidades-mei/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Oportunidade excluída com sucesso", response["message"])

	mockService.AssertExpectations(t)
}

// Test Delete - Service Error
func TestOportunidadeMEIHandler_Delete_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockOportunidadeMEIService)
	handler := v1.NewOportunidadeMEIHandler(mockService)

	r := gin.New()
	r.DELETE("/api/v1/oportunidades-mei/:id", handler.Delete)

	mockService.On("Delete", mock.Anything, 1).Return(errors.New("delete error"))

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/oportunidades-mei/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockService.AssertExpectations(t)
}

// Test Publish - Success
func TestOportunidadeMEIHandler_Publish_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockOportunidadeMEIService)
	handler := v1.NewOportunidadeMEIHandler(mockService)

	r := gin.New()
	r.PUT("/api/v1/oportunidades-mei/:id/publish", handler.Publish)

	oportunidade := &models.OportunidadeMEI{
		ID:     1,
		Titulo: "Published Oportunidade",
		Status: models.StatusOportunidadeActive,
	}

	mockService.On("Publish", mock.Anything, 1).Return(nil)
	mockService.On("GetByID", mock.Anything, 1).Return(oportunidade, nil)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/oportunidades-mei/1/publish", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.OportunidadeMEI
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, models.StatusOportunidadeActive, response.Status)

	mockService.AssertExpectations(t)
}

// Test Publish - Service Error
func TestOportunidadeMEIHandler_Publish_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockOportunidadeMEIService)
	handler := v1.NewOportunidadeMEIHandler(mockService)

	r := gin.New()
	r.PUT("/api/v1/oportunidades-mei/:id/publish", handler.Publish)

	mockService.On("Publish", mock.Anything, 1).Return(errors.New("cannot publish"))

	req := httptest.NewRequest(http.MethodPut, "/api/v1/oportunidades-mei/1/publish", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	mockService.AssertExpectations(t)
}

// Test Publish - GetByID Error After Publish
func TestOportunidadeMEIHandler_Publish_GetByIDError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockOportunidadeMEIService)
	handler := v1.NewOportunidadeMEIHandler(mockService)

	r := gin.New()
	r.PUT("/api/v1/oportunidades-mei/:id/publish", handler.Publish)

	mockService.On("Publish", mock.Anything, 1).Return(nil)
	mockService.On("GetByID", mock.Anything, 1).Return(nil, errors.New("database error"))

	req := httptest.NewRequest(http.MethodPut, "/api/v1/oportunidades-mei/1/publish", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockService.AssertExpectations(t)
}

// Test List - Success
func TestOportunidadeMEIHandler_List_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockOportunidadeMEIService)
	handler := v1.NewOportunidadeMEIHandler(mockService)

	r := gin.New()
	r.GET("/api/v1/oportunidades-mei", handler.List)

	oportunidades := []*models.OportunidadeMEI{
		{ID: 1, Titulo: "Oportunidade 1", Status: models.StatusOportunidadeActive},
		{ID: 2, Titulo: "Oportunidade 2", Status: models.StatusOportunidadeActive},
	}

	mockService.On("List", mock.Anything, mock.MatchedBy(func(filters map[string]interface{}) bool {
		return filters["status"] == models.StatusOportunidadeActive
	}), "", 1, 10).Return(oportunidades, 2, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/oportunidades-mei?page=1&pageSize=10", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	mockService.AssertExpectations(t)
}

// Test List - With Status Filter
func TestOportunidadeMEIHandler_List_WithStatusFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockOportunidadeMEIService)
	handler := v1.NewOportunidadeMEIHandler(mockService)

	r := gin.New()
	r.GET("/api/v1/oportunidades-mei", handler.List)

	drafts := []*models.OportunidadeMEI{
		{ID: 1, Titulo: "Draft 1", Status: models.StatusOportunidadeDraft},
	}

	mockService.On("List", mock.Anything, mock.MatchedBy(func(filters map[string]interface{}) bool {
		return filters["status"] == models.StatusOportunidadeDraft
	}), "", 1, 10).Return(drafts, 1, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/oportunidades-mei?status=draft", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

// Test List - Service Error
func TestOportunidadeMEIHandler_List_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockOportunidadeMEIService)
	handler := v1.NewOportunidadeMEIHandler(mockService)

	r := gin.New()
	r.GET("/api/v1/oportunidades-mei", handler.List)

	mockService.On("List", mock.Anything, mock.Anything, "", 1, 10).Return(nil, 0, errors.New("database error"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/oportunidades-mei", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockService.AssertExpectations(t)
}

// Test ListDrafts - Success
func TestOportunidadeMEIHandler_ListDrafts_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockOportunidadeMEIService)
	handler := v1.NewOportunidadeMEIHandler(mockService)

	r := gin.New()
	r.GET("/api/v1/oportunidades-mei/drafts", handler.ListDrafts)

	drafts := []*models.OportunidadeMEI{
		{ID: 1, Titulo: "Draft 1", Status: models.StatusOportunidadeDraft},
	}

	mockService.On("ListDrafts", mock.Anything, "", "", 1, 10).Return(drafts, 1, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/oportunidades-mei/drafts", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	mockService.AssertExpectations(t)
}

// Test ListDrafts - With Filters
func TestOportunidadeMEIHandler_ListDrafts_WithFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockOportunidadeMEIService)
	handler := v1.NewOportunidadeMEIHandler(mockService)

	r := gin.New()
	r.GET("/api/v1/oportunidades-mei/drafts", handler.ListDrafts)

	drafts := []*models.OportunidadeMEI{
		{ID: 1, Titulo: "Draft Test", OrgaoID: "org123", Status: models.StatusOportunidadeDraft},
	}

	mockService.On("ListDrafts", mock.Anything, "org123", "Test", 1, 10).Return(drafts, 1, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/oportunidades-mei/drafts?orgaoId=org123&titulo=Test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

// Test ListDrafts - Service Error
func TestOportunidadeMEIHandler_ListDrafts_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockOportunidadeMEIService)
	handler := v1.NewOportunidadeMEIHandler(mockService)

	r := gin.New()
	r.GET("/api/v1/oportunidades-mei/drafts", handler.ListDrafts)

	mockService.On("ListDrafts", mock.Anything, "", "", 1, 10).Return(nil, 0, errors.New("database error"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/oportunidades-mei/drafts", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockService.AssertExpectations(t)
}

// Test CreateDraft - Service Error
func TestOportunidadeMEIHandler_CreateDraft_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockOportunidadeMEIService)
	handler := v1.NewOportunidadeMEIHandler(mockService)

	r := gin.New()
	r.POST("/api/v1/oportunidades-mei/draft", handler.CreateDraft)

	oportunidade := models.OportunidadeMEI{
		Titulo: "Test Draft",
	}

	mockService.On("Create", mock.Anything, mock.Anything, true).Return(0, errors.New("validation failed: missing required fields"))

	body, _ := json.Marshal(oportunidade)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/oportunidades-mei/draft", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "validation failed")

	mockService.AssertExpectations(t)
}

// Test List - With OrgaoID Filter
func TestOportunidadeMEIHandler_List_WithOrgaoIDFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockOportunidadeMEIService)
	handler := v1.NewOportunidadeMEIHandler(mockService)

	r := gin.New()
	r.GET("/api/v1/oportunidades-mei", handler.List)

	oportunidades := []*models.OportunidadeMEI{
		{ID: 1, Titulo: "Oportunidade Org", OrgaoID: "org123", Status: models.StatusOportunidadeActive},
	}

	mockService.On("List", mock.Anything, mock.MatchedBy(func(filters map[string]interface{}) bool {
		return filters["orgao_id"] == "org123" && filters["status"] == models.StatusOportunidadeActive
	}), "", 1, 10).Return(oportunidades, 1, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/oportunidades-mei?orgaoId=org123", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

// Test List - With Titulo Search
func TestOportunidadeMEIHandler_List_WithTituloSearch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockOportunidadeMEIService)
	handler := v1.NewOportunidadeMEIHandler(mockService)

	r := gin.New()
	r.GET("/api/v1/oportunidades-mei", handler.List)

	oportunidades := []*models.OportunidadeMEI{
		{ID: 1, Titulo: "Desenvolvedor", Status: models.StatusOportunidadeActive},
	}

	mockService.On("List", mock.Anything, mock.MatchedBy(func(filters map[string]interface{}) bool {
		return filters["status"] == models.StatusOportunidadeActive
	}), "Desenvolvedor", 1, 10).Return(oportunidades, 1, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/oportunidades-mei?titulo=Desenvolvedor", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

// Test List - With Expired Status
func TestOportunidadeMEIHandler_List_WithExpiredStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockOportunidadeMEIService)
	handler := v1.NewOportunidadeMEIHandler(mockService)

	r := gin.New()
	r.GET("/api/v1/oportunidades-mei", handler.List)

	oportunidades := []*models.OportunidadeMEI{
		{ID: 1, Titulo: "Expired Oportunidade", Status: models.StatusOportunidadeExpired},
	}

	mockService.On("List", mock.Anything, mock.MatchedBy(func(filters map[string]interface{}) bool {
		return filters["status"] == models.StatusOportunidadeExpired
	}), "", 1, 10).Return(oportunidades, 1, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/oportunidades-mei?status=expired", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

// Test List - Invalid Pagination Values
func TestOportunidadeMEIHandler_List_InvalidPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockOportunidadeMEIService)
	handler := v1.NewOportunidadeMEIHandler(mockService)

	r := gin.New()
	r.GET("/api/v1/oportunidades-mei", handler.List)

	oportunidades := []*models.OportunidadeMEI{}

	// Should normalize page to 1 and pageSize to 10
	mockService.On("List", mock.Anything, mock.Anything, "", 1, 10).Return(oportunidades, 0, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/oportunidades-mei?page=-5&pageSize=5000", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

// Test List - Empty Result Set
func TestOportunidadeMEIHandler_List_EmptyResult(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockOportunidadeMEIService)
	handler := v1.NewOportunidadeMEIHandler(mockService)

	r := gin.New()
	r.GET("/api/v1/oportunidades-mei", handler.List)

	oportunidades := []*models.OportunidadeMEI{}

	mockService.On("List", mock.Anything, mock.Anything, "", 1, 10).Return(oportunidades, 0, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/oportunidades-mei", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 0, int(response["meta"].(map[string]interface{})["total"].(float64)))

	mockService.AssertExpectations(t)
}

// Test ListDrafts - Invalid Pagination
func TestOportunidadeMEIHandler_ListDrafts_InvalidPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockOportunidadeMEIService)
	handler := v1.NewOportunidadeMEIHandler(mockService)

	r := gin.New()
	r.GET("/api/v1/oportunidades-mei/drafts", handler.ListDrafts)

	drafts := []*models.OportunidadeMEI{}

	// Should normalize page to 1 and pageSize to 10
	mockService.On("ListDrafts", mock.Anything, "", "", 1, 10).Return(drafts, 0, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/oportunidades-mei/drafts?page=0&pageSize=2000", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

// Test Publish - Not Found After Successful Publish
func TestOportunidadeMEIHandler_Publish_NotFoundAfterPublish(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockOportunidadeMEIService)
	handler := v1.NewOportunidadeMEIHandler(mockService)

	r := gin.New()
	r.PUT("/api/v1/oportunidades-mei/:id/publish", handler.Publish)

	mockService.On("Publish", mock.Anything, 1).Return(nil)
	mockService.On("GetByID", mock.Anything, 1).Return(nil, nil)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/oportunidades-mei/1/publish", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// This is an edge case - publish succeeded but GetByID returns nil
	// Handler returns 200 with nil body (JSON serialization will show null)
	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

