package empregabilidade_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	handlers "github.com/prefeitura-rio/app-go-api/internal/handlers/v1/empregabilidade"
	empmodels "github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
)

type mockZonaRepoH struct {
	entity    *empmodels.Zona
	listItems []*empmodels.Zona
	listTotal int
	err       error
}

func (m *mockZonaRepoH) Create(_ context.Context, _ *empmodels.Zona) (uuid.UUID, error) {
	if m.err != nil {
		return uuid.Nil, m.err
	}
	return uuid.New(), nil
}
func (m *mockZonaRepoH) GetByID(_ context.Context, _ uuid.UUID) (*empmodels.Zona, error) {
	return m.entity, m.err
}
func (m *mockZonaRepoH) Update(_ context.Context, _ *empmodels.Zona) error { return m.err }
func (m *mockZonaRepoH) Delete(_ context.Context, _ uuid.UUID) error       { return m.err }
func (m *mockZonaRepoH) List(_ context.Context, _ map[string]interface{}, _, _ int) ([]*empmodels.Zona, int, error) {
	return m.listItems, m.listTotal, m.err
}

func setupZonaRouter(repo services.ZonaRepositoryInterface) *gin.Engine {
	r := gin.New()
	svc := services.NewZonaServiceWithInterface(repo)
	h := handlers.NewZonaHandler(svc)
	r.POST("/zonas", h.Create)
	r.GET("/zonas", h.List)
	r.GET("/zonas/:id", h.GetByID)
	r.PUT("/zonas/:id", h.Update)
	r.DELETE("/zonas/:id", h.Delete)
	return r
}

func TestZonaHandler_Create_Success(t *testing.T) {
	repo := &mockZonaRepoH{}
	r := setupZonaRouter(repo)
	body := bytes.NewBufferString(`{"descricao":"Zona Norte"}`)
	req := httptest.NewRequest(http.MethodPost, "/zonas", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
}

func TestZonaHandler_Create_Error(t *testing.T) {
	repo := &mockZonaRepoH{err: fmt.Errorf("db error")}
	r := setupZonaRouter(repo)
	body := bytes.NewBufferString(`{"descricao":"Zona Norte"}`)
	req := httptest.NewRequest(http.MethodPost, "/zonas", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusCreated {
		t.Error("expected error status, got 201")
	}
}

func TestZonaHandler_Create_BadJSON(t *testing.T) {
	repo := &mockZonaRepoH{}
	r := setupZonaRouter(repo)
	body := bytes.NewBufferString(`invalid json`)
	req := httptest.NewRequest(http.MethodPost, "/zonas", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestZonaHandler_List(t *testing.T) {
	repo := &mockZonaRepoH{
		listItems: []*empmodels.Zona{{ID: uuid.New(), Descricao: "Centro"}},
		listTotal: 1,
	}
	r := setupZonaRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/zonas", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestZonaHandler_List_Error(t *testing.T) {
	repo := &mockZonaRepoH{err: fmt.Errorf("db error")}
	r := setupZonaRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/zonas", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestZonaHandler_GetByID_Found(t *testing.T) {
	id := uuid.New()
	repo := &mockZonaRepoH{entity: &empmodels.Zona{ID: id, Descricao: "Centro"}}
	r := setupZonaRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/zonas/"+id.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestZonaHandler_GetByID_NotFound(t *testing.T) {
	repo := &mockZonaRepoH{entity: nil}
	r := setupZonaRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/zonas/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestZonaHandler_GetByID_InvalidID(t *testing.T) {
	repo := &mockZonaRepoH{}
	r := setupZonaRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/zonas/bad-id", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestZonaHandler_GetByID_ServiceError(t *testing.T) {
	repo := &mockZonaRepoH{err: fmt.Errorf("db error")}
	r := setupZonaRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/zonas/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestZonaHandler_Update_Success(t *testing.T) {
	id := uuid.New()
	repo := &mockZonaRepoH{}
	r := setupZonaRouter(repo)
	body := bytes.NewBufferString(`{"descricao":"Zona Sul"}`)
	req := httptest.NewRequest(http.MethodPut, "/zonas/"+id.String(), body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestZonaHandler_Update_InvalidID(t *testing.T) {
	repo := &mockZonaRepoH{}
	r := setupZonaRouter(repo)
	body := bytes.NewBufferString(`{"descricao":"Zona Sul"}`)
	req := httptest.NewRequest(http.MethodPut, "/zonas/bad-id", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestZonaHandler_Update_BadJSON(t *testing.T) {
	repo := &mockZonaRepoH{}
	r := setupZonaRouter(repo)
	body := bytes.NewBufferString(`invalid json`)
	req := httptest.NewRequest(http.MethodPut, "/zonas/"+uuid.New().String(), body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestZonaHandler_Update_Error(t *testing.T) {
	repo := &mockZonaRepoH{err: fmt.Errorf("db error")}
	r := setupZonaRouter(repo)
	body := bytes.NewBufferString(`{"descricao":"Zona Sul"}`)
	req := httptest.NewRequest(http.MethodPut, "/zonas/"+uuid.New().String(), body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestZonaHandler_Delete_Success(t *testing.T) {
	repo := &mockZonaRepoH{}
	r := setupZonaRouter(repo)
	req := httptest.NewRequest(http.MethodDelete, "/zonas/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestZonaHandler_Delete_InvalidID(t *testing.T) {
	repo := &mockZonaRepoH{}
	r := setupZonaRouter(repo)
	req := httptest.NewRequest(http.MethodDelete, "/zonas/bad-id", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestZonaHandler_Delete_Error(t *testing.T) {
	repo := &mockZonaRepoH{err: fmt.Errorf("db error")}
	r := setupZonaRouter(repo)
	req := httptest.NewRequest(http.MethodDelete, "/zonas/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}
