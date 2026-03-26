package empregabilidade_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	handlers "github.com/prefeitura-rio/app-go-api/internal/handlers/v1/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	empmodels "github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	empRepository "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
)

// ──────────────────────────────────────────────────────────────────────────────
// Mock CandidaturaRepository (implements CandidaturaRepositoryInterface)
// ──────────────────────────────────────────────────────────────────────────────

type mockCandidaturaRepoH struct {
	entity    *empmodels.Candidatura
	listItems []*empmodels.Candidatura
	listTotal int
	exists    bool
	err       error
	countMap  map[empmodels.StatusCandidatura]int64
}

func (m *mockCandidaturaRepoH) Create(_ context.Context, _ *empmodels.Candidatura) (uuid.UUID, error) {
	if m.err != nil {
		return uuid.Nil, m.err
	}
	return uuid.New(), nil
}

func (m *mockCandidaturaRepoH) GetByID(_ context.Context, _ uuid.UUID) (*empmodels.Candidatura, error) {
	return m.entity, m.err
}

func (m *mockCandidaturaRepoH) Update(_ context.Context, _ *empmodels.Candidatura) error {
	return m.err
}

func (m *mockCandidaturaRepoH) Delete(_ context.Context, _ uuid.UUID) error {
	return m.err
}

func (m *mockCandidaturaRepoH) List(_ context.Context, _ empmodels.CandidaturaFilter, _, _ int) ([]*empmodels.Candidatura, int, error) {
	return m.listItems, m.listTotal, m.err
}

func (m *mockCandidaturaRepoH) CheckExistingCandidatura(_ context.Context, _ string, _ uuid.UUID) (bool, error) {
	return m.exists, m.err
}

func (m *mockCandidaturaRepoH) UpdateStatus(_ context.Context, _ uuid.UUID, _ empmodels.StatusCandidatura) error {
	return m.err
}

func (m *mockCandidaturaRepoH) UpdateEtapa(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
	return m.err
}

func (m *mockCandidaturaRepoH) BulkUpdateStatus(_ context.Context, _ uuid.UUID, _ []string, _ empmodels.StatusCandidatura) (empRepository.BulkUpdateResult, error) {
	if m.err != nil {
		return empRepository.BulkUpdateResult{}, m.err
	}
	return empRepository.BulkUpdateResult{Updated: 1, FailedCPFs: nil}, nil
}

func (m *mockCandidaturaRepoH) BulkGetByCPFs(_ context.Context, _ uuid.UUID, _ []string) ([]*empmodels.Candidatura, error) {
	return m.listItems, m.err
}

func (m *mockCandidaturaRepoH) BulkUpdateEtapa(_ context.Context, _ []uuid.UUID, _ uuid.UUID) error {
	return m.err
}

func (m *mockCandidaturaRepoH) BulkSaveAndUpdateStatusByVagaID(_ context.Context, _ uuid.UUID, _ empmodels.StatusCandidatura) error {
	return m.err
}

func (m *mockCandidaturaRepoH) BulkRestoreStatusByVagaID(_ context.Context, _ uuid.UUID) error {
	return m.err
}

func (m *mockCandidaturaRepoH) CountByStatus(_ context.Context, _ empmodels.CandidaturaFilter) (map[empmodels.StatusCandidatura]int64, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.countMap != nil {
		return m.countMap, nil
	}
	return map[empmodels.StatusCandidatura]int64{}, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Mock VagaRepo for CandidaturaService (implements VagaRepositoryInterface)
// ──────────────────────────────────────────────────────────────────────────────

type mockVagaForCandidaturaH struct {
	entity *empmodels.Vaga
	err    error
}

func (m *mockVagaForCandidaturaH) GetByID(_ context.Context, _ uuid.UUID) (*empmodels.Vaga, error) {
	return m.entity, m.err
}

// ──────────────────────────────────────────────────────────────────────────────
// Mock CurriculoService for CandidaturaService
// ──────────────────────────────────────────────────────────────────────────────

type mockCurriculoSvcH struct {
	err error
}

func (m *mockCurriculoSvcH) GetCurriculoCompleto(_ context.Context, _ string) (*empmodels.CurriculoCompleto, error) {
	return nil, m.err
}

// ──────────────────────────────────────────────────────────────────────────────
// Router setup
// ──────────────────────────────────────────────────────────────────────────────

func setupCandidaturaRouter(
	candRepo services.CandidaturaRepositoryInterface,
	vagaRepo services.VagaRepositoryInterface,
	cpf string,
	isAdmin bool,
) *gin.Engine {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if cpf != "" {
			c.Set(middlewares.UserCPFKey, cpf)
		}
		if isAdmin {
			c.Set(middlewares.UserRoleKey, "ADMIN")
		}
		c.Next()
	})
	curriculoSvc := &mockCurriculoSvcH{}
	svc := services.NewCandidaturaService(candRepo, vagaRepo, curriculoSvc, nil, nil)
	h := handlers.NewCandidaturaHandler(svc)
	r.POST("/candidaturas", h.Create)
	r.GET("/candidaturas", h.List)
	r.GET("/candidaturas/usuario/:cpf", h.ListByCPF)
	r.GET("/candidaturas/:id", h.GetByID)
	r.PUT("/candidaturas/:id", h.Update)
	r.DELETE("/candidaturas/:id", h.Delete)
	r.PUT("/candidaturas/:id/status", h.UpdateStatus)
	r.PUT("/candidaturas/:id/approve", h.Approve)
	r.PUT("/candidaturas/:id/reject", h.Reject)
	r.PUT("/candidaturas/bulk-status", h.BulkUpdateStatus)
	r.PUT("/candidaturas/bulk-etapa", h.BulkUpdateEtapa)
	r.PUT("/candidaturas/:id/etapa", h.UpdateEtapa)
	return r
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests: Create
// ──────────────────────────────────────────────────────────────────────────────

