package v1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	v1 "github.com/prefeitura-rio/app-go-api/internal/handlers/v1"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

type mockEmpresaRepoForHandler struct {
	createID  int
	createErr error
	entity    *models.Empresa
	getErr    error
	updateErr error
	deleteErr error
	listItems []*models.Empresa
	listTotal int
	listErr   error
}

func (m *mockEmpresaRepoForHandler) Create(_ context.Context, e *models.Empresa) (int, error) {
	if m.createErr != nil {
		return 0, m.createErr
	}
	if m.createID == 0 {
		m.createID = 1
	}
	return m.createID, nil
}

func (m *mockEmpresaRepoForHandler) GetByID(_ context.Context, id int) (*models.Empresa, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.entity, nil
}

func (m *mockEmpresaRepoForHandler) Update(_ context.Context, e *models.Empresa) error {
	return m.updateErr
}

func (m *mockEmpresaRepoForHandler) Delete(_ context.Context, id int) error {
	return m.deleteErr
}

func (m *mockEmpresaRepoForHandler) List(_ context.Context, _ map[string]interface{}, _, _ int) ([]*models.Empresa, int, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	if m.listItems == nil {
		return []*models.Empresa{}, 0, nil
	}
	return m.listItems, m.listTotal, nil
}

func setupEmpresaRouter(repo services.EmpresaRepositoryInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	svc := services.NewEmpresaServiceWithInterface(repo)
	h := v1.NewEmpresaHandler(svc)
	r.POST("/api/v1/empresas", h.Create)
	r.GET("/api/v1/empresas", h.List)
	r.GET("/api/v1/empresas/:id", h.GetByID)
	r.PUT("/api/v1/empresas/:id", h.Update)
	r.DELETE("/api/v1/empresas/:id", h.Delete)
	return r
}

// Test handler creation
func TestNewEmpresaHandler(t *testing.T) {
	svc := services.NewEmpresaServiceWithInterface(&mockEmpresaRepoForHandler{})
	handler := v1.NewEmpresaHandler(svc)
	if handler == nil {
		t.Error("NewEmpresaHandler: expected non-nil handler")
	}
}

// Test Create - Success
func TestEmpresaHandler_Create_Success(t *testing.T) {
	repo := &mockEmpresaRepoForHandler{createID: 1}
	router := setupEmpresaRouter(repo)
	body := []byte(`{"nome":"Empresa XYZ"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/empresas", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("Create: expected 201, got %d", w.Code)
	}
}

// Test Create - Invalid JSON
func TestEmpresaHandler_Create_InvalidJSON(t *testing.T) {
	repo := &mockEmpresaRepoForHandler{}
	router := setupEmpresaRouter(repo)
	body := []byte(`{invalid json}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/empresas", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Create InvalidJSON: expected 400, got %d", w.Code)
	}
}

// Test Create - Service Error
func TestEmpresaHandler_Create_ServiceError(t *testing.T) {
	repo := &mockEmpresaRepoForHandler{createErr: errors.New("database error")}
	router := setupEmpresaRouter(repo)
	body := []byte(`{"nome":"Test"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/empresas", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Create ServiceError: expected 500, got %d", w.Code)
	}
}

// Test GetByID - Success
func TestEmpresaHandler_GetByID_Success(t *testing.T) {
	repo := &mockEmpresaRepoForHandler{entity: &models.Empresa{ID: 1, Nome: "Test Empresa"}}
	router := setupEmpresaRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/empresas/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("GetByID: expected 200, got %d", w.Code)
	}
	var out models.Empresa
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("GetByID unmarshal: %v", err)
	}
	if out.Nome != "Test Empresa" {
		t.Errorf("GetByID: expected Nome 'Test Empresa', got %s", out.Nome)
	}
}

// Test GetByID - Invalid ID
func TestEmpresaHandler_GetByID_InvalidID(t *testing.T) {
	repo := &mockEmpresaRepoForHandler{}
	router := setupEmpresaRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/empresas/invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("GetByID InvalidID: expected 400, got %d", w.Code)
	}
}

// Test GetByID - Not Found
func TestEmpresaHandler_GetByID_NotFound(t *testing.T) {
	repo := &mockEmpresaRepoForHandler{entity: nil}
	router := setupEmpresaRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/empresas/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("GetByID NotFound: expected 404, got %d", w.Code)
	}
}

// Test GetByID - Service Error
func TestEmpresaHandler_GetByID_ServiceError(t *testing.T) {
	repo := &mockEmpresaRepoForHandler{getErr: errors.New("database error")}
	router := setupEmpresaRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/empresas/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("GetByID ServiceError: expected 500, got %d", w.Code)
	}
}

// Test List - Success
func TestEmpresaHandler_List_Success(t *testing.T) {
	repo := &mockEmpresaRepoForHandler{
		listItems: []*models.Empresa{{ID: 1, Nome: "Empresa A"}},
		listTotal: 1,
	}
	router := setupEmpresaRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/empresas?page=1&pageSize=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("List: expected 200, got %d", w.Code)
	}
	var out struct {
		Data []*models.Empresa `json:"data"`
		Meta struct {
			Total int `json:"total"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("List unmarshal: %v", err)
	}
	if len(out.Data) != 1 || out.Meta.Total != 1 {
		t.Errorf("List: expected 1 item total 1, got %d %d", len(out.Data), out.Meta.Total)
	}
}

