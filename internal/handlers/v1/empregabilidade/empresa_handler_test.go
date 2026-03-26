package empregabilidade_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	handlers "github.com/prefeitura-rio/app-go-api/internal/handlers/v1/empregabilidade"
	empmodels "github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
)

// ──────────────────────────────────────────────────────────────────────────────
// Mock for EmpresaRepositoryInterface (empregabilidade)
// ──────────────────────────────────────────────────────────────────────────────

type mockEmpresaRepoEmpH struct {
	cnpj      string
	entity    *empmodels.Empresa
	listItems []*empmodels.Empresa
	listTotal int
	err       error
}

func (m *mockEmpresaRepoEmpH) Create(_ context.Context, e *empmodels.Empresa) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	if m.cnpj != "" {
		return m.cnpj, nil
	}
	return "12345678000100", nil
}
func (m *mockEmpresaRepoEmpH) GetByID(_ context.Context, _ string) (*empmodels.Empresa, error) {
	return m.entity, m.err
}
func (m *mockEmpresaRepoEmpH) Update(_ context.Context, _ *empmodels.Empresa) error { return m.err }
func (m *mockEmpresaRepoEmpH) Delete(_ context.Context, _ string) error             { return m.err }
func (m *mockEmpresaRepoEmpH) List(_ context.Context, _ empmodels.EmpresaFilter, _, _ int) ([]*empmodels.Empresa, int, error) {
	return m.listItems, m.listTotal, m.err
}
func (m *mockEmpresaRepoEmpH) Upsert(_ context.Context, _ *empmodels.Empresa) error { return m.err }

func setupEmpresaHandlerRouter(repo services.EmpresaRepositoryInterface) *gin.Engine {
	r := gin.New()
	svc := services.NewEmpresaServiceWithInterface(repo)
	h := handlers.NewEmpresaHandler(svc)
	r.POST("/empresas", h.Create)
	r.GET("/empresas", h.List)
	r.GET("/empresas/:cnpj", h.GetByID)
	r.PUT("/empresas/:cnpj", h.Update)
	r.DELETE("/empresas/:cnpj", h.Delete)
	r.GET("/empresas/consulta/:cnpj", h.ConsultaCNPJ)
	return r
}

func TestEmpresaHandler_Create_Success(t *testing.T) {
	repo := &mockEmpresaRepoEmpH{}
	r := setupEmpresaHandlerRouter(repo)
	body := bytes.NewBufferString(`{"cnpj":"12345678000100","razao_social":"Empresa Teste"}`)
	req := httptest.NewRequest(http.MethodPost, "/empresas", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestEmpresaHandler_Create_BadJSON(t *testing.T) {
	repo := &mockEmpresaRepoEmpH{}
	r := setupEmpresaHandlerRouter(repo)
	body := bytes.NewBufferString(`invalid`)
	req := httptest.NewRequest(http.MethodPost, "/empresas", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestEmpresaHandler_Create_Error(t *testing.T) {
	repo := &mockEmpresaRepoEmpH{err: fmt.Errorf("db error")}
	r := setupEmpresaHandlerRouter(repo)
	body := bytes.NewBufferString(`{"cnpj":"12345678000100"}`)
	req := httptest.NewRequest(http.MethodPost, "/empresas", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestEmpresaHandler_List_Success(t *testing.T) {
	repo := &mockEmpresaRepoEmpH{
		listItems: []*empmodels.Empresa{{CNPJ: "12345678000100", RazaoSocial: "Empresa Teste"}},
		listTotal: 1,
	}
	r := setupEmpresaHandlerRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/empresas", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestEmpresaHandler_List_WithFilters(t *testing.T) {
	repo := &mockEmpresaRepoEmpH{listItems: []*empmodels.Empresa{}, listTotal: 0}
	r := setupEmpresaHandlerRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/empresas?search=teste&cnpj=123&page=2&pageSize=5", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestEmpresaHandler_List_Error(t *testing.T) {
	repo := &mockEmpresaRepoEmpH{err: fmt.Errorf("db error")}
	r := setupEmpresaHandlerRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/empresas", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestEmpresaHandler_GetByID_Found(t *testing.T) {
	repo := &mockEmpresaRepoEmpH{entity: &empmodels.Empresa{CNPJ: "12345678000100", RazaoSocial: "Empresa Teste"}}
	r := setupEmpresaHandlerRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/empresas/12345678000100", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestEmpresaHandler_GetByID_NotFound(t *testing.T) {
	repo := &mockEmpresaRepoEmpH{entity: nil}
	r := setupEmpresaHandlerRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/empresas/99999999000100", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestEmpresaHandler_GetByID_Error(t *testing.T) {
	repo := &mockEmpresaRepoEmpH{err: fmt.Errorf("db error")}
	r := setupEmpresaHandlerRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/empresas/12345678000100", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestEmpresaHandler_Update_Success(t *testing.T) {
	repo := &mockEmpresaRepoEmpH{}
	r := setupEmpresaHandlerRouter(repo)
	body := bytes.NewBufferString(`{"razao_social":"Nova Razao"}`)
	req := httptest.NewRequest(http.MethodPut, "/empresas/12345678000100", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestEmpresaHandler_Update_BadJSON(t *testing.T) {
	repo := &mockEmpresaRepoEmpH{}
	r := setupEmpresaHandlerRouter(repo)
	body := bytes.NewBufferString(`invalid json`)
	req := httptest.NewRequest(http.MethodPut, "/empresas/12345678000100", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestEmpresaHandler_Update_Error(t *testing.T) {
	repo := &mockEmpresaRepoEmpH{err: fmt.Errorf("db error")}
	r := setupEmpresaHandlerRouter(repo)
	body := bytes.NewBufferString(`{"razao_social":"Nova"}`)
	req := httptest.NewRequest(http.MethodPut, "/empresas/12345678000100", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestEmpresaHandler_Delete_Success(t *testing.T) {
	repo := &mockEmpresaRepoEmpH{}
	r := setupEmpresaHandlerRouter(repo)
	req := httptest.NewRequest(http.MethodDelete, "/empresas/12345678000100", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestEmpresaHandler_Delete_Error(t *testing.T) {
	repo := &mockEmpresaRepoEmpH{err: fmt.Errorf("db error")}
	r := setupEmpresaHandlerRouter(repo)
	req := httptest.NewRequest(http.MethodDelete, "/empresas/12345678000100", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestEmpresaHandler_ConsultaCNPJ_NoService(t *testing.T) {
	repo := &mockEmpresaRepoEmpH{}
	r := setupEmpresaHandlerRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/empresas/consulta/12345678000100", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	// No CNPJ consulta service injected, expect 503
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}
}
