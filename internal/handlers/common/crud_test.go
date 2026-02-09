package common_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prefeitura-rio/app-go-api/internal/handlers/common"
	"github.com/prefeitura-rio/app-go-api/internal/models"
)

type mockCRUDService struct {
	createID    int
	createErr   error
	entity      *models.Categoria
	getErr      error
	updateErr   error
	deleteErr   error
	listItems   []*models.Categoria
	listTotal   int
	listErr     error
}

func (m *mockCRUDService) Create(ctx context.Context, entity *models.Categoria) (int, error) {
	if m.createErr != nil {
		return 0, m.createErr
	}
	if m.createID == 0 {
		m.createID = 1
	}
	return m.createID, nil
}

func (m *mockCRUDService) GetByID(ctx context.Context, id int) (*models.Categoria, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.entity, nil
}

func (m *mockCRUDService) Update(ctx context.Context, entity *models.Categoria) error {
	return m.updateErr
}

func (m *mockCRUDService) Delete(ctx context.Context, id int) error {
	return m.deleteErr
}

func (m *mockCRUDService) List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*models.Categoria, int, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	if m.listItems == nil {
		m.listItems = []*models.Categoria{}
	}
	return m.listItems, m.listTotal, nil
}

func setupCRUDRouter(s common.CRUDService[models.Categoria]) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := common.NewCRUDHandler(s, "Categoria")
	r.POST("/categorias", h.Create)
	r.GET("/categorias", h.List)
	r.GET("/categorias/:id", h.GetByID)
	r.PUT("/categorias/:id", h.Update)
	r.DELETE("/categorias/:id", h.Delete)
	return r
}

func TestCRUDHandler_Create_Success(t *testing.T) {
	mock := &mockCRUDService{createID: 42}
	router := setupCRUDRouter(mock)
	body := []byte(`{"nome":"TI"}`)
	req := httptest.NewRequest(http.MethodPost, "/categorias", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Create: expected status 201, got %d", w.Code)
	}
	var out models.Categoria
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("Create: unmarshal response: %v", err)
	}
	if out.ID != 42 {
		t.Errorf("Create: expected id 42, got %d", out.ID)
	}
	if out.Nome != "TI" {
		t.Errorf("Create: expected nome TI, got %s", out.Nome)
	}
}

func TestCRUDHandler_Create_InvalidJSON(t *testing.T) {
	mock := &mockCRUDService{}
	router := setupCRUDRouter(mock)
	req := httptest.NewRequest(http.MethodPost, "/categorias", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Create invalid JSON: expected 400, got %d", w.Code)
	}
}

func TestCRUDHandler_Create_ServiceError(t *testing.T) {
	mock := &mockCRUDService{createErr: errors.New("db error")}
	router := setupCRUDRouter(mock)
	body := []byte(`{"nome":"TI"}`)
	req := httptest.NewRequest(http.MethodPost, "/categorias", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Create service error: expected 500, got %d", w.Code)
	}
}

func TestCRUDHandler_GetByID_Success(t *testing.T) {
	mock := &mockCRUDService{
		entity: &models.Categoria{ID: 1, Nome: "TI"},
	}
	router := setupCRUDRouter(mock)
	req := httptest.NewRequest(http.MethodGet, "/categorias/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetByID: expected 200, got %d", w.Code)
	}
	var out models.Categoria
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("GetByID: unmarshal: %v", err)
	}
	if out.ID != 1 || out.Nome != "TI" {
		t.Errorf("GetByID: unexpected body: id=%d nome=%s", out.ID, out.Nome)
	}
}

func TestCRUDHandler_GetByID_InvalidID(t *testing.T) {
	mock := &mockCRUDService{}
	router := setupCRUDRouter(mock)
	req := httptest.NewRequest(http.MethodGet, "/categorias/abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("GetByID invalid id: expected 400, got %d", w.Code)
	}
}

func TestCRUDHandler_GetByID_NotFound(t *testing.T) {
	mock := &mockCRUDService{entity: nil}
	router := setupCRUDRouter(mock)
	req := httptest.NewRequest(http.MethodGet, "/categorias/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("GetByID not found: expected 404, got %d", w.Code)
	}
}

func TestCRUDHandler_Update_Success(t *testing.T) {
	mock := &mockCRUDService{}
	router := setupCRUDRouter(mock)
	body := []byte(`{"id":1,"nome":"TI Atualizado"}`)
	req := httptest.NewRequest(http.MethodPut, "/categorias/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Update: expected 200, got %d", w.Code)
	}
}

func TestCRUDHandler_Update_InvalidID(t *testing.T) {
	mock := &mockCRUDService{}
	router := setupCRUDRouter(mock)
	body := []byte(`{"nome":"TI"}`)
	req := httptest.NewRequest(http.MethodPut, "/categorias/x", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Update invalid id: expected 400, got %d", w.Code)
	}
}

func TestCRUDHandler_Delete_Success(t *testing.T) {
	mock := &mockCRUDService{}
	router := setupCRUDRouter(mock)
	req := httptest.NewRequest(http.MethodDelete, "/categorias/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Delete: expected 200, got %d", w.Code)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("Delete: unmarshal: %v", err)
	}
	if out["message"] != "Categoria excluída com sucesso" {
		t.Errorf("Delete: unexpected message: %v", out["message"])
	}
}

func TestCRUDHandler_Delete_InvalidID(t *testing.T) {
	mock := &mockCRUDService{}
	router := setupCRUDRouter(mock)
	req := httptest.NewRequest(http.MethodDelete, "/categorias/invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Delete invalid id: expected 400, got %d", w.Code)
	}
}

func TestCRUDHandler_List_Success(t *testing.T) {
	mock := &mockCRUDService{
		listItems: []*models.Categoria{{ID: 1, Nome: "A"}, {ID: 2, Nome: "B"}},
		listTotal: 2,
	}
	router := setupCRUDRouter(mock)
	req := httptest.NewRequest(http.MethodGet, "/categorias?page=1&pageSize=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("List: expected 200, got %d", w.Code)
	}
	var out struct {
		Data []*models.Categoria `json:"data"`
		Meta struct {
			Page     int `json:"page"`
			PageSize int `json:"page_size"`
			Total    int `json:"total"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("List: unmarshal: %v", err)
	}
	if len(out.Data) != 2 || out.Meta.Total != 2 {
		t.Errorf("List: expected 2 items and total 2, got len=%d total=%d", len(out.Data), out.Meta.Total)
	}
}

func TestCRUDHandler_List_PaginationDefaults(t *testing.T) {
	mock := &mockCRUDService{listTotal: 0}
	router := setupCRUDRouter(mock)
	req := httptest.NewRequest(http.MethodGet, "/categorias", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("List no query: expected 200, got %d", w.Code)
	}
	var out struct {
		Meta struct {
			Page     int `json:"page"`
			PageSize int `json:"page_size"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("List: unmarshal: %v", err)
	}
	if out.Meta.Page != 1 || out.Meta.PageSize != 10 {
		t.Errorf("List: expected page=1 pageSize=10, got page=%d pageSize=%d", out.Meta.Page, out.Meta.PageSize)
	}
}
