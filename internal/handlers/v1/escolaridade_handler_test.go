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

type mockEscolaridadeRepoForHandler struct {
	createID  int
	createErr error
	entity    *models.Escolaridade
	getErr    error
	updateErr error
	deleteErr error
	listItems []*models.Escolaridade
	listTotal int
	listErr   error
}

func (m *mockEscolaridadeRepoForHandler) Create(_ context.Context, _ *models.Escolaridade) (int, error) {
	if m.createErr != nil {
		return 0, m.createErr
	}
	if m.createID == 0 {
		m.createID = 1
	}
	return m.createID, nil
}

func (m *mockEscolaridadeRepoForHandler) GetByID(_ context.Context, id int) (*models.Escolaridade, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.entity, nil
}

func (m *mockEscolaridadeRepoForHandler) Update(_ context.Context, _ *models.Escolaridade) error {
	return m.updateErr
}

func (m *mockEscolaridadeRepoForHandler) Delete(_ context.Context, _ int) error {
	return m.deleteErr
}

func (m *mockEscolaridadeRepoForHandler) List(_ context.Context, _ map[string]interface{}, _, _ int) ([]*models.Escolaridade, int, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	if m.listItems == nil {
		return []*models.Escolaridade{}, 0, nil
	}
	return m.listItems, m.listTotal, nil
}

func setupEscolaridadeRouter(repo services.EscolaridadeRepositoryInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	svc := services.NewEscolaridadeService(repo)
	h := v1.NewEscolaridadeHandler(svc)
	r.POST("/api/v1/escolaridades", h.Create)
	r.GET("/api/v1/escolaridades", h.List)
	r.GET("/api/v1/escolaridades/:id", h.GetByID)
	r.PUT("/api/v1/escolaridades/:id", h.Update)
	r.DELETE("/api/v1/escolaridades/:id", h.Delete)
	return r
}

func TestEscolaridadeHandler_CreateAndGet(t *testing.T) {
	repo := &mockEscolaridadeRepoForHandler{}
	router := setupEscolaridadeRouter(repo)
	body := []byte(`{"nivel":"Superior"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/escolaridades", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("Create: expected 201, got %d", w.Code)
	}

	repo.entity = &models.Escolaridade{ID: 1, Nivel: "Superior"}
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/escolaridades/1", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Errorf("GetByID: expected 200, got %d", w2.Code)
	}
	var out models.Escolaridade
	if err := json.Unmarshal(w2.Body.Bytes(), &out); err != nil {
		t.Fatalf("GetByID unmarshal: %v", err)
	}
	if out.Nivel != "Superior" {
		t.Errorf("GetByID: expected Nivel Superior, got %s", out.Nivel)
	}
}

func TestEscolaridadeHandler_List(t *testing.T) {
	repo := &mockEscolaridadeRepoForHandler{
		listItems: []*models.Escolaridade{{ID: 1, Nivel: "Médio"}},
		listTotal: 1,
	}
	router := setupEscolaridadeRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/escolaridades", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("List: expected 200, got %d", w.Code)
	}
}

// Test Update - Success
func TestEscolaridadeHandler_Update_Success(t *testing.T) {
	repo := &mockEscolaridadeRepoForHandler{entity: &models.Escolaridade{ID: 1, Nivel: "Old"}}
	router := setupEscolaridadeRouter(repo)
	body := []byte(`{"nivel":"New"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/escolaridades/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Update: expected 200, got %d", w.Code)
	}
}

// Test Update - Invalid ID
func TestEscolaridadeHandler_Update_InvalidID(t *testing.T) {
	repo := &mockEscolaridadeRepoForHandler{}
	router := setupEscolaridadeRouter(repo)
	body := []byte(`{"nivel":"Test"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/escolaridades/invalid", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Update InvalidID: expected 400, got %d", w.Code)
	}
}

// Test Update - Invalid JSON
func TestEscolaridadeHandler_Update_InvalidJSON(t *testing.T) {
	repo := &mockEscolaridadeRepoForHandler{}
	router := setupEscolaridadeRouter(repo)
	body := []byte(`{invalid}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/escolaridades/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Update InvalidJSON: expected 400, got %d", w.Code)
	}
}

// Test Update - Service Error
func TestEscolaridadeHandler_Update_ServiceError(t *testing.T) {
	repo := &mockEscolaridadeRepoForHandler{
		entity:    &models.Escolaridade{ID: 1, Nivel: "Test"},
		updateErr: errors.New("database error"),
	}
	router := setupEscolaridadeRouter(repo)
	body := []byte(`{"nivel":"Updated"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/escolaridades/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Update ServiceError: expected 500, got %d", w.Code)
	}
}

// Test Delete - Success
func TestEscolaridadeHandler_Delete_Success(t *testing.T) {
	repo := &mockEscolaridadeRepoForHandler{}
	router := setupEscolaridadeRouter(repo)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/escolaridades/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Delete: expected 200, got %d", w.Code)
	}
}

// Test Delete - Invalid ID
func TestEscolaridadeHandler_Delete_InvalidID(t *testing.T) {
	repo := &mockEscolaridadeRepoForHandler{}
	router := setupEscolaridadeRouter(repo)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/escolaridades/invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Delete InvalidID: expected 400, got %d", w.Code)
	}
}

// Test Delete - Service Error
func TestEscolaridadeHandler_Delete_ServiceError(t *testing.T) {
	repo := &mockEscolaridadeRepoForHandler{deleteErr: errors.New("database error")}
	router := setupEscolaridadeRouter(repo)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/escolaridades/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Delete ServiceError: expected 500, got %d", w.Code)
	}
}

// Test GetByID - Not Found
func TestEscolaridadeHandler_GetByID_NotFound(t *testing.T) {
	repo := &mockEscolaridadeRepoForHandler{entity: nil}
	router := setupEscolaridadeRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/escolaridades/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("GetByID NotFound: expected 404, got %d", w.Code)
	}
}

// Test GetByID - Invalid ID
func TestEscolaridadeHandler_GetByID_InvalidID(t *testing.T) {
	repo := &mockEscolaridadeRepoForHandler{}
	router := setupEscolaridadeRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/escolaridades/invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("GetByID InvalidID: expected 400, got %d", w.Code)
	}
}

// Test List - Service Error
func TestEscolaridadeHandler_List_ServiceError(t *testing.T) {
	repo := &mockEscolaridadeRepoForHandler{listErr: errors.New("database error")}
	router := setupEscolaridadeRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/escolaridades", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("List ServiceError: expected 500, got %d", w.Code)
	}
}

// Test Create - Invalid JSON
func TestEscolaridadeHandler_Create_InvalidJSON(t *testing.T) {
	repo := &mockEscolaridadeRepoForHandler{}
	router := setupEscolaridadeRouter(repo)
	body := []byte(`{invalid}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/escolaridades", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Create InvalidJSON: expected 400, got %d", w.Code)
	}
}

// Test Create - Service Error
func TestEscolaridadeHandler_Create_ServiceError(t *testing.T) {
	repo := &mockEscolaridadeRepoForHandler{createErr: errors.New("database error")}
	router := setupEscolaridadeRouter(repo)
	body := []byte(`{"nivel":"Test"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/escolaridades", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Create ServiceError: expected 500, got %d", w.Code)
	}
}
