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

func init() {
	gin.SetMode(gin.TestMode)
}

// ──────────────────────────────────────────────────────────────────────────────
// Mock repositories
// ──────────────────────────────────────────────────────────────────────────────

// --- RegimeContratacao ---

type mockRegimeContratacaoRepoH struct {
	entity    *empmodels.RegimeContratacao
	listItems []*empmodels.RegimeContratacao
	listTotal int
	err       error
}

func (m *mockRegimeContratacaoRepoH) Create(_ context.Context, e *empmodels.RegimeContratacao) (uuid.UUID, error) {
	if m.err != nil {
		return uuid.Nil, m.err
	}
	return uuid.New(), nil
}
func (m *mockRegimeContratacaoRepoH) GetByID(_ context.Context, _ uuid.UUID) (*empmodels.RegimeContratacao, error) {
	return m.entity, m.err
}
func (m *mockRegimeContratacaoRepoH) Update(_ context.Context, _ *empmodels.RegimeContratacao) error {
	return m.err
}
func (m *mockRegimeContratacaoRepoH) Delete(_ context.Context, _ uuid.UUID) error { return m.err }
func (m *mockRegimeContratacaoRepoH) List(_ context.Context, _ map[string]interface{}, _, _ int) ([]*empmodels.RegimeContratacao, int, error) {
	return m.listItems, m.listTotal, m.err
}

func setupRegimeContratacaoRouter(repo services.RegimeContratacaoRepositoryInterface) *gin.Engine {
	r := gin.New()
	svc := services.NewRegimeContratacaoServiceWithInterface(repo)
	h := handlers.NewRegimeContratacaoHandler(svc)
	r.POST("/regimes", h.Create)
	r.GET("/regimes", h.List)
	r.GET("/regimes/:id", h.GetByID)
	r.PUT("/regimes/:id", h.Update)
	r.DELETE("/regimes/:id", h.Delete)
	return r
}

