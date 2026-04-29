package v1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	v1 "github.com/prefeitura-rio/app-go-api/internal/handlers/v1"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

type mockAcessibilidadeRepoForHandler struct {
	createID  int
	createErr error
	entity    *models.Acessibilidade
	getErr    error
	updateErr error
	deleteErr error
	listItems []*models.Acessibilidade
	listTotal int
	listErr   error
}

func (m *mockAcessibilidadeRepoForHandler) Create(_ context.Context, _ *models.Acessibilidade) (int, error) {
	if m.createErr != nil {
		return 0, m.createErr
	}
	if m.createID == 0 {
		m.createID = 1
	}
	return m.createID, nil
}

func (m *mockAcessibilidadeRepoForHandler) GetByID(_ context.Context, id int) (*models.Acessibilidade, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.entity, nil
}

func (m *mockAcessibilidadeRepoForHandler) Update(_ context.Context, _ *models.Acessibilidade) error {
	return m.updateErr
}

func (m *mockAcessibilidadeRepoForHandler) Delete(_ context.Context, _ int) error {
	return m.deleteErr
}

func (m *mockAcessibilidadeRepoForHandler) List(_ context.Context, _ map[string]interface{}, _, _ int) ([]*models.Acessibilidade, int, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	if m.listItems == nil {
		return []*models.Acessibilidade{}, 0, nil
	}
	return m.listItems, m.listTotal, nil
}

func setupAcessibilidadeRouter(repo services.AcessibilidadeRepositoryInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	svc := services.NewAcessibilidadeService(repo)
	h := v1.NewAcessibilidadeHandler(svc)
	r.POST("/api/v1/acessibilidades", h.Create)
	r.GET("/api/v1/acessibilidades", h.List)
	r.GET("/api/v1/acessibilidades/:id", h.GetByID)
	r.PUT("/api/v1/acessibilidades/:id", h.Update)
	r.DELETE("/api/v1/acessibilidades/:id", h.Delete)
	return r
}

func TestAcessibilidadeHandler_CreateAndGet(t *testing.T) {
	repo := &mockAcessibilidadeRepoForHandler{}
	router := setupAcessibilidadeRouter(repo)
	body := []byte(`{"nome":"Rampa"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/acessibilidades", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("Create: expected 201, got %d", w.Code)
	}

	repo.entity = &models.Acessibilidade{ID: 1, Nome: "Rampa"}
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/acessibilidades/1", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Errorf("GetByID: expected 200, got %d", w2.Code)
	}
	var out models.Acessibilidade
	if err := json.Unmarshal(w2.Body.Bytes(), &out); err != nil {
		t.Fatalf("GetByID unmarshal: %v", err)
	}
	if out.Nome != "Rampa" {
		t.Errorf("GetByID: expected Nome Rampa, got %s", out.Nome)
	}
}

func TestAcessibilidadeHandler_List(t *testing.T) {
	repo := &mockAcessibilidadeRepoForHandler{
		listItems: []*models.Acessibilidade{{ID: 1, Nome: "A"}},
		listTotal: 1,
	}
	router := setupAcessibilidadeRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/acessibilidades", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("List: expected 200, got %d", w.Code)
	}
}

// Test Update - Success
func TestAcessibilidadeHandler_Update_Success(t *testing.T) {
	repo := &mockAcessibilidadeRepoForHandler{}
	repo.entity = &models.Acessibilidade{ID: 1, Nome: "Old Name"}
	router := setupAcessibilidadeRouter(repo)
	body := []byte(`{"nome":"New Name"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/acessibilidades/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Update: expected 200, got %d", w.Code)
	}
}

// Test Update - Invalid ID
func TestAcessibilidadeHandler_Update_InvalidID(t *testing.T) {
	repo := &mockAcessibilidadeRepoForHandler{}
	router := setupAcessibilidadeRouter(repo)
	body := []byte(`{"nome":"Test"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/acessibilidades/invalid", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Update InvalidID: expected 400, got %d", w.Code)
	}
}

// Test Update - Invalid JSON
func TestAcessibilidadeHandler_Update_InvalidJSON(t *testing.T) {
	repo := &mockAcessibilidadeRepoForHandler{}
	router := setupAcessibilidadeRouter(repo)
	body := []byte(`{invalid}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/acessibilidades/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Update InvalidJSON: expected 400, got %d", w.Code)
	}
}

// Test Delete - Success
func TestAcessibilidadeHandler_Delete_Success(t *testing.T) {
	repo := &mockAcessibilidadeRepoForHandler{}
	router := setupAcessibilidadeRouter(repo)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/acessibilidades/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Delete: expected 200, got %d", w.Code)
	}
}

// Test Delete - Invalid ID
func TestAcessibilidadeHandler_Delete_InvalidID(t *testing.T) {
	repo := &mockAcessibilidadeRepoForHandler{}
	router := setupAcessibilidadeRouter(repo)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/acessibilidades/invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Delete InvalidID: expected 400, got %d", w.Code)
	}
}
