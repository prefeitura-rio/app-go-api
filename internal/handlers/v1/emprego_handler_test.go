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
	"github.com/prefeitura-rio/app-go-api/internal/services"
	"github.com/stretchr/testify/assert"
)

type mockEmpregoRepoForHandler struct {
	createID  int
	createErr error
	entity    *models.Emprego
	getErr    error
	updateErr error
	deleteErr error
	listItems []*models.Emprego
	listTotal int
	listErr   error
}

func (m *mockEmpregoRepoForHandler) Create(_ context.Context, e *models.Emprego) (int, error) {
	if m.createErr != nil {
		return 0, m.createErr
	}
	m.createID = 1
	return m.createID, nil
}

func (m *mockEmpregoRepoForHandler) GetByID(_ context.Context, id int) (*models.Emprego, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.entity, nil
}

func (m *mockEmpregoRepoForHandler) Update(_ context.Context, e *models.Emprego) error {
	return m.updateErr
}

func (m *mockEmpregoRepoForHandler) Delete(_ context.Context, id int) error {
	return m.deleteErr
}

func (m *mockEmpregoRepoForHandler) List(_ context.Context, _ map[string]interface{}, _, _ int) ([]*models.Emprego, int, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	if m.listItems == nil {
		return []*models.Emprego{}, 0, nil
	}
	return m.listItems, m.listTotal, nil
}

func setupEmpregoRouter(repo services.EmpregoRepositoryInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	svc := services.NewEmpregoServiceWithInterface(repo)
	h := v1.NewEmpregoHandler(svc, nil, nil)
	r.POST("/api/v1/empregos", h.Create)
	r.GET("/api/v1/empregos", h.List)
	r.GET("/api/v1/empregos/:id", h.GetByID)
	r.PUT("/api/v1/empregos/:id", h.Update)
	r.DELETE("/api/v1/empregos/:id", h.Delete)
	return r
}

