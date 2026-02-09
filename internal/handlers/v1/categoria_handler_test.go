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
	"github.com/prefeitura-rio/app-go-api/internal/handlers/v1"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

type mockCategoriaRepoForHandler struct {
	createID  int
	createErr error
	entity    *models.Categoria
	getErr    error
	updateErr error
	deleteErr error
	listItems []*models.Categoria
	listTotal int
	listErr   error
}

func (m *mockCategoriaRepoForHandler) Create(_ context.Context, c *models.Categoria) (int, error) {
	if m.createErr != nil {
		return 0, m.createErr
	}
	m.createID = 1
	return m.createID, nil
}

func (m *mockCategoriaRepoForHandler) GetByID(_ context.Context, id int) (*models.Categoria, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.entity, nil
}

func (m *mockCategoriaRepoForHandler) Update(_ context.Context, c *models.Categoria) error {
	return m.updateErr
}

func (m *mockCategoriaRepoForHandler) Delete(_ context.Context, id int) error {
	return m.deleteErr
}

func (m *mockCategoriaRepoForHandler) List(_ context.Context, _ map[string]interface{}, _, _ int) ([]*models.Categoria, int, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	if m.listItems == nil {
		return []*models.Categoria{}, 0, nil
	}
	return m.listItems, m.listTotal, nil
}

func setupCategoriaRouter(repo services.CategoriaRepositoryInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	svc := services.NewCategoriaService(repo)
	h := v1.NewCategoriaHandler(svc)
	r.POST("/api/v1/categorias", h.Create)
	r.GET("/api/v1/categorias", h.List)
	r.GET("/api/v1/categorias/:id", h.GetByID)
	r.PUT("/api/v1/categorias/:id", h.Update)
	r.DELETE("/api/v1/categorias/:id", h.Delete)
	return r
}

func TestCategoriaHandler_Create(t *testing.T) {
	repo := &mockCategoriaRepoForHandler{}
	router := setupCategoriaRouter(repo)
	body := []byte(`{"nome":"TI"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/categorias", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("Create: expected 201, got %d", w.Code)
	}
}

func TestCategoriaHandler_List_Default(t *testing.T) {
	repo := &mockCategoriaRepoForHandler{
		listItems: []*models.Categoria{{ID: 1, Nome: "A"}},
		listTotal: 1,
	}
	router := setupCategoriaRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/categorias?page=1&pageSize=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("List: expected 200, got %d", w.Code)
	}
	var out struct {
		Data []*models.Categoria `json:"data"`
		Meta struct {
			Total int `json:"total"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("List: unmarshal: %v", err)
	}
	if len(out.Data) != 1 || out.Meta.Total != 1 {
		t.Errorf("List: expected 1 item total 1, got %d %d", len(out.Data), out.Meta.Total)
	}
}

func TestCategoriaHandler_List_OnlyWithCourses(t *testing.T) {
	repo := &mockCategoriaRepoForHandler{listTotal: 0}
	router := setupCategoriaRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/categorias?onlyWithCourses=true&daysTolerance=15&isVisible=true", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("List onlyWithCourses: expected 200, got %d", w.Code)
	}
}

func TestCategoriaHandler_GetByID_Success(t *testing.T) {
	repo := &mockCategoriaRepoForHandler{entity: &models.Categoria{ID: 1, Nome: "TI"}}
	router := setupCategoriaRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/categorias/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("GetByID: expected 200, got %d", w.Code)
	}
}

func TestCategoriaHandler_GetByID_NotFound(t *testing.T) {
	repo := &mockCategoriaRepoForHandler{entity: nil}
	router := setupCategoriaRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/categorias/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("GetByID not found: expected 404, got %d", w.Code)
	}
}

func TestCategoriaHandler_List_ServiceError(t *testing.T) {
	repo := &mockCategoriaRepoForHandler{listErr: errors.New("db error")}
	router := setupCategoriaRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/categorias", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("List service error: expected 500, got %d", w.Code)
	}
}