// Test List - Empty
func TestEmpresaHandler_List_Empty(t *testing.T) {
	repo := &mockEmpresaRepoForHandler{listItems: []*models.Empresa{}, listTotal: 0}
	router := setupEmpresaRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/empresas", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("List Empty: expected 200, got %d", w.Code)
	}
}

// Test List - Service Error
func TestEmpresaHandler_List_ServiceError(t *testing.T) {
	repo := &mockEmpresaRepoForHandler{listErr: errors.New("database error")}
	router := setupEmpresaRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/empresas", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("List ServiceError: expected 500, got %d", w.Code)
	}
}

// Test Update - Success
func TestEmpresaHandler_Update_Success(t *testing.T) {
	repo := &mockEmpresaRepoForHandler{entity: &models.Empresa{ID: 1, Nome: "Old Name"}}
	router := setupEmpresaRouter(repo)
	body := []byte(`{"nome":"New Name"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/empresas/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Update: expected 200, got %d", w.Code)
	}
}

// Test Update - Invalid ID
func TestEmpresaHandler_Update_InvalidID(t *testing.T) {
	repo := &mockEmpresaRepoForHandler{}
	router := setupEmpresaRouter(repo)
	body := []byte(`{"nome":"Test"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/empresas/invalid", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Update InvalidID: expected 400, got %d", w.Code)
	}
}

// Test Update - Invalid JSON
func TestEmpresaHandler_Update_InvalidJSON(t *testing.T) {
	repo := &mockEmpresaRepoForHandler{}
	router := setupEmpresaRouter(repo)
	body := []byte(`{invalid}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/empresas/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Update InvalidJSON: expected 400, got %d", w.Code)
	}
}

// Test Update - Service Error
func TestEmpresaHandler_Update_ServiceError(t *testing.T) {
	repo := &mockEmpresaRepoForHandler{
		entity:    &models.Empresa{ID: 1, Nome: "Test"},
		updateErr: errors.New("database error"),
	}
	router := setupEmpresaRouter(repo)
	body := []byte(`{"nome":"Updated"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/empresas/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Update ServiceError: expected 500, got %d", w.Code)
	}
}

// Test Delete - Success
func TestEmpresaHandler_Delete_Success(t *testing.T) {
	repo := &mockEmpresaRepoForHandler{}
	router := setupEmpresaRouter(repo)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/empresas/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Delete: expected 200, got %d", w.Code)
	}
}

// Test Delete - Invalid ID
func TestEmpresaHandler_Delete_InvalidID(t *testing.T) {
	repo := &mockEmpresaRepoForHandler{}
	router := setupEmpresaRouter(repo)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/empresas/invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Delete InvalidID: expected 400, got %d", w.Code)
	}
}

// Test Delete - Service Error
func TestEmpresaHandler_Delete_ServiceError(t *testing.T) {
	repo := &mockEmpresaRepoForHandler{deleteErr: errors.New("database error")}
	router := setupEmpresaRouter(repo)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/empresas/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Delete ServiceError: expected 500, got %d", w.Code)
	}
}
