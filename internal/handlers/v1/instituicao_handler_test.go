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

type mockInstituicaoRepoForHandler struct {
	createID  int
	createErr error
	entity    *models.InstituicaoEnsino
	getErr    error
	updateErr error
	deleteErr error
	listItems []*models.InstituicaoEnsino
	listTotal int
	listErr   error
}

func (m *mockInstituicaoRepoForHandler) Create(_ context.Context, i *models.InstituicaoEnsino) (int, error) {
	if m.createErr != nil {
		return 0, m.createErr
	}
	if m.createID == 0 {
		m.createID = 1
	}
	return m.createID, nil
}

func (m *mockInstituicaoRepoForHandler) GetByID(_ context.Context, id int) (*models.InstituicaoEnsino, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.entity, nil
}

func (m *mockInstituicaoRepoForHandler) Update(_ context.Context, i *models.InstituicaoEnsino) error {
	return m.updateErr
}

func (m *mockInstituicaoRepoForHandler) Delete(_ context.Context, id int) error {
	return m.deleteErr
}

func (m *mockInstituicaoRepoForHandler) List(_ context.Context, _ map[string]interface{}, _, _ int) ([]*models.InstituicaoEnsino, int, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	if m.listItems == nil {
		return []*models.InstituicaoEnsino{}, 0, nil
	}
	return m.listItems, m.listTotal, nil
}

func setupInstituicaoRouter(repo services.InstituicaoRepositoryInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	svc := services.NewInstituicaoServiceWithInterface(repo)
	h := v1.NewInstituicaoHandler(svc)
	r.POST("/api/v1/instituicoes", h.Create)
	r.GET("/api/v1/instituicoes", h.List)
	r.GET("/api/v1/instituicoes/:id", h.GetByID)
	r.PUT("/api/v1/instituicoes/:id", h.Update)
	r.DELETE("/api/v1/instituicoes/:id", h.Delete)
	return r
}

// Test handler creation
func TestNewInstituicaoHandler(t *testing.T) {
	svc := services.NewInstituicaoServiceWithInterface(&mockInstituicaoRepoForHandler{})
	handler := v1.NewInstituicaoHandler(svc)
	if handler == nil {
		t.Error("NewInstituicaoHandler: expected non-nil handler")
	}
}

// Test Create - Success
func TestInstituicaoHandler_Create_Success(t *testing.T) {
	repo := &mockInstituicaoRepoForHandler{createID: 1}
	router := setupInstituicaoRouter(repo)
	body := []byte(`{"nome":"UFRJ"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/instituicoes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("Create: expected 201, got %d", w.Code)
	}
}

// Test Create - Invalid JSON
func TestInstituicaoHandler_Create_InvalidJSON(t *testing.T) {
	repo := &mockInstituicaoRepoForHandler{}
	router := setupInstituicaoRouter(repo)
	body := []byte(`{invalid json}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/instituicoes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Create InvalidJSON: expected 400, got %d", w.Code)
	}
}