func TestEmpregoHandler_Create_Success(t *testing.T) {
	repo := &mockEmpregoRepoForHandler{}
	router := setupEmpregoRouter(repo)

	body := []byte(`{
		"titulo": "Desenvolvedor Go",
		"descricao": "Vaga para desenvolvedor Go",
		"status": "ABERTO",
		"tipo_contratacao": "CLT"
	}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/empregos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestEmpregoHandler_Create_InvalidJSON(t *testing.T) {
	repo := &mockEmpregoRepoForHandler{}
	router := setupEmpregoRouter(repo)

	body := []byte(`{invalid json}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/empregos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Dados inválidos")
}

func TestEmpregoHandler_Create_ValidationError(t *testing.T) {
	repo := &mockEmpregoRepoForHandler{
		createErr: errors.New("campo obrigatório"),
	}
	router := setupEmpregoRouter(repo)

	body := []byte(`{"titulo": ""}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/empregos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEmpregoHandler_GetByID_Success(t *testing.T) {
	repo := &mockEmpregoRepoForHandler{
		entity: &models.Emprego{ID: 1, Titulo: "Test"},
	}
	router := setupEmpregoRouter(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/empregos/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Test")
}

func TestEmpregoHandler_GetByID_InvalidID(t *testing.T) {
	repo := &mockEmpregoRepoForHandler{}
	router := setupEmpregoRouter(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/empregos/invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "ID inválido")
}

func TestEmpregoHandler_GetByID_NotFound(t *testing.T) {
	repo := &mockEmpregoRepoForHandler{
		entity: nil,
	}
	router := setupEmpregoRouter(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/empregos/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Emprego não encontrado")
}

func TestEmpregoHandler_Update_Success(t *testing.T) {
	repo := &mockEmpregoRepoForHandler{}
	router := setupEmpregoRouter(repo)

	body := []byte(`{
		"titulo": "Updated",
		"status": "ABERTO",
		"tipo_contratacao": "CLT"
	}`)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/empregos/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEmpregoHandler_Update_InvalidID(t *testing.T) {
	repo := &mockEmpregoRepoForHandler{}
	router := setupEmpregoRouter(repo)

	body := []byte(`{"titulo": "Updated"}`)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/empregos/invalid", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEmpregoHandler_Delete_Success(t *testing.T) {
	repo := &mockEmpregoRepoForHandler{
		entity: &models.Emprego{ID: 1},
	}
	router := setupEmpregoRouter(repo)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/empregos/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "excluído com sucesso")
}

func TestEmpregoHandler_Delete_NotFound(t *testing.T) {
	repo := &mockEmpregoRepoForHandler{
		entity: nil,
	}
	router := setupEmpregoRouter(repo)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/empregos/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestEmpregoHandler_List_Default(t *testing.T) {
	repo := &mockEmpregoRepoForHandler{
		listItems: []*models.Emprego{{ID: 1, Titulo: "Test"}},
		listTotal: 1,
	}
	router := setupEmpregoRouter(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/empregos?page=1&pageSize=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "data")
	assert.Contains(t, w.Body.String(), "meta")
}

func TestEmpregoHandler_List_WithFilters(t *testing.T) {
	repo := &mockEmpregoRepoForHandler{
		listItems: []*models.Emprego{},
		listTotal: 0,
	}
	router := setupEmpregoRouter(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/empregos?status=aberta&orgao_id=123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEmpregoHandler_List_InvalidPagination(t *testing.T) {
	repo := &mockEmpregoRepoForHandler{
		listItems: []*models.Emprego{},
		listTotal: 0,
	}
	router := setupEmpregoRouter(repo)

	// Test negative page
	req := httptest.NewRequest(http.MethodGet, "/api/v1/empregos?page=-1&pageSize=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Test excessive pageSize
	req = httptest.NewRequest(http.MethodGet, "/api/v1/empregos?page=1&pageSize=2000", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNewEmpregoHandler(t *testing.T) {
	repo := &mockEmpregoRepoForHandler{}
	svc := services.NewEmpregoServiceWithInterface(repo)
	handler := v1.NewEmpregoHandler(svc, nil, nil)

	assert.NotNil(t, handler)
}

// Test Delete - GetByID error
func TestEmpregoHandler_Delete_GetByIDError(t *testing.T) {
	repo := &mockEmpregoRepoForHandler{
		getErr: errors.New("database connection error"),
	}
	router := setupEmpregoRouter(repo)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/empregos/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Erro ao buscar emprego")
}

// Test Delete - Delete service error
func TestEmpregoHandler_Delete_DeleteError(t *testing.T) {
	repo := &mockEmpregoRepoForHandler{
		entity:    &models.Emprego{ID: 1},
		deleteErr: errors.New("foreign key constraint"),
	}
	router := setupEmpregoRouter(repo)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/empregos/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Erro ao excluir emprego")
}

// Test Delete - Invalid ID
func TestEmpregoHandler_Delete_InvalidID(t *testing.T) {
	repo := &mockEmpregoRepoForHandler{}
	router := setupEmpregoRouter(repo)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/empregos/abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "ID inválido")
}

// Test Create - Database error (not validation)
func TestEmpregoHandler_Create_DatabaseError(t *testing.T) {
	repo := &mockEmpregoRepoForHandler{
		createErr: errors.New("database connection failed"),
	}
	router := setupEmpregoRouter(repo)

	body := []byte(`{
		"titulo": "Desenvolvedor",
		"status": "ABERTO",
		"tipo_contratacao": "CLT"
	}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/empregos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Database errors go through ParseDatabaseError which returns appropriate status
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusInternalServerError)
}

// Test Update - Validation error
func TestEmpregoHandler_Update_ValidationError(t *testing.T) {
	repo := &mockEmpregoRepoForHandler{
		updateErr: errors.New("campo obrigatório"),
	}
	router := setupEmpregoRouter(repo)

	body := []byte(`{"titulo": ""}`)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/empregos/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test Update - Database error
func TestEmpregoHandler_Update_DatabaseError(t *testing.T) {
	repo := &mockEmpregoRepoForHandler{
		updateErr: errors.New("constraint violation"),
	}
	router := setupEmpregoRouter(repo)

	body := []byte(`{"titulo": "Updated"}`)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/empregos/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test Update - Invalid JSON
func TestEmpregoHandler_Update_InvalidJSON(t *testing.T) {
	repo := &mockEmpregoRepoForHandler{}
	router := setupEmpregoRouter(repo)

	body := []byte(`{invalid}`)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/empregos/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test GetByID - Service error
func TestEmpregoHandler_GetByID_ServiceError(t *testing.T) {
	repo := &mockEmpregoRepoForHandler{
		getErr: errors.New("database timeout"),
	}
	router := setupEmpregoRouter(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/empregos/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Erro ao buscar emprego")
}

// Test List - Service error
func TestEmpregoHandler_List_ServiceError(t *testing.T) {
	repo := &mockEmpregoRepoForHandler{
		listErr: errors.New("database error"),
	}
	router := setupEmpregoRouter(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/empregos?page=1&pageSize=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Erro ao listar empregos")
}

// Test List - empresa_id filter
func TestEmpregoHandler_List_WithEmpresaIDFilter(t *testing.T) {
	repo := &mockEmpregoRepoForHandler{
		listItems: []*models.Emprego{{ID: 1, Titulo: "Test"}},
		listTotal: 1,
	}
	router := setupEmpregoRouter(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/empregos?empresa_id=42", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Test List - escolaridade_id filter
func TestEmpregoHandler_List_WithEscolaridadeIDFilter(t *testing.T) {
	repo := &mockEmpregoRepoForHandler{
		listItems: []*models.Emprego{{ID: 1, Titulo: "Test"}},
		listTotal: 1,
	}
	router := setupEmpregoRouter(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/empregos?escolaridade_id=5", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Test List - tipo_contratacao filter
func TestEmpregoHandler_List_WithTipoContratacaoFilter(t *testing.T) {
	repo := &mockEmpregoRepoForHandler{
		listItems: []*models.Emprego{{ID: 1, Titulo: "Test"}},
		listTotal: 1,
	}
	router := setupEmpregoRouter(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/empregos?tipo_contratacao=CLT", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Test List - jornada_trabalho filter
func TestEmpregoHandler_List_WithJornadaTrabalhoFilter(t *testing.T) {
	repo := &mockEmpregoRepoForHandler{
		listItems: []*models.Emprego{{ID: 1, Titulo: "Test"}},
		listTotal: 1,
	}
	router := setupEmpregoRouter(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/empregos?jornada_trabalho=INTEGRAL", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Test List - multiple filters
func TestEmpregoHandler_List_WithMultipleFilters(t *testing.T) {
	repo := &mockEmpregoRepoForHandler{
		listItems: []*models.Emprego{},
		listTotal: 0,
	}
	router := setupEmpregoRouter(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/empregos?status=ABERTO&tipo_contratacao=CLT&empresa_id=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