func TestCandidaturaHandler_Create_AsAdmin_Success(t *testing.T) {
	activeVaga := &empmodels.Vaga{
		ID:     uuid.MustParse(validUUID),
		Status: empmodels.StatusVagaPublicadoAtivo,
	}
	vagaRepo := &mockVagaForCandidaturaH{entity: activeVaga}
	candRepo := &mockCandidaturaRepoH{exists: false}
	r := setupCandidaturaRouter(candRepo, vagaRepo, "admin-user", true)
	body := bodyOf(`{"cpf":"12345678900","id_vaga":"` + validUUID + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/candidaturas", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCandidaturaHandler_Create_AsUser_Success(t *testing.T) {
	activeVaga := &empmodels.Vaga{
		ID:     uuid.MustParse(validUUID),
		Status: empmodels.StatusVagaPublicadoAtivo,
	}
	vagaRepo := &mockVagaForCandidaturaH{entity: activeVaga}
	candRepo := &mockCandidaturaRepoH{exists: false}
	r := setupCandidaturaRouter(candRepo, vagaRepo, "12345678900", false)
	body := bodyOf(`{"id_vaga":"` + validUUID + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/candidaturas", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCandidaturaHandler_Create_Unauthorized(t *testing.T) {
	vagaRepo := &mockVagaForCandidaturaH{}
	candRepo := &mockCandidaturaRepoH{}
	r := setupCandidaturaRouter(candRepo, vagaRepo, "", false) // no CPF in context
	body := bodyOf(`{"id_vaga":"` + validUUID + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/candidaturas", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestCandidaturaHandler_Create_BadJSON(t *testing.T) {
	vagaRepo := &mockVagaForCandidaturaH{}
	candRepo := &mockCandidaturaRepoH{}
	r := setupCandidaturaRouter(candRepo, vagaRepo, "12345678900", false)
	body := bodyOf(`{bad}`)
	req := httptest.NewRequest(http.MethodPost, "/candidaturas", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests: List
// ──────────────────────────────────────────────────────────────────────────────

func TestCandidaturaHandler_List_Success(t *testing.T) {
	candRepo := &mockCandidaturaRepoH{
		listItems: []*empmodels.Candidatura{{CPF: "12345678900"}},
		listTotal: 1,
	}
	vagaRepo := &mockVagaForCandidaturaH{}
	r := setupCandidaturaRouter(candRepo, vagaRepo, "admin", true)
	req := httptest.NewRequest(http.MethodGet, "/candidaturas", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestCandidaturaHandler_List_ServiceError(t *testing.T) {
	candRepo := &mockCandidaturaRepoH{err: errTest}
	vagaRepo := &mockVagaForCandidaturaH{}
	r := setupCandidaturaRouter(candRepo, vagaRepo, "admin", true)
	req := httptest.NewRequest(http.MethodGet, "/candidaturas", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests: ListByCPF
// ──────────────────────────────────────────────────────────────────────────────

func TestCandidaturaHandler_ListByCPF_AsOwner(t *testing.T) {
	candRepo := &mockCandidaturaRepoH{listItems: []*empmodels.Candidatura{}, listTotal: 0}
	vagaRepo := &mockVagaForCandidaturaH{}
	r := setupCandidaturaRouter(candRepo, vagaRepo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/candidaturas/usuario/12345678900", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCandidaturaHandler_ListByCPF_Unauthorized(t *testing.T) {
	candRepo := &mockCandidaturaRepoH{}
	vagaRepo := &mockVagaForCandidaturaH{}
	r := setupCandidaturaRouter(candRepo, vagaRepo, "", false)
	req := httptest.NewRequest(http.MethodGet, "/candidaturas/usuario/12345678900", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestCandidaturaHandler_ListByCPF_Forbidden(t *testing.T) {
	candRepo := &mockCandidaturaRepoH{}
	vagaRepo := &mockVagaForCandidaturaH{}
	r := setupCandidaturaRouter(candRepo, vagaRepo, "99999999999", false)
	req := httptest.NewRequest(http.MethodGet, "/candidaturas/usuario/12345678900", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests: GetByID
// ──────────────────────────────────────────────────────────────────────────────

func TestCandidaturaHandler_GetByID_AsOwner(t *testing.T) {
	id := uuid.MustParse(validUUID)
	candRepo := &mockCandidaturaRepoH{entity: &empmodels.Candidatura{ID: id, CPF: "12345678900"}}
	vagaRepo := &mockVagaForCandidaturaH{}
	r := setupCandidaturaRouter(candRepo, vagaRepo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/candidaturas/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestCandidaturaHandler_GetByID_NotFound(t *testing.T) {
	candRepo := &mockCandidaturaRepoH{entity: nil}
	vagaRepo := &mockVagaForCandidaturaH{}
	r := setupCandidaturaRouter(candRepo, vagaRepo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/candidaturas/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestCandidaturaHandler_GetByID_InvalidID(t *testing.T) {
	candRepo := &mockCandidaturaRepoH{}
	vagaRepo := &mockVagaForCandidaturaH{}
	r := setupCandidaturaRouter(candRepo, vagaRepo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/candidaturas/bad-id", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCandidaturaHandler_GetByID_Forbidden(t *testing.T) {
	id := uuid.MustParse(validUUID)
	candRepo := &mockCandidaturaRepoH{entity: &empmodels.Candidatura{ID: id, CPF: "12345678900"}}
	vagaRepo := &mockVagaForCandidaturaH{}
	// Different user CPF
	r := setupCandidaturaRouter(candRepo, vagaRepo, "99999999999", false)
	req := httptest.NewRequest(http.MethodGet, "/candidaturas/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests: Delete
// ──────────────────────────────────────────────────────────────────────────────

func TestCandidaturaHandler_Delete_AsOwner(t *testing.T) {
	id := uuid.MustParse(validUUID)
	candRepo := &mockCandidaturaRepoH{entity: &empmodels.Candidatura{ID: id, CPF: "12345678900"}}
	vagaRepo := &mockVagaForCandidaturaH{}
	r := setupCandidaturaRouter(candRepo, vagaRepo, "12345678900", false)
	req := httptest.NewRequest(http.MethodDelete, "/candidaturas/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestCandidaturaHandler_Delete_NotFound(t *testing.T) {
	candRepo := &mockCandidaturaRepoH{entity: nil}
	vagaRepo := &mockVagaForCandidaturaH{}
	r := setupCandidaturaRouter(candRepo, vagaRepo, "12345678900", false)
	req := httptest.NewRequest(http.MethodDelete, "/candidaturas/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestCandidaturaHandler_Delete_InvalidID(t *testing.T) {
	candRepo := &mockCandidaturaRepoH{}
	vagaRepo := &mockVagaForCandidaturaH{}
	r := setupCandidaturaRouter(candRepo, vagaRepo, "12345678900", false)
	req := httptest.NewRequest(http.MethodDelete, "/candidaturas/bad-id", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests: Approve / Reject
// ──────────────────────────────────────────────────────────────────────────────

func TestCandidaturaHandler_Approve_InvalidID(t *testing.T) {
	candRepo := &mockCandidaturaRepoH{}
	vagaRepo := &mockVagaForCandidaturaH{}
	r := setupCandidaturaRouter(candRepo, vagaRepo, "admin", true)
	req := httptest.NewRequest(http.MethodPut, "/candidaturas/bad-id/approve", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCandidaturaHandler_Reject_InvalidID(t *testing.T) {
	candRepo := &mockCandidaturaRepoH{}
	vagaRepo := &mockVagaForCandidaturaH{}
	r := setupCandidaturaRouter(candRepo, vagaRepo, "admin", true)
	req := httptest.NewRequest(http.MethodPut, "/candidaturas/bad-id/reject", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests: BulkUpdateStatus / BulkUpdateEtapa
// ──────────────────────────────────────────────────────────────────────────────

func TestCandidaturaHandler_BulkUpdateStatus_BadJSON(t *testing.T) {
	candRepo := &mockCandidaturaRepoH{}
	vagaRepo := &mockVagaForCandidaturaH{}
	r := setupCandidaturaRouter(candRepo, vagaRepo, "admin", true)
	body := bodyOf(`{bad}`)
	req := httptest.NewRequest(http.MethodPut, "/candidaturas/bulk-status", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCandidaturaHandler_BulkUpdateEtapa_BadJSON(t *testing.T) {
	candRepo := &mockCandidaturaRepoH{}
	vagaRepo := &mockVagaForCandidaturaH{}
	r := setupCandidaturaRouter(candRepo, vagaRepo, "admin", true)
	body := bodyOf(`{bad}`)
	req := httptest.NewRequest(http.MethodPut, "/candidaturas/bulk-etapa", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests: UpdateEtapa
// ──────────────────────────────────────────────────────────────────────────────

func TestCandidaturaHandler_UpdateEtapa_InvalidID(t *testing.T) {
	candRepo := &mockCandidaturaRepoH{}
	vagaRepo := &mockVagaForCandidaturaH{}
	r := setupCandidaturaRouter(candRepo, vagaRepo, "admin", true)
	body := bodyOf(`{"id_etapa_atual":"` + validUUID + `"}`)
	req := httptest.NewRequest(http.MethodPut, "/candidaturas/bad-id/etapa", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