// Test Create - Service Error
func TestInstituicaoHandler_Create_ServiceError(t *testing.T) {
	repo := &mockInstituicaoRepoForHandler{createErr: errors.New("database error")}
	router := setupInstituicaoRouter(repo)
	body := []byte(`{"nome":"Test"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/instituicoes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Create ServiceError: expected 500, got %d", w.Code)
	}
}

// Test GetByID - Success
func TestInstituicaoHandler_GetByID_Success(t *testing.T) {
	repo := &mockInstituicaoRepoForHandler{entity: &models.InstituicaoEnsino{ID: 1, Nome: "UFRJ"}}
	router := setupInstituicaoRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/instituicoes/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("GetByID: expected 200, got %d", w.Code)
	}
	var out models.InstituicaoEnsino
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("GetByID unmarshal: %v", err)
	}
	if out.Nome != "UFRJ" {
		t.Errorf("GetByID: expected Nome 'UFRJ', got %s", out.Nome)
	}
}

// Test GetByID - Invalid ID
func TestInstituicaoHandler_GetByID_InvalidID(t *testing.T) {
	repo := &mockInstituicaoRepoForHandler{}
	router := setupInstituicaoRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/instituicoes/invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("GetByID InvalidID: expected 400, got %d", w.Code)
	}
}

// Test GetByID - Not Found
func TestInstituicaoHandler_GetByID_NotFound(t *testing.T) {
	repo := &mockInstituicaoRepoForHandler{entity: nil}
	router := setupInstituicaoRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/instituicoes/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("GetByID NotFound: expected 404, got %d", w.Code)
	}
}

// Test GetByID - Service Error
func TestInstituicaoHandler_GetByID_ServiceError(t *testing.T) {
	repo := &mockInstituicaoRepoForHandler{getErr: errors.New("database error")}
	router := setupInstituicaoRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/instituicoes/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("GetByID ServiceError: expected 500, got %d", w.Code)
	}
}

// Test List - Success
func TestInstituicaoHandler_List_Success(t *testing.T) {
	repo := &mockInstituicaoRepoForHandler{
		listItems: []*models.InstituicaoEnsino{{ID: 1, Nome: "UFRJ"}},
		listTotal: 1,
	}
	router := setupInstituicaoRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/instituicoes?page=1&pageSize=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("List: expected 200, got %d", w.Code)
	}
	var out struct {
		Data []*models.InstituicaoEnsino `json:"data"`
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
func TestInstituicaoHandler_List_Empty(t *testing.T) {
	repo := &mockInstituicaoRepoForHandler{listItems: []*models.InstituicaoEnsino{}, listTotal: 0}
	router := setupInstituicaoRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/instituicoes", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("List Empty: expected 200, got %d", w.Code)
	}
}

// Test List - Service Error
func TestInstituicaoHandler_List_ServiceError(t *testing.T) {
	repo := &mockInstituicaoRepoForHandler{listErr: errors.New("database error")}
	router := setupInstituicaoRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/instituicoes", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("List ServiceError: expected 500, got %d", w.Code)
	}
}

// Test Update - Success
func TestInstituicaoHandler_Update_Success(t *testing.T) {
	repo := &mockInstituicaoRepoForHandler{entity: &models.InstituicaoEnsino{ID: 1, Nome: "Old Name"}}
	router := setupInstituicaoRouter(repo)
	body := []byte(`{"nome":"New Name"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/instituicoes/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Update: expected 200, got %d", w.Code)
	}
}

// Test Update - Invalid ID
func TestInstituicaoHandler_Update_InvalidID(t *testing.T) {
	repo := &mockInstituicaoRepoForHandler{}
	router := setupInstituicaoRouter(repo)
	body := []byte(`{"nome":"Test"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/instituicoes/invalid", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Update InvalidID: expected 400, got %d", w.Code)
	}
}

// Test Update - Invalid JSON
func TestInstituicaoHandler_Update_InvalidJSON(t *testing.T) {
	repo := &mockInstituicaoRepoForHandler{}
	router := setupInstituicaoRouter(repo)
	body := []byte(`{invalid}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/instituicoes/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Update InvalidJSON: expected 400, got %d", w.Code)
	}
}


// Test Update - Service Error
func TestInstituicaoHandler_Update_ServiceError(t *testing.T) {
	repo := &mockInstituicaoRepoForHandler{
		entity:    &models.InstituicaoEnsino{ID: 1, Nome: "Test"},
		updateErr: errors.New("database error"),
	}
	router := setupInstituicaoRouter(repo)
	body := []byte(`{"nome":"Updated"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/instituicoes/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Update ServiceError: expected 500, got %d", w.Code)
	}
}

// Test Delete - Success
func TestInstituicaoHandler_Delete_Success(t *testing.T) {
	repo := &mockInstituicaoRepoForHandler{}
	router := setupInstituicaoRouter(repo)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/instituicoes/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Delete: expected 200, got %d", w.Code)
	}
}

// Test Delete - Invalid ID
func TestInstituicaoHandler_Delete_InvalidID(t *testing.T) {
	repo := &mockInstituicaoRepoForHandler{}
	router := setupInstituicaoRouter(repo)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/instituicoes/invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Delete InvalidID: expected 400, got %d", w.Code)
	}
}

// Test Delete - Service Error
func TestInstituicaoHandler_Delete_ServiceError(t *testing.T) {
	repo := &mockInstituicaoRepoForHandler{deleteErr: errors.New("database error")}
	router := setupInstituicaoRouter(repo)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/instituicoes/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Delete ServiceError: expected 500, got %d", w.Code)
	}
}