func TestRegimeContratacaoHandler_Create_Success(t *testing.T) {
	repo := &mockRegimeContratacaoRepoH{}
	r := setupRegimeContratacaoRouter(repo)
	body := bytes.NewBufferString(`{"descricao":"CLT"}`)
	req := httptest.NewRequest(http.MethodPost, "/regimes", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
}

func TestRegimeContratacaoHandler_Create_Error(t *testing.T) {
	repo := &mockRegimeContratacaoRepoH{err: fmt.Errorf("db error")}
	r := setupRegimeContratacaoRouter(repo)
	body := bytes.NewBufferString(`{"descricao":"CLT"}`)
	req := httptest.NewRequest(http.MethodPost, "/regimes", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusCreated {
		t.Error("expected error status, got 201")
	}
}

func TestRegimeContratacaoHandler_Create_BadJSON(t *testing.T) {
	repo := &mockRegimeContratacaoRepoH{}
	r := setupRegimeContratacaoRouter(repo)
	body := bytes.NewBufferString(`invalid json`)
	req := httptest.NewRequest(http.MethodPost, "/regimes", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestRegimeContratacaoHandler_List(t *testing.T) {
	repo := &mockRegimeContratacaoRepoH{
		listItems: []*empmodels.RegimeContratacao{{ID: uuid.New(), Descricao: "CLT"}},
		listTotal: 1,
	}
	r := setupRegimeContratacaoRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/regimes", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestRegimeContratacaoHandler_List_Error(t *testing.T) {
	repo := &mockRegimeContratacaoRepoH{err: fmt.Errorf("db error")}
	r := setupRegimeContratacaoRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/regimes", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestRegimeContratacaoHandler_GetByID_Found(t *testing.T) {
	id := uuid.New()
	repo := &mockRegimeContratacaoRepoH{entity: &empmodels.RegimeContratacao{ID: id, Descricao: "CLT"}}
	r := setupRegimeContratacaoRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/regimes/"+id.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestRegimeContratacaoHandler_GetByID_NotFound(t *testing.T) {
	repo := &mockRegimeContratacaoRepoH{entity: nil}
	r := setupRegimeContratacaoRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/regimes/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestRegimeContratacaoHandler_GetByID_InvalidID(t *testing.T) {
	repo := &mockRegimeContratacaoRepoH{}
	r := setupRegimeContratacaoRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/regimes/not-a-uuid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestRegimeContratacaoHandler_GetByID_ServiceError(t *testing.T) {
	repo := &mockRegimeContratacaoRepoH{err: fmt.Errorf("db error")}
	r := setupRegimeContratacaoRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/regimes/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestRegimeContratacaoHandler_Update_Success(t *testing.T) {
	repo := &mockRegimeContratacaoRepoH{}
	r := setupRegimeContratacaoRouter(repo)
	body := bytes.NewBufferString(`{"descricao":"PJ"}`)
	req := httptest.NewRequest(http.MethodPut, "/regimes/"+uuid.New().String(), body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestRegimeContratacaoHandler_Update_InvalidID(t *testing.T) {
	repo := &mockRegimeContratacaoRepoH{}
	r := setupRegimeContratacaoRouter(repo)
	body := bytes.NewBufferString(`{"descricao":"PJ"}`)
	req := httptest.NewRequest(http.MethodPut, "/regimes/bad-id", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestRegimeContratacaoHandler_Delete_Success(t *testing.T) {
	repo := &mockRegimeContratacaoRepoH{}
	r := setupRegimeContratacaoRouter(repo)
	req := httptest.NewRequest(http.MethodDelete, "/regimes/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestRegimeContratacaoHandler_Delete_InvalidID(t *testing.T) {
	repo := &mockRegimeContratacaoRepoH{}
	r := setupRegimeContratacaoRouter(repo)
	req := httptest.NewRequest(http.MethodDelete, "/regimes/bad-id", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// ModeloTrabalho handler tests
// ──────────────────────────────────────────────────────────────────────────────

type mockModeloTrabalhoRepoH struct {
	entity    *empmodels.ModeloTrabalho
	listItems []*empmodels.ModeloTrabalho
	listTotal int
	err       error
}

func (m *mockModeloTrabalhoRepoH) Create(_ context.Context, _ *empmodels.ModeloTrabalho) (uuid.UUID, error) {
	if m.err != nil {
		return uuid.Nil, m.err
	}
	return uuid.New(), nil
}
func (m *mockModeloTrabalhoRepoH) GetByID(_ context.Context, _ uuid.UUID) (*empmodels.ModeloTrabalho, error) {
	return m.entity, m.err
}
func (m *mockModeloTrabalhoRepoH) Update(_ context.Context, _ *empmodels.ModeloTrabalho) error {
	return m.err
}
func (m *mockModeloTrabalhoRepoH) Delete(_ context.Context, _ uuid.UUID) error { return m.err }
func (m *mockModeloTrabalhoRepoH) List(_ context.Context, _ map[string]interface{}, _, _ int) ([]*empmodels.ModeloTrabalho, int, error) {
	return m.listItems, m.listTotal, m.err
}

func setupModeloTrabalhoRouter(repo services.ModeloTrabalhoRepositoryInterface) *gin.Engine {
	r := gin.New()
	svc := services.NewModeloTrabalhoServiceWithInterface(repo)
	h := handlers.NewModeloTrabalhoHandler(svc)
	r.POST("/modelos", h.Create)
	r.GET("/modelos", h.List)
	r.GET("/modelos/:id", h.GetByID)
	r.PUT("/modelos/:id", h.Update)
	r.DELETE("/modelos/:id", h.Delete)
	return r
}

func TestModeloTrabalhoHandler_Create_Success(t *testing.T) {
	repo := &mockModeloTrabalhoRepoH{}
	r := setupModeloTrabalhoRouter(repo)
	body := bytes.NewBufferString(`{"descricao":"Remoto"}`)
	req := httptest.NewRequest(http.MethodPost, "/modelos", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
}

func TestModeloTrabalhoHandler_Create_Error(t *testing.T) {
	repo := &mockModeloTrabalhoRepoH{err: fmt.Errorf("db error")}
	r := setupModeloTrabalhoRouter(repo)
	body := bytes.NewBufferString(`{"descricao":"Remoto"}`)
	req := httptest.NewRequest(http.MethodPost, "/modelos", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusCreated {
		t.Error("expected error status")
	}
}

func TestModeloTrabalhoHandler_List(t *testing.T) {
	repo := &mockModeloTrabalhoRepoH{
		listItems: []*empmodels.ModeloTrabalho{{ID: uuid.New(), Descricao: "Remoto"}},
		listTotal: 1,
	}
	r := setupModeloTrabalhoRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/modelos", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestModeloTrabalhoHandler_List_Error(t *testing.T) {
	repo := &mockModeloTrabalhoRepoH{err: fmt.Errorf("db error")}
	r := setupModeloTrabalhoRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/modelos", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestModeloTrabalhoHandler_GetByID_Found(t *testing.T) {
	id := uuid.New()
	repo := &mockModeloTrabalhoRepoH{entity: &empmodels.ModeloTrabalho{ID: id, Descricao: "Remoto"}}
	r := setupModeloTrabalhoRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/modelos/"+id.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestModeloTrabalhoHandler_GetByID_NotFound(t *testing.T) {
	repo := &mockModeloTrabalhoRepoH{entity: nil}
	r := setupModeloTrabalhoRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/modelos/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestModeloTrabalhoHandler_GetByID_InvalidID(t *testing.T) {
	repo := &mockModeloTrabalhoRepoH{}
	r := setupModeloTrabalhoRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/modelos/bad-id", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestModeloTrabalhoHandler_GetByID_ServiceError(t *testing.T) {
	repo := &mockModeloTrabalhoRepoH{err: fmt.Errorf("db error")}
	r := setupModeloTrabalhoRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/modelos/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestModeloTrabalhoHandler_Create_BadJSON(t *testing.T) {
	repo := &mockModeloTrabalhoRepoH{}
	r := setupModeloTrabalhoRouter(repo)
	body := bytes.NewBufferString(`invalid json`)
	req := httptest.NewRequest(http.MethodPost, "/modelos", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestModeloTrabalhoHandler_Update_Success(t *testing.T) {
	repo := &mockModeloTrabalhoRepoH{}
	r := setupModeloTrabalhoRouter(repo)
	body := bytes.NewBufferString(`{"descricao":"Hibrido"}`)
	req := httptest.NewRequest(http.MethodPut, "/modelos/"+uuid.New().String(), body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestModeloTrabalhoHandler_Update_Error(t *testing.T) {
	repo := &mockModeloTrabalhoRepoH{err: fmt.Errorf("update error")}
	r := setupModeloTrabalhoRouter(repo)
	body := bytes.NewBufferString(`{"descricao":"Hibrido"}`)
	req := httptest.NewRequest(http.MethodPut, "/modelos/"+uuid.New().String(), body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestModeloTrabalhoHandler_Delete_Success(t *testing.T) {
	repo := &mockModeloTrabalhoRepoH{}
	r := setupModeloTrabalhoRouter(repo)
	req := httptest.NewRequest(http.MethodDelete, "/modelos/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestModeloTrabalhoHandler_Delete_InvalidID(t *testing.T) {
	repo := &mockModeloTrabalhoRepoH{}
	r := setupModeloTrabalhoRouter(repo)
	req := httptest.NewRequest(http.MethodDelete, "/modelos/bad-id", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// TipoPCD handler tests
// ──────────────────────────────────────────────────────────────────────────────

type mockTipoPCDRepoH struct {
	entity    *empmodels.TipoPCD
	listItems []*empmodels.TipoPCD
	listTotal int
	err       error
}

func (m *mockTipoPCDRepoH) Create(_ context.Context, _ *empmodels.TipoPCD) (uuid.UUID, error) {
	if m.err != nil {
		return uuid.Nil, m.err
	}
	return uuid.New(), nil
}
func (m *mockTipoPCDRepoH) GetByID(_ context.Context, _ uuid.UUID) (*empmodels.TipoPCD, error) {
	return m.entity, m.err
}
func (m *mockTipoPCDRepoH) Update(_ context.Context, _ *empmodels.TipoPCD) error { return m.err }
func (m *mockTipoPCDRepoH) Delete(_ context.Context, _ uuid.UUID) error          { return m.err }
func (m *mockTipoPCDRepoH) List(_ context.Context, _ map[string]interface{}, _, _ int) ([]*empmodels.TipoPCD, int, error) {
	return m.listItems, m.listTotal, m.err
}

func setupTipoPCDRouter(repo services.TipoPCDRepositoryInterface) *gin.Engine {
	r := gin.New()
	svc := services.NewTipoPCDServiceWithInterface(repo)
	h := handlers.NewTipoPCDHandler(svc)
	r.POST("/tipos-pcd", h.Create)
	r.GET("/tipos-pcd", h.List)
	r.GET("/tipos-pcd/:id", h.GetByID)
	r.PUT("/tipos-pcd/:id", h.Update)
	r.DELETE("/tipos-pcd/:id", h.Delete)
	return r
}

func TestTipoPCDHandler_CRUD(t *testing.T) {
	id := uuid.New()

	t.Run("Create_Success", func(t *testing.T) {
		repo := &mockTipoPCDRepoH{}
		r := setupTipoPCDRouter(repo)
		body := bytes.NewBufferString(`{"descricao":"Visual"}`)
		req := httptest.NewRequest(http.MethodPost, "/tipos-pcd", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", w.Code)
		}
	})

	t.Run("Create_BadJSON", func(t *testing.T) {
		repo := &mockTipoPCDRepoH{}
		r := setupTipoPCDRouter(repo)
		body := bytes.NewBufferString(`not json`)
		req := httptest.NewRequest(http.MethodPost, "/tipos-pcd", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("List_Success", func(t *testing.T) {
		repo := &mockTipoPCDRepoH{listItems: []*empmodels.TipoPCD{{ID: id}}, listTotal: 1}
		r := setupTipoPCDRouter(repo)
		req := httptest.NewRequest(http.MethodGet, "/tipos-pcd", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("GetByID_Found", func(t *testing.T) {
		repo := &mockTipoPCDRepoH{entity: &empmodels.TipoPCD{ID: id}}
		r := setupTipoPCDRouter(repo)
		req := httptest.NewRequest(http.MethodGet, "/tipos-pcd/"+id.String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("GetByID_NotFound", func(t *testing.T) {
		repo := &mockTipoPCDRepoH{entity: nil}
		r := setupTipoPCDRouter(repo)
		req := httptest.NewRequest(http.MethodGet, "/tipos-pcd/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})

	t.Run("GetByID_InvalidID", func(t *testing.T) {
		repo := &mockTipoPCDRepoH{}
		r := setupTipoPCDRouter(repo)
		req := httptest.NewRequest(http.MethodGet, "/tipos-pcd/bad", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("GetByID_ServiceError", func(t *testing.T) {
		repo := &mockTipoPCDRepoH{err: fmt.Errorf("db error")}
		r := setupTipoPCDRouter(repo)
		req := httptest.NewRequest(http.MethodGet, "/tipos-pcd/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == http.StatusOK {
			t.Error("expected error status")
		}
	})

	t.Run("List_Error", func(t *testing.T) {
		repo := &mockTipoPCDRepoH{err: fmt.Errorf("db error")}
		r := setupTipoPCDRouter(repo)
		req := httptest.NewRequest(http.MethodGet, "/tipos-pcd", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == http.StatusOK {
			t.Error("expected error status")
		}
	})

	t.Run("Create_Error", func(t *testing.T) {
		repo := &mockTipoPCDRepoH{err: fmt.Errorf("db error")}
		r := setupTipoPCDRouter(repo)
		body := bytes.NewBufferString(`{"descricao":"Visual"}`)
		req := httptest.NewRequest(http.MethodPost, "/tipos-pcd", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == http.StatusCreated {
			t.Error("expected error status")
		}
	})

	t.Run("Update_Success", func(t *testing.T) {
		repo := &mockTipoPCDRepoH{}
		r := setupTipoPCDRouter(repo)
		body := bytes.NewBufferString(`{"descricao":"Auditiva"}`)
		req := httptest.NewRequest(http.MethodPut, "/tipos-pcd/"+id.String(), body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("Update_InvalidID", func(t *testing.T) {
		repo := &mockTipoPCDRepoH{}
		r := setupTipoPCDRouter(repo)
		body := bytes.NewBufferString(`{"descricao":"Auditiva"}`)
		req := httptest.NewRequest(http.MethodPut, "/tipos-pcd/bad-id", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("Delete_Success", func(t *testing.T) {
		repo := &mockTipoPCDRepoH{}
		r := setupTipoPCDRouter(repo)
		req := httptest.NewRequest(http.MethodDelete, "/tipos-pcd/"+id.String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("Delete_InvalidID", func(t *testing.T) {
		repo := &mockTipoPCDRepoH{}
		r := setupTipoPCDRouter(repo)
		req := httptest.NewRequest(http.MethodDelete, "/tipos-pcd/bad", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// Idioma handler tests
// ──────────────────────────────────────────────────────────────────────────────

type mockIdiomaRepoH struct {
	entity    *empmodels.Idioma
	listItems []*empmodels.Idioma
	listTotal int
	err       error
}

func (m *mockIdiomaRepoH) Create(_ context.Context, _ *empmodels.Idioma) (uuid.UUID, error) {
	if m.err != nil {
		return uuid.Nil, m.err
	}
	return uuid.New(), nil
}
func (m *mockIdiomaRepoH) GetByID(_ context.Context, _ uuid.UUID) (*empmodels.Idioma, error) {
	return m.entity, m.err
}
func (m *mockIdiomaRepoH) Update(_ context.Context, _ *empmodels.Idioma) error { return m.err }
func (m *mockIdiomaRepoH) Delete(_ context.Context, _ uuid.UUID) error         { return m.err }
func (m *mockIdiomaRepoH) List(_ context.Context, _ map[string]interface{}, _, _ int) ([]*empmodels.Idioma, int, error) {
	return m.listItems, m.listTotal, m.err
}

func setupIdiomaRouter(repo services.IdiomaRepositoryInterface) *gin.Engine {
	r := gin.New()
	svc := services.NewIdiomaServiceWithInterface(repo)
	h := handlers.NewIdiomaHandler(svc)
	r.POST("/idiomas", h.Create)
	r.GET("/idiomas", h.List)
	r.GET("/idiomas/:id", h.GetByID)
	r.PUT("/idiomas/:id", h.Update)
	r.DELETE("/idiomas/:id", h.Delete)
	return r
}

func TestIdiomaHandler_CRUD(t *testing.T) {
	id := uuid.New()

	t.Run("Create_Success", func(t *testing.T) {
		repo := &mockIdiomaRepoH{}
		r := setupIdiomaRouter(repo)
		body := bytes.NewBufferString(`{"descricao":"Inglês"}`)
		req := httptest.NewRequest(http.MethodPost, "/idiomas", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", w.Code)
		}
	})

	t.Run("List_Success", func(t *testing.T) {
		repo := &mockIdiomaRepoH{listItems: []*empmodels.Idioma{{ID: id}}, listTotal: 1}
		r := setupIdiomaRouter(repo)
		req := httptest.NewRequest(http.MethodGet, "/idiomas", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("List_Error", func(t *testing.T) {
		repo := &mockIdiomaRepoH{err: fmt.Errorf("db error")}
		r := setupIdiomaRouter(repo)
		req := httptest.NewRequest(http.MethodGet, "/idiomas", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == http.StatusOK {
			t.Error("expected error status")
		}
	})

	t.Run("GetByID_Found", func(t *testing.T) {
		repo := &mockIdiomaRepoH{entity: &empmodels.Idioma{ID: id}}
		r := setupIdiomaRouter(repo)
		req := httptest.NewRequest(http.MethodGet, "/idiomas/"+id.String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("GetByID_NotFound", func(t *testing.T) {
		repo := &mockIdiomaRepoH{entity: nil}
		r := setupIdiomaRouter(repo)
		req := httptest.NewRequest(http.MethodGet, "/idiomas/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})

	t.Run("GetByID_InvalidID", func(t *testing.T) {
		repo := &mockIdiomaRepoH{}
		r := setupIdiomaRouter(repo)
		req := httptest.NewRequest(http.MethodGet, "/idiomas/bad", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("GetByID_ServiceError", func(t *testing.T) {
		repo := &mockIdiomaRepoH{err: fmt.Errorf("db error")}
		r := setupIdiomaRouter(repo)
		req := httptest.NewRequest(http.MethodGet, "/idiomas/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == http.StatusOK {
			t.Error("expected error status")
		}
	})

	t.Run("Create_BadJSON", func(t *testing.T) {
		repo := &mockIdiomaRepoH{}
		r := setupIdiomaRouter(repo)
		body := bytes.NewBufferString(`not json`)
		req := httptest.NewRequest(http.MethodPost, "/idiomas", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("Create_Error", func(t *testing.T) {
		repo := &mockIdiomaRepoH{err: fmt.Errorf("db error")}
		r := setupIdiomaRouter(repo)
		body := bytes.NewBufferString(`{"descricao":"Inglês"}`)
		req := httptest.NewRequest(http.MethodPost, "/idiomas", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == http.StatusCreated {
			t.Error("expected error status")
		}
	})

	t.Run("Update_Success", func(t *testing.T) {
		repo := &mockIdiomaRepoH{}
		r := setupIdiomaRouter(repo)
		body := bytes.NewBufferString(`{"descricao":"Espanhol"}`)
		req := httptest.NewRequest(http.MethodPut, "/idiomas/"+id.String(), body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("Update_InvalidID", func(t *testing.T) {
		repo := &mockIdiomaRepoH{}
		r := setupIdiomaRouter(repo)
		body := bytes.NewBufferString(`{"descricao":"Espanhol"}`)
		req := httptest.NewRequest(http.MethodPut, "/idiomas/bad-id", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("Delete_Success", func(t *testing.T) {
		repo := &mockIdiomaRepoH{}
		r := setupIdiomaRouter(repo)
		req := httptest.NewRequest(http.MethodDelete, "/idiomas/"+id.String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("Delete_InvalidID", func(t *testing.T) {
		repo := &mockIdiomaRepoH{}
		r := setupIdiomaRouter(repo)
		req := httptest.NewRequest(http.MethodDelete, "/idiomas/bad", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// NivelIdioma handler tests
// ──────────────────────────────────────────────────────────────────────────────

type mockNivelIdiomaRepoH struct {
	entity    *empmodels.NivelIdioma
	listItems []*empmodels.NivelIdioma
	listTotal int
	err       error
}

func (m *mockNivelIdiomaRepoH) Create(_ context.Context, _ *empmodels.NivelIdioma) (uuid.UUID, error) {
	if m.err != nil {
		return uuid.Nil, m.err
	}
	return uuid.New(), nil
}
func (m *mockNivelIdiomaRepoH) GetByID(_ context.Context, _ uuid.UUID) (*empmodels.NivelIdioma, error) {
	return m.entity, m.err
}
func (m *mockNivelIdiomaRepoH) Update(_ context.Context, _ *empmodels.NivelIdioma) error {
	return m.err
}
func (m *mockNivelIdiomaRepoH) Delete(_ context.Context, _ uuid.UUID) error { return m.err }
func (m *mockNivelIdiomaRepoH) List(_ context.Context, _ map[string]interface{}, _, _ int) ([]*empmodels.NivelIdioma, int, error) {
	return m.listItems, m.listTotal, m.err
}

func setupNivelIdiomaRouter(repo services.NivelIdiomaRepositoryInterface) *gin.Engine {
	r := gin.New()
	svc := services.NewNivelIdiomaServiceWithInterface(repo)
	h := handlers.NewNivelIdiomaHandler(svc)
	r.POST("/niveis", h.Create)
	r.GET("/niveis", h.List)
	r.GET("/niveis/:id", h.GetByID)
	r.PUT("/niveis/:id", h.Update)
	r.DELETE("/niveis/:id", h.Delete)
	return r
}

func TestNivelIdiomaHandler_CRUD(t *testing.T) {
	id := uuid.New()

	t.Run("Create_Success", func(t *testing.T) {
		repo := &mockNivelIdiomaRepoH{}
		r := setupNivelIdiomaRouter(repo)
		body := bytes.NewBufferString(`{"descricao":"Básico"}`)
		req := httptest.NewRequest(http.MethodPost, "/niveis", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", w.Code)
		}
	})

	t.Run("List_Success", func(t *testing.T) {
		repo := &mockNivelIdiomaRepoH{listItems: []*empmodels.NivelIdioma{{ID: id}}, listTotal: 1}
		r := setupNivelIdiomaRouter(repo)
		req := httptest.NewRequest(http.MethodGet, "/niveis", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("GetByID_Found", func(t *testing.T) {
		repo := &mockNivelIdiomaRepoH{entity: &empmodels.NivelIdioma{ID: id}}
		r := setupNivelIdiomaRouter(repo)
		req := httptest.NewRequest(http.MethodGet, "/niveis/"+id.String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("GetByID_NotFound", func(t *testing.T) {
		repo := &mockNivelIdiomaRepoH{entity: nil}
		r := setupNivelIdiomaRouter(repo)
		req := httptest.NewRequest(http.MethodGet, "/niveis/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})

	t.Run("Update_Success", func(t *testing.T) {
		repo := &mockNivelIdiomaRepoH{}
		r := setupNivelIdiomaRouter(repo)
		body := bytes.NewBufferString(`{"descricao":"Avançado"}`)
		req := httptest.NewRequest(http.MethodPut, "/niveis/"+id.String(), body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("Delete_Success", func(t *testing.T) {
		repo := &mockNivelIdiomaRepoH{}
		r := setupNivelIdiomaRouter(repo)
		req := httptest.NewRequest(http.MethodDelete, "/niveis/"+id.String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("GetByID_InvalidID", func(t *testing.T) {
		repo := &mockNivelIdiomaRepoH{}
		r := setupNivelIdiomaRouter(repo)
		req := httptest.NewRequest(http.MethodGet, "/niveis/bad-id", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("GetByID_ServiceError", func(t *testing.T) {
		repo := &mockNivelIdiomaRepoH{err: fmt.Errorf("db error")}
		r := setupNivelIdiomaRouter(repo)
		req := httptest.NewRequest(http.MethodGet, "/niveis/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == http.StatusOK {
			t.Error("expected error status")
		}
	})

	t.Run("List_Error", func(t *testing.T) {
		repo := &mockNivelIdiomaRepoH{err: fmt.Errorf("db error")}
		r := setupNivelIdiomaRouter(repo)
		req := httptest.NewRequest(http.MethodGet, "/niveis", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == http.StatusOK {
			t.Error("expected error status")
		}
	})

	t.Run("Create_BadJSON", func(t *testing.T) {
		repo := &mockNivelIdiomaRepoH{}
		r := setupNivelIdiomaRouter(repo)
		body := bytes.NewBufferString(`not json`)
		req := httptest.NewRequest(http.MethodPost, "/niveis", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("Create_Error", func(t *testing.T) {
		repo := &mockNivelIdiomaRepoH{err: fmt.Errorf("db error")}
		r := setupNivelIdiomaRouter(repo)
		body := bytes.NewBufferString(`{"descricao":"Básico"}`)
		req := httptest.NewRequest(http.MethodPost, "/niveis", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == http.StatusCreated {
			t.Error("expected error status")
		}
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// Escolaridade handler tests
// ──────────────────────────────────────────────────────────────────────────────

type mockEscolaridadeRepoH struct {
	entity    *empmodels.Escolaridade
	listItems []*empmodels.Escolaridade
	listTotal int
	err       error
}

func (m *mockEscolaridadeRepoH) Create(_ context.Context, _ *empmodels.Escolaridade) (uuid.UUID, error) {
	if m.err != nil {
		return uuid.Nil, m.err
	}
	return uuid.New(), nil
}
func (m *mockEscolaridadeRepoH) GetByID(_ context.Context, _ uuid.UUID) (*empmodels.Escolaridade, error) {
	return m.entity, m.err
}
func (m *mockEscolaridadeRepoH) Update(_ context.Context, _ *empmodels.Escolaridade) error {
	return m.err
}
func (m *mockEscolaridadeRepoH) Delete(_ context.Context, _ uuid.UUID) error { return m.err }
func (m *mockEscolaridadeRepoH) List(_ context.Context, _ map[string]interface{}, _, _ int) ([]*empmodels.Escolaridade, int, error) {
	return m.listItems, m.listTotal, m.err
}

func setupEscolaridadeRouter(repo services.EmpEscolaridadeRepositoryInterface) *gin.Engine {
	r := gin.New()
	svc := services.NewEscolaridadeServiceWithInterface(repo)
	h := handlers.NewEscolaridadeHandler(svc)
	r.POST("/escolaridades", h.Create)
	r.GET("/escolaridades", h.List)
	r.GET("/escolaridades/:id", h.GetByID)
	r.PUT("/escolaridades/:id", h.Update)
	r.DELETE("/escolaridades/:id", h.Delete)
	return r
}

func TestEscolaridadeHandler_CRUD(t *testing.T) {
	id := uuid.New()

	t.Run("Create_Success", func(t *testing.T) {
		repo := &mockEscolaridadeRepoH{}
		r := setupEscolaridadeRouter(repo)
		body := bytes.NewBufferString(`{"descricao":"Superior"}`)
		req := httptest.NewRequest(http.MethodPost, "/escolaridades", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", w.Code)
		}
	})

	t.Run("List_Success", func(t *testing.T) {
		repo := &mockEscolaridadeRepoH{listItems: []*empmodels.Escolaridade{{ID: id}}, listTotal: 1}
		r := setupEscolaridadeRouter(repo)
		req := httptest.NewRequest(http.MethodGet, "/escolaridades", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("GetByID_Found", func(t *testing.T) {
		repo := &mockEscolaridadeRepoH{entity: &empmodels.Escolaridade{ID: id}}
		r := setupEscolaridadeRouter(repo)
		req := httptest.NewRequest(http.MethodGet, "/escolaridades/"+id.String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("GetByID_NotFound", func(t *testing.T) {
		repo := &mockEscolaridadeRepoH{entity: nil}
		r := setupEscolaridadeRouter(repo)
		req := httptest.NewRequest(http.MethodGet, "/escolaridades/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})

	t.Run("Update_Success", func(t *testing.T) {
		repo := &mockEscolaridadeRepoH{}
		r := setupEscolaridadeRouter(repo)
		body := bytes.NewBufferString(`{"descricao":"Pós-graduação"}`)
		req := httptest.NewRequest(http.MethodPut, "/escolaridades/"+id.String(), body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("Delete_Success", func(t *testing.T) {
		repo := &mockEscolaridadeRepoH{}
		r := setupEscolaridadeRouter(repo)
		req := httptest.NewRequest(http.MethodDelete, "/escolaridades/"+id.String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("GetByID_InvalidID", func(t *testing.T) {
		repo := &mockEscolaridadeRepoH{}
		r := setupEscolaridadeRouter(repo)
		req := httptest.NewRequest(http.MethodGet, "/escolaridades/bad-id", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("GetByID_ServiceError", func(t *testing.T) {
		repo := &mockEscolaridadeRepoH{err: fmt.Errorf("db error")}
		r := setupEscolaridadeRouter(repo)
		req := httptest.NewRequest(http.MethodGet, "/escolaridades/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == http.StatusOK {
			t.Error("expected error status")
		}
	})

	t.Run("List_Error", func(t *testing.T) {
		repo := &mockEscolaridadeRepoH{err: fmt.Errorf("db error")}
		r := setupEscolaridadeRouter(repo)
		req := httptest.NewRequest(http.MethodGet, "/escolaridades", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == http.StatusOK {
			t.Error("expected error status")
		}
	})

	t.Run("Create_BadJSON", func(t *testing.T) {
		repo := &mockEscolaridadeRepoH{}
		r := setupEscolaridadeRouter(repo)
		body := bytes.NewBufferString(`not json`)
		req := httptest.NewRequest(http.MethodPost, "/escolaridades", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("Create_Error", func(t *testing.T) {
		repo := &mockEscolaridadeRepoH{err: fmt.Errorf("db error")}
		r := setupEscolaridadeRouter(repo)
		body := bytes.NewBufferString(`{"descricao":"Superior"}`)
		req := httptest.NewRequest(http.MethodPost, "/escolaridades", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == http.StatusCreated {
			t.Error("expected error status")
		}
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// TipoConquista handler tests
// ──────────────────────────────────────────────────────────────────────────────

type mockTipoConquistaRepoH struct {
	entity    *empmodels.TipoConquista
	listItems []*empmodels.TipoConquista
	listTotal int
	err       error
}

func (m *mockTipoConquistaRepoH) Create(_ context.Context, _ *empmodels.TipoConquista) (uuid.UUID, error) {
	if m.err != nil {
		return uuid.Nil, m.err
	}
	return uuid.New(), nil
}
func (m *mockTipoConquistaRepoH) GetByID(_ context.Context, _ uuid.UUID) (*empmodels.TipoConquista, error) {
	return m.entity, m.err
}
func (m *mockTipoConquistaRepoH) Update(_ context.Context, _ *empmodels.TipoConquista) error {
	return m.err
}
func (m *mockTipoConquistaRepoH) Delete(_ context.Context, _ uuid.UUID) error { return m.err }
func (m *mockTipoConquistaRepoH) List(_ context.Context, _ map[string]interface{}, _, _ int) ([]*empmodels.TipoConquista, int, error) {
	return m.listItems, m.listTotal, m.err
}

func setupTipoConquistaRouter(repo services.TipoConquistaRepositoryInterface) *gin.Engine {
	r := gin.New()
	svc := services.NewTipoConquistaServiceWithInterface(repo)
	h := handlers.NewTipoConquistaHandler(svc)
	r.POST("/tipos-conquista", h.Create)
	r.GET("/tipos-conquista", h.List)
	r.GET("/tipos-conquista/:id", h.GetByID)
	r.PUT("/tipos-conquista/:id", h.Update)
	r.DELETE("/tipos-conquista/:id", h.Delete)
	return r
}

func TestTipoConquistaHandler_CRUD(t *testing.T) {
	id := uuid.New()

	t.Run("Create_Success", func(t *testing.T) {
		repo := &mockTipoConquistaRepoH{}
		r := setupTipoConquistaRouter(repo)
		body := bytes.NewBufferString(`{"descricao":"Certificado"}`)
		req := httptest.NewRequest(http.MethodPost, "/tipos-conquista", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", w.Code)
		}
	})

	t.Run("List_Success", func(t *testing.T) {
		repo := &mockTipoConquistaRepoH{listItems: []*empmodels.TipoConquista{{ID: id}}, listTotal: 1}
		r := setupTipoConquistaRouter(repo)
		req := httptest.NewRequest(http.MethodGet, "/tipos-conquista", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("GetByID_Found", func(t *testing.T) {
		repo := &mockTipoConquistaRepoH{entity: &empmodels.TipoConquista{ID: id}}
		r := setupTipoConquistaRouter(repo)
		req := httptest.NewRequest(http.MethodGet, "/tipos-conquista/"+id.String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("GetByID_NotFound", func(t *testing.T) {
		repo := &mockTipoConquistaRepoH{entity: nil}
		r := setupTipoConquistaRouter(repo)
		req := httptest.NewRequest(http.MethodGet, "/tipos-conquista/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})

	t.Run("Update_Success", func(t *testing.T) {
		repo := &mockTipoConquistaRepoH{}
		r := setupTipoConquistaRouter(repo)
		body := bytes.NewBufferString(`{"descricao":"Premio"}`)
		req := httptest.NewRequest(http.MethodPut, "/tipos-conquista/"+id.String(), body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("Delete_Success", func(t *testing.T) {
		repo := &mockTipoConquistaRepoH{}
		r := setupTipoConquistaRouter(repo)
		req := httptest.NewRequest(http.MethodDelete, "/tipos-conquista/"+id.String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("GetByID_InvalidID", func(t *testing.T) {
		repo := &mockTipoConquistaRepoH{}
		r := setupTipoConquistaRouter(repo)
		req := httptest.NewRequest(http.MethodGet, "/tipos-conquista/bad-id", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("GetByID_ServiceError", func(t *testing.T) {
		repo := &mockTipoConquistaRepoH{err: fmt.Errorf("db error")}
		r := setupTipoConquistaRouter(repo)
		req := httptest.NewRequest(http.MethodGet, "/tipos-conquista/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == http.StatusOK {
			t.Error("expected error status")
		}
	})

	t.Run("List_Error", func(t *testing.T) {
		repo := &mockTipoConquistaRepoH{err: fmt.Errorf("db error")}
		r := setupTipoConquistaRouter(repo)
		req := httptest.NewRequest(http.MethodGet, "/tipos-conquista", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == http.StatusOK {
			t.Error("expected error status")
		}
	})

	t.Run("Create_BadJSON", func(t *testing.T) {
		repo := &mockTipoConquistaRepoH{}
		r := setupTipoConquistaRouter(repo)
		body := bytes.NewBufferString(`not json`)
		req := httptest.NewRequest(http.MethodPost, "/tipos-conquista", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("Create_Error", func(t *testing.T) {
		repo := &mockTipoConquistaRepoH{err: fmt.Errorf("db error")}
		r := setupTipoConquistaRouter(repo)
		body := bytes.NewBufferString(`{"descricao":"Certificado"}`)
		req := httptest.NewRequest(http.MethodPost, "/tipos-conquista", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == http.StatusCreated {
			t.Error("expected error status")
		}
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// SituacaoAtual handler tests
// ──────────────────────────────────────────────────────────────────────────────

type mockSituacaoAtualRepoH struct {
	entity    *empmodels.SituacaoAtual
	listItems []*empmodels.SituacaoAtual
	listTotal int
	err       error
}

func (m *mockSituacaoAtualRepoH) Create(_ context.Context, _ *empmodels.SituacaoAtual) (uuid.UUID, error) {
	if m.err != nil {
		return uuid.Nil, m.err
	}
	return uuid.New(), nil
}
func (m *mockSituacaoAtualRepoH) GetByID(_ context.Context, _ uuid.UUID) (*empmodels.SituacaoAtual, error) {
	return m.entity, m.err
}
func (m *mockSituacaoAtualRepoH) Update(_ context.Context, _ *empmodels.SituacaoAtual) error {
	return m.err
}
func (m *mockSituacaoAtualRepoH) Delete(_ context.Context, _ uuid.UUID) error { return m.err }
func (m *mockSituacaoAtualRepoH) List(_ context.Context, _ map[string]interface{}, _, _ int) ([]*empmodels.SituacaoAtual, int, error) {
	return m.listItems, m.listTotal, m.err
}

func setupSituacaoAtualRouter(repo services.SituacaoAtualRepositoryInterface) *gin.Engine {
	r := gin.New()
	svc := services.NewSituacaoAtualServiceWithInterface(repo)
	h := handlers.NewSituacaoAtualHandler(svc)
	r.POST("/situacoes", h.Create)
	r.GET("/situacoes", h.List)
	r.GET("/situacoes/:id", h.GetByID)
	r.PUT("/situacoes/:id", h.Update)
	r.DELETE("/situacoes/:id", h.Delete)
	return r
}

func TestSituacaoAtualHandler_CRUD(t *testing.T) {
	id := uuid.New()

	t.Run("Create_Success", func(t *testing.T) {
		repo := &mockSituacaoAtualRepoH{}
		r := setupSituacaoAtualRouter(repo)
		body := bytes.NewBufferString(`{"descricao":"Empregado"}`)
		req := httptest.NewRequest(http.MethodPost, "/situacoes", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", w.Code)
		}
	})

	t.Run("List_Success", func(t *testing.T) {
		repo := &mockSituacaoAtualRepoH{listItems: []*empmodels.SituacaoAtual{{ID: id}}, listTotal: 1}
		r := setupSituacaoAtualRouter(repo)
		req := httptest.NewRequest(http.MethodGet, "/situacoes", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("GetByID_Found", func(t *testing.T) {
		repo := &mockSituacaoAtualRepoH{entity: &empmodels.SituacaoAtual{ID: id}}
		r := setupSituacaoAtualRouter(repo)
		req := httptest.NewRequest(http.MethodGet, "/situacoes/"+id.String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("GetByID_NotFound", func(t *testing.T) {
		repo := &mockSituacaoAtualRepoH{entity: nil}
		r := setupSituacaoAtualRouter(repo)
		req := httptest.NewRequest(http.MethodGet, "/situacoes/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})

	t.Run("Update_Success", func(t *testing.T) {
		repo := &mockSituacaoAtualRepoH{}
		r := setupSituacaoAtualRouter(repo)
		body := bytes.NewBufferString(`{"descricao":"Desempregado"}`)
		req := httptest.NewRequest(http.MethodPut, "/situacoes/"+id.String(), body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("Delete_Success", func(t *testing.T) {
		repo := &mockSituacaoAtualRepoH{}
		r := setupSituacaoAtualRouter(repo)
		req := httptest.NewRequest(http.MethodDelete, "/situacoes/"+id.String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("GetByID_InvalidID", func(t *testing.T) {
		repo := &mockSituacaoAtualRepoH{}
		r := setupSituacaoAtualRouter(repo)
		req := httptest.NewRequest(http.MethodGet, "/situacoes/bad-id", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("GetByID_ServiceError", func(t *testing.T) {
		repo := &mockSituacaoAtualRepoH{err: fmt.Errorf("db error")}
		r := setupSituacaoAtualRouter(repo)
		req := httptest.NewRequest(http.MethodGet, "/situacoes/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == http.StatusOK {
			t.Error("expected error status")
		}
	})

	t.Run("List_Error", func(t *testing.T) {
		repo := &mockSituacaoAtualRepoH{err: fmt.Errorf("db error")}
		r := setupSituacaoAtualRouter(repo)
		req := httptest.NewRequest(http.MethodGet, "/situacoes", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == http.StatusOK {
			t.Error("expected error status")
		}
	})

	t.Run("Create_BadJSON", func(t *testing.T) {
		repo := &mockSituacaoAtualRepoH{}
		r := setupSituacaoAtualRouter(repo)
		body := bytes.NewBufferString(`not json`)
		req := httptest.NewRequest(http.MethodPost, "/situacoes", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("Create_Error", func(t *testing.T) {
		repo := &mockSituacaoAtualRepoH{err: fmt.Errorf("db error")}
		r := setupSituacaoAtualRouter(repo)
		body := bytes.NewBufferString(`{"descricao":"Empregado"}`)
		req := httptest.NewRequest(http.MethodPost, "/situacoes", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == http.StatusCreated {
			t.Error("expected error status")
		}
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// Disponibilidade handler tests
// ──────────────────────────────────────────────────────────────────────────────

type mockDisponibilidadeRepoH struct {
	entity    *empmodels.Disponibilidade
	listItems []*empmodels.Disponibilidade
	listTotal int
	err       error
}

func (m *mockDisponibilidadeRepoH) Create(_ context.Context, _ *empmodels.Disponibilidade) (uuid.UUID, error) {
	if m.err != nil {
		return uuid.Nil, m.err
	}
	return uuid.New(), nil
}
func (m *mockDisponibilidadeRepoH) GetByID(_ context.Context, _ uuid.UUID) (*empmodels.Disponibilidade, error) {
	return m.entity, m.err
}
func (m *mockDisponibilidadeRepoH) Update(_ context.Context, _ *empmodels.Disponibilidade) error {
	return m.err
}
func (m *mockDisponibilidadeRepoH) Delete(_ context.Context, _ uuid.UUID) error { return m.err }
func (m *mockDisponibilidadeRepoH) List(_ context.Context, _ map[string]interface{}, _, _ int) ([]*empmodels.Disponibilidade, int, error) {
	return m.listItems, m.listTotal, m.err
}

func setupDisponibilidadeRouter(repo services.DisponibilidadeRepositoryInterface) *gin.Engine {
	r := gin.New()
	svc := services.NewDisponibilidadeServiceWithInterface(repo)
	h := handlers.NewDisponibilidadeHandler(svc)
	r.POST("/disponibilidades", h.Create)
	r.GET("/disponibilidades", h.List)
	r.GET("/disponibilidades/:id", h.GetByID)
	r.PUT("/disponibilidades/:id", h.Update)
	r.DELETE("/disponibilidades/:id", h.Delete)
	return r
}

func TestDisponibilidadeHandler_CRUD(t *testing.T) {
	id := uuid.New()

	t.Run("Create_Success", func(t *testing.T) {
		repo := &mockDisponibilidadeRepoH{}
		r := setupDisponibilidadeRouter(repo)
		body := bytes.NewBufferString(`{"descricao":"Imediata"}`)
		req := httptest.NewRequest(http.MethodPost, "/disponibilidades", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", w.Code)
		}
	})

	t.Run("List_Success", func(t *testing.T) {
		repo := &mockDisponibilidadeRepoH{listItems: []*empmodels.Disponibilidade{{ID: id}}, listTotal: 1}
		r := setupDisponibilidadeRouter(repo)
		req := httptest.NewRequest(http.MethodGet, "/disponibilidades", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("GetByID_Found", func(t *testing.T) {
		repo := &mockDisponibilidadeRepoH{entity: &empmodels.Disponibilidade{ID: id}}
		r := setupDisponibilidadeRouter(repo)
		req := httptest.NewRequest(http.MethodGet, "/disponibilidades/"+id.String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("GetByID_NotFound", func(t *testing.T) {
		repo := &mockDisponibilidadeRepoH{entity: nil}
		r := setupDisponibilidadeRouter(repo)
		req := httptest.NewRequest(http.MethodGet, "/disponibilidades/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})

	t.Run("Update_Success", func(t *testing.T) {
		repo := &mockDisponibilidadeRepoH{}
		r := setupDisponibilidadeRouter(repo)
		body := bytes.NewBufferString(`{"descricao":"30 dias"}`)
		req := httptest.NewRequest(http.MethodPut, "/disponibilidades/"+id.String(), body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("Delete_Success", func(t *testing.T) {
		repo := &mockDisponibilidadeRepoH{}
		r := setupDisponibilidadeRouter(repo)
		req := httptest.NewRequest(http.MethodDelete, "/disponibilidades/"+id.String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("GetByID_InvalidID", func(t *testing.T) {
		repo := &mockDisponibilidadeRepoH{}
		r := setupDisponibilidadeRouter(repo)
		req := httptest.NewRequest(http.MethodGet, "/disponibilidades/bad-id", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("GetByID_ServiceError", func(t *testing.T) {
		repo := &mockDisponibilidadeRepoH{err: fmt.Errorf("db error")}
		r := setupDisponibilidadeRouter(repo)
		req := httptest.NewRequest(http.MethodGet, "/disponibilidades/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == http.StatusOK {
			t.Error("expected error status")
		}
	})

	t.Run("List_Error", func(t *testing.T) {
		repo := &mockDisponibilidadeRepoH{err: fmt.Errorf("db error")}
		r := setupDisponibilidadeRouter(repo)
		req := httptest.NewRequest(http.MethodGet, "/disponibilidades", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == http.StatusOK {
			t.Error("expected error status")
		}
	})

	t.Run("Create_BadJSON", func(t *testing.T) {
		repo := &mockDisponibilidadeRepoH{}
		r := setupDisponibilidadeRouter(repo)
		body := bytes.NewBufferString(`not json`)
		req := httptest.NewRequest(http.MethodPost, "/disponibilidades", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("Create_Error", func(t *testing.T) {
		repo := &mockDisponibilidadeRepoH{err: fmt.Errorf("db error")}
		r := setupDisponibilidadeRouter(repo)
		body := bytes.NewBufferString(`{"descricao":"Imediata"}`)
		req := httptest.NewRequest(http.MethodPost, "/disponibilidades", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == http.StatusCreated {
			t.Error("expected error status")
		}
	})
}
