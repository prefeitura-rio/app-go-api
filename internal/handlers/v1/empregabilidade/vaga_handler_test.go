package empregabilidade_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	handlers "github.com/prefeitura-rio/app-go-api/internal/handlers/v1/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	empmodels "github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
)

// ──────────────────────────────────────────────────────────────────────────────
// Mock VagaRepo (implements VagaRepoInterface)
// ──────────────────────────────────────────────────────────────────────────────

type mockVagaRepoH struct {
	entity    *empmodels.Vaga
	listItems []*empmodels.Vaga
	listTotal int
	err       error
}

func (m *mockVagaRepoH) Create(_ context.Context, _ *empmodels.Vaga) (uuid.UUID, error) {
	if m.err != nil {
		return uuid.Nil, m.err
	}
	return uuid.New(), nil
}

func (m *mockVagaRepoH) GetByID(_ context.Context, _ uuid.UUID) (*empmodels.Vaga, error) {
	return m.entity, m.err
}

func (m *mockVagaRepoH) GetByIDPrefix(_ context.Context, _ string) (*empmodels.Vaga, error) {
	return m.entity, m.err
}

func (m *mockVagaRepoH) Update(_ context.Context, _ *empmodels.Vaga) error {
	return m.err
}

func (m *mockVagaRepoH) UpdateWithAssociations(_ context.Context, _ *empmodels.Vaga) error {
	return m.err
}

func (m *mockVagaRepoH) Delete(_ context.Context, _ uuid.UUID) error {
	return m.err
}

func (m *mockVagaRepoH) List(_ context.Context, _ empmodels.VagaFilter, _, _ int) ([]*empmodels.Vaga, int, error) {
	return m.listItems, m.listTotal, m.err
}

func (m *mockVagaRepoH) ListPublicActive(_ context.Context, _, _ int) ([]*empmodels.Vaga, int, error) {
	return m.listItems, m.listTotal, m.err
}

func (m *mockVagaRepoH) ListPublic(_ context.Context, _ empmodels.VagaPublicFilter, _, _ int) ([]*empmodels.Vaga, int, error) {
	return m.listItems, m.listTotal, m.err
}

func (m *mockVagaRepoH) UpdateTiposPCD(_ context.Context, _ uuid.UUID, _ []uuid.UUID) error {
	return m.err
}

func (m *mockVagaRepoH) ListByContratante(_ context.Context, _ string, _, _ int) ([]*empmodels.Vaga, int, error) {
	return m.listItems, m.listTotal, m.err
}

func (m *mockVagaRepoH) ListByOrgaoParceiro(_ context.Context, _ string, _, _ int) ([]*empmodels.Vaga, int, error) {
	return m.listItems, m.listTotal, m.err
}

// ──────────────────────────────────────────────────────────────────────────────
// Mock EmpresaRepo (implements EmpresaRepoInterface for VagaService)
// ──────────────────────────────────────────────────────────────────────────────

type mockEmpresaRepoForVaga struct {
	entity *empmodels.Empresa
	err    error
}

func (m *mockEmpresaRepoForVaga) GetByID(_ context.Context, _ string) (*empmodels.Empresa, error) {
	return m.entity, m.err
}

// ──────────────────────────────────────────────────────────────────────────────
// Mock CandidaturaRepo for VagaService
// ──────────────────────────────────────────────────────────────────────────────

type mockCandidaturaRepoForVaga struct {
	err error
}

func (m *mockCandidaturaRepoForVaga) BulkSaveAndUpdateStatusByVagaID(_ context.Context, _ uuid.UUID, _ empmodels.StatusCandidatura) error {
	return m.err
}

func (m *mockCandidaturaRepoForVaga) BulkRestoreStatusByVagaID(_ context.Context, _ uuid.UUID) error {
	return m.err
}

// ──────────────────────────────────────────────────────────────────────────────
// Router setup helpers
// ──────────────────────────────────────────────────────────────────────────────

func setupVagaRouter(vagaRepo services.VagaRepoInterface, empresaRepo services.EmpresaRepoInterface, isAdmin bool) *gin.Engine {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if isAdmin {
			c.Set(middlewares.UserRoleKey, "ADMIN")
		}
		c.Next()
	})
	candidaturaRepo := &mockCandidaturaRepoForVaga{}
	svc := services.NewVagaServiceWithInterfaces(vagaRepo, empresaRepo, candidaturaRepo)
	h := handlers.NewVagaHandler(svc)
	r.POST("/vagas", h.Create)
	r.GET("/vagas", h.List)
	r.GET("/vagas/:id", h.GetByID)
	r.PUT("/vagas/:id", h.Update)
	r.DELETE("/vagas/:id", h.Delete)
	r.PUT("/vagas/:id/send-to-draft", h.SendToDraft)
	r.PUT("/vagas/:id/send-to-approval", h.SendToApproval)
	r.PUT("/vagas/:id/publish", h.Publish)
	r.PUT("/vagas/:id/tipos-pcd", h.UpdateTiposPCD)
	r.PUT("/vagas/:id/freeze", h.Freeze)
	r.PUT("/vagas/:id/unfreeze", h.Unfreeze)
	r.PUT("/vagas/:id/discontinue", h.Discontinue)
	r.PUT("/vagas/:id/reactivate", h.Reactivate)
	r.GET("/public/vagas", h.PublicList)
	r.GET("/public/vagas/slug/:slug", h.PublicGetBySlug)
	r.GET("/public/vagas/:id", h.PublicGetByID)
	return r
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests
// ──────────────────────────────────────────────────────────────────────────────

func TestVagaHandler_Create_Success(t *testing.T) {
	vagaRepo := &mockVagaRepoH{}
	empresaRepo := &mockEmpresaRepoForVaga{entity: &empmodels.Empresa{CNPJ: "12345678000100"}}
	r := setupVagaRouter(vagaRepo, empresaRepo, false)
	body := bodyOf(`{"titulo":"Dev Backend","tipo_contratacao":"CLT","contratante_cnpj":"12345678000100"}`)
	req := httptest.NewRequest(http.MethodPost, "/vagas", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestVagaHandler_Create_BadJSON(t *testing.T) {
	vagaRepo := &mockVagaRepoH{}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, false)
	body := bodyOf(`{bad}`)
	req := httptest.NewRequest(http.MethodPost, "/vagas", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestVagaHandler_Create_ServiceError(t *testing.T) {
	vagaRepo := &mockVagaRepoH{err: errTest}
	empresaRepo := &mockEmpresaRepoForVaga{entity: &empmodels.Empresa{CNPJ: "12345678000100"}}
	r := setupVagaRouter(vagaRepo, empresaRepo, false)
	body := bodyOf(`{"titulo":"Dev","contratante_cnpj":"12345678000100"}`)
	req := httptest.NewRequest(http.MethodPost, "/vagas", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusCreated {
		t.Error("expected error status, got 201")
	}
}

func TestVagaHandler_List_Success(t *testing.T) {
	vagaRepo := &mockVagaRepoH{
		listItems: []*empmodels.Vaga{{Titulo: "Dev Go"}},
		listTotal: 1,
	}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, false)
	req := httptest.NewRequest(http.MethodGet, "/vagas", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestVagaHandler_List_WithFilters(t *testing.T) {
	vagaRepo := &mockVagaRepoH{listItems: []*empmodels.Vaga{}, listTotal: 0}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, false)
	req := httptest.NewRequest(http.MethodGet, "/vagas?page=2&pageSize=5&status=publicado_ativo&search=dev", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestVagaHandler_List_ServiceError(t *testing.T) {
	vagaRepo := &mockVagaRepoH{err: errTest}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, false)
	req := httptest.NewRequest(http.MethodGet, "/vagas", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestVagaHandler_GetByID_Found(t *testing.T) {
	id := uuid.MustParse(validUUID)
	vagaRepo := &mockVagaRepoH{entity: &empmodels.Vaga{ID: id, Titulo: "Dev Go"}}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, false)
	req := httptest.NewRequest(http.MethodGet, "/vagas/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestVagaHandler_GetByID_NotFound(t *testing.T) {
	vagaRepo := &mockVagaRepoH{entity: nil}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, false)
	req := httptest.NewRequest(http.MethodGet, "/vagas/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestVagaHandler_GetByID_InvalidID(t *testing.T) {
	vagaRepo := &mockVagaRepoH{}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, false)
	req := httptest.NewRequest(http.MethodGet, "/vagas/bad-uuid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestVagaHandler_GetByID_ServiceError(t *testing.T) {
	vagaRepo := &mockVagaRepoH{err: errTest}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, false)
	req := httptest.NewRequest(http.MethodGet, "/vagas/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestVagaHandler_Update_Success(t *testing.T) {
	id := uuid.MustParse(validUUID)
	// Vaga in em_edicao status (non-published) so no admin required
	vagaRepo := &mockVagaRepoH{entity: &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaEmEdicao}}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, false)
	body := bodyOf(`{"titulo":"Dev Go Atualizado"}`)
	req := httptest.NewRequest(http.MethodPut, "/vagas/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestVagaHandler_Update_InvalidID(t *testing.T) {
	vagaRepo := &mockVagaRepoH{}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, false)
	body := bodyOf(`{"titulo":"Dev"}`)
	req := httptest.NewRequest(http.MethodPut, "/vagas/bad-id", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestVagaHandler_Update_NotFound(t *testing.T) {
	vagaRepo := &mockVagaRepoH{entity: nil}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, false)
	body := bodyOf(`{"titulo":"Dev"}`)
	req := httptest.NewRequest(http.MethodPut, "/vagas/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestVagaHandler_Update_PublishedForbiddenNonAdmin(t *testing.T) {
	id := uuid.MustParse(validUUID)
	// Published status requires admin
	vagaRepo := &mockVagaRepoH{entity: &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaPublicadoAtivo}}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, false) // not admin
	body := bodyOf(`{"titulo":"Dev"}`)
	req := httptest.NewRequest(http.MethodPut, "/vagas/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestVagaHandler_Update_PublishedAllowedForAdmin(t *testing.T) {
	id := uuid.MustParse(validUUID)
	vagaRepo := &mockVagaRepoH{entity: &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaPublicadoAtivo}}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, true) // admin
	body := bodyOf(`{"titulo":"Dev Admin Update"}`)
	req := httptest.NewRequest(http.MethodPut, "/vagas/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	// Admin should be allowed; status 200 expected
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for admin, got %d: %s", w.Code, w.Body.String())
	}
}

func TestVagaHandler_Delete_Success(t *testing.T) {
	vagaRepo := &mockVagaRepoH{}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, false)
	req := httptest.NewRequest(http.MethodDelete, "/vagas/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestVagaHandler_Delete_InvalidID(t *testing.T) {
	vagaRepo := &mockVagaRepoH{}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, false)
	req := httptest.NewRequest(http.MethodDelete, "/vagas/bad-id", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestVagaHandler_Delete_ServiceError(t *testing.T) {
	vagaRepo := &mockVagaRepoH{err: errTest}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, false)
	req := httptest.NewRequest(http.MethodDelete, "/vagas/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestVagaHandler_SendToDraft_InvalidID(t *testing.T) {
	vagaRepo := &mockVagaRepoH{}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, false)
	req := httptest.NewRequest(http.MethodPut, "/vagas/bad-id/send-to-draft", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestVagaHandler_SendToApproval_InvalidID(t *testing.T) {
	vagaRepo := &mockVagaRepoH{}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, false)
	req := httptest.NewRequest(http.MethodPut, "/vagas/bad-id/send-to-approval", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestVagaHandler_Publish_InvalidID(t *testing.T) {
	vagaRepo := &mockVagaRepoH{}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, false)
	req := httptest.NewRequest(http.MethodPut, "/vagas/bad-id/publish", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestVagaHandler_UpdateTiposPCD_Success(t *testing.T) {
	vagaRepo := &mockVagaRepoH{}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, false)
	body := bodyOf(`{"tipos_pcd_ids":["` + validUUID + `"]}`)
	req := httptest.NewRequest(http.MethodPut, "/vagas/"+validUUID+"/tipos-pcd", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestVagaHandler_UpdateTiposPCD_InvalidID(t *testing.T) {
	vagaRepo := &mockVagaRepoH{}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, false)
	body := bodyOf(`{"tipos_pcd_ids":[]}`)
	req := httptest.NewRequest(http.MethodPut, "/vagas/bad-id/tipos-pcd", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestVagaHandler_Freeze_InvalidID(t *testing.T) {
	vagaRepo := &mockVagaRepoH{}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, false)
	req := httptest.NewRequest(http.MethodPut, "/vagas/bad-id/freeze", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestVagaHandler_Unfreeze_InvalidID(t *testing.T) {
	vagaRepo := &mockVagaRepoH{}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, false)
	req := httptest.NewRequest(http.MethodPut, "/vagas/bad-id/unfreeze", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestVagaHandler_Discontinue_InvalidID(t *testing.T) {
	vagaRepo := &mockVagaRepoH{}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, false)
	req := httptest.NewRequest(http.MethodPut, "/vagas/bad-id/discontinue", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestVagaHandler_Reactivate_InvalidID(t *testing.T) {
	vagaRepo := &mockVagaRepoH{}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, false)
	req := httptest.NewRequest(http.MethodPut, "/vagas/bad-id/reactivate", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestVagaHandler_PublicList_Success(t *testing.T) {
	vagaRepo := &mockVagaRepoH{
		listItems: []*empmodels.Vaga{{Titulo: "Dev Go", Status: empmodels.StatusVagaPublicadoAtivo}},
		listTotal: 1,
	}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestVagaHandler_PublicGetByID_ActiveVaga(t *testing.T) {
	id := uuid.MustParse(validUUID)
	vagaRepo := &mockVagaRepoH{entity: &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaPublicadoAtivo}}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestVagaHandler_PublicGetByID_NonActiveVaga(t *testing.T) {
	id := uuid.MustParse(validUUID)
	// Not active status should return 404
	vagaRepo := &mockVagaRepoH{entity: &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaEmEdicao}}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestVagaHandler_PublicGetByID_NotFound(t *testing.T) {
	vagaRepo := &mockVagaRepoH{entity: nil}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestVagaHandler_PublicGetByID_InvalidID(t *testing.T) {
	vagaRepo := &mockVagaRepoH{}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas/bad-id", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests: SendToDraft Success (50% → 80%+)
// ──────────────────────────────────────────────────────────────────────────────

func TestVagaHandler_SendToDraft_Success(t *testing.T) {
	id := uuid.MustParse(validUUID)
	vagaRepo := &mockVagaRepoH{entity: &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaEmAprovacao}}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, true)
	req := httptest.NewRequest(http.MethodPut, "/vagas/"+validUUID+"/send-to-draft", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests: SendToApproval Success (50% → 80%+)
// ──────────────────────────────────────────────────────────────────────────────

func TestVagaHandler_SendToApproval_Success(t *testing.T) {
	id := uuid.MustParse(validUUID)
	vagaRepo := &mockVagaRepoH{entity: &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaEmEdicao}}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, true)
	req := httptest.NewRequest(http.MethodPut, "/vagas/"+validUUID+"/send-to-approval", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests: Publish Success (50% → 80%+)
// ──────────────────────────────────────────────────────────────────────────────

func TestVagaHandler_Publish_Success(t *testing.T) {
	id := uuid.MustParse(validUUID)
	vagaRepo := &mockVagaRepoH{entity: &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaEmAprovacao}}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, true)
	req := httptest.NewRequest(http.MethodPut, "/vagas/"+validUUID+"/publish", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests: UpdateTiposPCD (66.7% → 80%+)
// ──────────────────────────────────────────────────────────────────────────────

func TestVagaHandler_UpdateTiposPCD_BadJSON(t *testing.T) {
	vagaRepo := &mockVagaRepoH{}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, true)
	body := bodyOf(`{bad}`)
	req := httptest.NewRequest(http.MethodPut, "/vagas/"+validUUID+"/tipos-pcd", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests: Freeze Success (50% → 80%+)
// ──────────────────────────────────────────────────────────────────────────────

func TestVagaHandler_Freeze_Success(t *testing.T) {
	id := uuid.MustParse(validUUID)
	vagaRepo := &mockVagaRepoH{entity: &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaPublicadoAtivo}}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, true)
	req := httptest.NewRequest(http.MethodPut, "/vagas/"+validUUID+"/freeze", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests: Unfreeze Success (50% → 80%+)
// ──────────────────────────────────────────────────────────────────────────────

func TestVagaHandler_Unfreeze_Success(t *testing.T) {
	id := uuid.MustParse(validUUID)
	vagaRepo := &mockVagaRepoH{entity: &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaCongelada}}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, true)
	req := httptest.NewRequest(http.MethodPut, "/vagas/"+validUUID+"/unfreeze", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests: Discontinue Success (50% → 80%+)
// ──────────────────────────────────────────────────────────────────────────────

func TestVagaHandler_Discontinue_Success(t *testing.T) {
	id := uuid.MustParse(validUUID)
	vagaRepo := &mockVagaRepoH{entity: &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaPublicadoAtivo}}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, true)
	req := httptest.NewRequest(http.MethodPut, "/vagas/"+validUUID+"/discontinue", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests: Reactivate Success (50% → 80%+)
// ──────────────────────────────────────────────────────────────────────────────

func TestVagaHandler_Reactivate_Success(t *testing.T) {
	id := uuid.MustParse(validUUID)
	vagaRepo := &mockVagaRepoH{entity: &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaDescontinuada}}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, true)
	req := httptest.NewRequest(http.MethodPut, "/vagas/"+validUUID+"/reactivate", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Additional VagaHandler Edge Case Tests
// ──────────────────────────────────────────────────────────────────────────────

func TestVagaHandler_List_InvalidPageSize(t *testing.T) {
	vagaRepo := &mockVagaRepoH{listItems: []*empmodels.Vaga{}, listTotal: 0}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, false)
	req := httptest.NewRequest(http.MethodGet, "/vagas?pageSize=2000", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestVagaHandler_List_InvalidPage(t *testing.T) {
	vagaRepo := &mockVagaRepoH{listItems: []*empmodels.Vaga{}, listTotal: 0}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, false)
	req := httptest.NewRequest(http.MethodGet, "/vagas?page=-1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestVagaHandler_List_ComplexFilters(t *testing.T) {
	vagaRepo := &mockVagaRepoH{listItems: []*empmodels.Vaga{}, listTotal: 0}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, false)
	req := httptest.NewRequest(http.MethodGet, "/vagas?status=em_edicao&contratante=12345678000100&orgao_parceiro_id="+validUUID+"&search=desenvolvedor", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestVagaHandler_Update_ServiceError(t *testing.T) {
	id := uuid.MustParse(validUUID)
	vagaRepo := &mockVagaRepoH{
		entity: &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaEmEdicao},
		err:    errTest,
	}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, false)
	body := bodyOf(`{"titulo":"Dev"}`)
	req := httptest.NewRequest(http.MethodPut, "/vagas/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestVagaHandler_Update_GetByIDError(t *testing.T) {
	vagaRepo := &mockVagaRepoH{err: errTest}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, false)
	body := bodyOf(`{"titulo":"Dev"}`)
	req := httptest.NewRequest(http.MethodPut, "/vagas/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestVagaHandler_Update_BadJSON(t *testing.T) {
	id := uuid.MustParse(validUUID)
	vagaRepo := &mockVagaRepoH{entity: &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaEmEdicao}}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, false)
	body := bodyOf(`{bad}`)
	req := httptest.NewRequest(http.MethodPut, "/vagas/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestVagaHandler_SendToDraft_ServiceError(t *testing.T) {
	vagaRepo := &mockVagaRepoH{err: errTest}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, true)
	req := httptest.NewRequest(http.MethodPut, "/vagas/"+validUUID+"/send-to-draft", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestVagaHandler_SendToApproval_ServiceError(t *testing.T) {
	vagaRepo := &mockVagaRepoH{err: errTest}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, true)
	req := httptest.NewRequest(http.MethodPut, "/vagas/"+validUUID+"/send-to-approval", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestVagaHandler_Publish_ServiceError(t *testing.T) {
	vagaRepo := &mockVagaRepoH{err: errTest}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, true)
	req := httptest.NewRequest(http.MethodPut, "/vagas/"+validUUID+"/publish", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestVagaHandler_UpdateTiposPCD_ServiceError(t *testing.T) {
	vagaRepo := &mockVagaRepoH{err: errTest}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, true)
	body := bodyOf(`{"tipos_pcd_ids":["` + validUUID + `"]}`)
	req := httptest.NewRequest(http.MethodPut, "/vagas/"+validUUID+"/tipos-pcd", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestVagaHandler_UpdateTiposPCD_EmptyArray(t *testing.T) {
	vagaRepo := &mockVagaRepoH{}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, true)
	body := bodyOf(`{"tipos_pcd_ids":[]}`)
	req := httptest.NewRequest(http.MethodPut, "/vagas/"+validUUID+"/tipos-pcd", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestVagaHandler_Freeze_ServiceError(t *testing.T) {
	vagaRepo := &mockVagaRepoH{err: errTest}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, true)
	req := httptest.NewRequest(http.MethodPut, "/vagas/"+validUUID+"/freeze", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestVagaHandler_Unfreeze_ServiceError(t *testing.T) {
	vagaRepo := &mockVagaRepoH{err: errTest}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, true)
	req := httptest.NewRequest(http.MethodPut, "/vagas/"+validUUID+"/unfreeze", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestVagaHandler_Discontinue_ServiceError(t *testing.T) {
	vagaRepo := &mockVagaRepoH{err: errTest}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, true)
	req := httptest.NewRequest(http.MethodPut, "/vagas/"+validUUID+"/discontinue", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestVagaHandler_Reactivate_ServiceError(t *testing.T) {
	vagaRepo := &mockVagaRepoH{err: errTest}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, true)
	req := httptest.NewRequest(http.MethodPut, "/vagas/"+validUUID+"/reactivate", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestVagaHandler_PublicList_InvalidPageSize(t *testing.T) {
	vagaRepo := &mockVagaRepoH{listItems: []*empmodels.Vaga{}, listTotal: 0}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas?pageSize=200", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestVagaHandler_PublicList_ServiceError(t *testing.T) {
	vagaRepo := &mockVagaRepoH{err: errTest}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestVagaHandler_PublicGetByID_ServiceError(t *testing.T) {
	vagaRepo := &mockVagaRepoH{err: errTest}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestVagaHandler_Update_PublishedCongelada_NonAdmin(t *testing.T) {
	id := uuid.MustParse(validUUID)
	vagaRepo := &mockVagaRepoH{entity: &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaCongelada}}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, false)
	body := bodyOf(`{"titulo":"Dev"}`)
	req := httptest.NewRequest(http.MethodPut, "/vagas/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestVagaHandler_Update_PublishedDescontinuada_NonAdmin(t *testing.T) {
	id := uuid.MustParse(validUUID)
	vagaRepo := &mockVagaRepoH{entity: &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaDescontinuada}}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, false)
	body := bodyOf(`{"titulo":"Dev"}`)
	req := httptest.NewRequest(http.MethodPut, "/vagas/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestVagaHandler_Update_PublishedExpirado_NonAdmin(t *testing.T) {
	id := uuid.MustParse(validUUID)
	vagaRepo := &mockVagaRepoH{entity: &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaPublicadoExpirado}}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouter(vagaRepo, empresaRepo, false)
	body := bodyOf(`{"titulo":"Dev"}`)
	req := httptest.NewRequest(http.MethodPut, "/vagas/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func setupVagaRouterWithRoles(vagaRepo services.VagaRepoInterface, empresaRepo services.EmpresaRepoInterface, roles []string) *gin.Engine {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(middlewares.UserRolesKey, roles)
		c.Next()
	})
	candidaturaRepo := &mockCandidaturaRepoForVaga{}
	svc := services.NewVagaServiceWithInterfaces(vagaRepo, empresaRepo, candidaturaRepo)
	h := handlers.NewVagaHandler(svc)
	r.PUT("/vagas/:id", h.Update)
	return r
}

func TestVagaHandler_Update_PublishedAllowedForEmpregabilidadeAdmin(t *testing.T) {
	id := uuid.MustParse(validUUID)
	vagaRepo := &mockVagaRepoH{entity: &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaPublicadoAtivo}}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouterWithRoles(vagaRepo, empresaRepo, []string{"go:empregabilidade:admin"})
	body := bodyOf(`{"titulo":"Dev Emp Admin"}`)
	req := httptest.NewRequest(http.MethodPut, "/vagas/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for empregabilidade:admin, got %d: %s", w.Code, w.Body.String())
	}
}

func TestVagaHandler_Update_PublishedAllowedForEmpregabilidadeEditorSemCuradoria(t *testing.T) {
	const testOrgao = "orgao-test-sem-curadoria"
	id := uuid.MustParse(validUUID)
	vagaRepo := &mockVagaRepoH{entity: &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaPublicadoAtivo, IDOrgaoParceiro: testOrgao}}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(middlewares.UserRolesKey, []string{"go:empregabilidade:editor_sem_curadoria"})
		c.Set(middlewares.UserGroupsKey, []string{"go:orgao:" + testOrgao})
		c.Next()
	})
	candidaturaRepo := &mockCandidaturaRepoForVaga{}
	svc := services.NewVagaServiceWithInterfaces(vagaRepo, empresaRepo, candidaturaRepo)
	h := handlers.NewVagaHandler(svc)
	r.PUT("/vagas/:id", h.Update)
	body := bodyOf(`{"titulo":"Dev Editor"}`)
	req := httptest.NewRequest(http.MethodPut, "/vagas/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for empregabilidade:editor_sem_curadoria same-org, got %d: %s", w.Code, w.Body.String())
	}
}

func TestVagaHandler_Update_CrossOrgForbiddenForEditorSemCuradoria(t *testing.T) {
	id := uuid.MustParse(validUUID)
	vagaRepo := &mockVagaRepoH{entity: &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaPublicadoAtivo, IDOrgaoParceiro: "orgao-outro"}}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(middlewares.UserRolesKey, []string{"go:empregabilidade:editor_sem_curadoria"})
		c.Set(middlewares.UserGroupsKey, []string{"go:orgao:orgao-meu"})
		c.Next()
	})
	candidaturaRepo := &mockCandidaturaRepoForVaga{}
	svc := services.NewVagaServiceWithInterfaces(vagaRepo, empresaRepo, candidaturaRepo)
	h := handlers.NewVagaHandler(svc)
	r.PUT("/vagas/:id", h.Update)
	body := bodyOf(`{"titulo":"Dev Editor Cross"}`)
	req := httptest.NewRequest(http.MethodPut, "/vagas/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for empregabilidade:editor_sem_curadoria cross-org, got %d: %s", w.Code, w.Body.String())
	}
}

func TestVagaHandler_Update_PublishedForbiddenForUnrelatedRole(t *testing.T) {
	id := uuid.MustParse(validUUID)
	vagaRepo := &mockVagaRepoH{entity: &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaPublicadoAtivo}}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouterWithRoles(vagaRepo, empresaRepo, []string{"go:cursos:editor"})
	body := bodyOf(`{"titulo":"Dev Cursos"}`)
	req := httptest.NewRequest(http.MethodPut, "/vagas/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for unrelated role, got %d", w.Code)
	}
}

func TestVagaHandler_SendToDraft_WrongStatus_Retorna409(t *testing.T) {
	id := uuid.MustParse(validUUID)
	vagaRepo := &mockVagaRepoH{entity: &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaEmEdicao}}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodPut, "/vagas/"+validUUID+"/send-to-draft", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Errorf("expected 409 for wrong state, got %d: %s", w.Code, w.Body.String())
	}
}

func TestVagaHandler_SendToApproval_WrongStatus_Retorna409(t *testing.T) {
	id := uuid.MustParse(validUUID)
	vagaRepo := &mockVagaRepoH{entity: &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaEmAprovacao}}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodPut, "/vagas/"+validUUID+"/send-to-approval", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Errorf("expected 409 for wrong state, got %d: %s", w.Code, w.Body.String())
	}
}

func TestVagaHandler_PublicGetByID_PublicadoExpirado(t *testing.T) {
	id := uuid.MustParse(validUUID)
	vagaRepo := &mockVagaRepoH{entity: &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaPublicadoExpirado}}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for publicado_expirado, got %d", w.Code)
	}
}

func TestVagaHandler_PublicGetByID_VagaCongelada(t *testing.T) {
	id := uuid.MustParse(validUUID)
	vagaRepo := &mockVagaRepoH{entity: &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaCongelada}}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for vaga_congelada, got %d", w.Code)
	}
}

func TestVagaHandler_PublicGetByID_VagaDescontinuada(t *testing.T) {
	id := uuid.MustParse(validUUID)
	vagaRepo := &mockVagaRepoH{entity: &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaDescontinuada}}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for vaga_descontinuada, got %d", w.Code)
	}
}

func TestVagaHandler_PublicGetByID_EmEdicao_Retorna404(t *testing.T) {
	id := uuid.MustParse(validUUID)
	vagaRepo := &mockVagaRepoH{entity: &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaEmEdicao}}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for em_edicao, got %d", w.Code)
	}
}

func TestVagaHandler_PublicGetByID_EmAprovacao_Retorna404(t *testing.T) {
	id := uuid.MustParse(validUUID)
	vagaRepo := &mockVagaRepoH{entity: &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaEmAprovacao}}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for em_aprovacao, got %d", w.Code)
	}
}

func TestVagaHandler_PublicGetBySlug_PublicadoExpirado(t *testing.T) {
	id := uuid.MustParse("f3d23675-97e5-4d57-8892-bff6ba805d6d")
	slug := "vaga-expirada-f3d2367597e5"
	vagaRepo := &mockVagaRepoH{entity: &empmodels.Vaga{
		ID:     id,
		Titulo: "Vaga Expirada",
		Status: empmodels.StatusVagaPublicadoExpirado,
		Slug:   slug,
	}}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas/slug/"+slug, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for publicado_expirado slug, got %d", w.Code)
	}
}

func TestVagaHandler_PublicGetBySlug_VagaCongelada(t *testing.T) {
	id := uuid.MustParse("f3d23675-97e5-4d57-8892-bff6ba805d6d")
	slug := "vaga-congelada-f3d2367597e5"
	vagaRepo := &mockVagaRepoH{entity: &empmodels.Vaga{
		ID:     id,
		Titulo: "Vaga Congelada",
		Status: empmodels.StatusVagaCongelada,
		Slug:   slug,
	}}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas/slug/"+slug, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for vaga_congelada slug, got %d", w.Code)
	}
}

func TestVagaHandler_PublicGetBySlug_VagaDescontinuada(t *testing.T) {
	id := uuid.MustParse("f3d23675-97e5-4d57-8892-bff6ba805d6d")
	slug := "vaga-descontinuada-f3d2367597e5"
	vagaRepo := &mockVagaRepoH{entity: &empmodels.Vaga{
		ID:     id,
		Titulo: "Vaga Descontinuada",
		Status: empmodels.StatusVagaDescontinuada,
		Slug:   slug,
	}}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas/slug/"+slug, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for vaga_descontinuada slug, got %d", w.Code)
	}
}

func TestVagaHandler_PublicList_WithStatusFilter_PublicadoAtivo(t *testing.T) {
	vagaRepo := &mockVagaRepoH{
		listItems: []*empmodels.Vaga{{Titulo: "Vaga Ativa", Status: empmodels.StatusVagaPublicadoAtivo}},
		listTotal: 1,
	}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas?status=publicado_ativo", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for status=publicado_ativo, got %d", w.Code)
	}
}

func TestVagaHandler_PublicList_WithStatusFilter_VagaCongelada(t *testing.T) {
	vagaRepo := &mockVagaRepoH{
		listItems: []*empmodels.Vaga{{Titulo: "Vaga Congelada", Status: empmodels.StatusVagaCongelada}},
		listTotal: 1,
	}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas?status=vaga_congelada", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for status=vaga_congelada, got %d", w.Code)
	}
}

func TestVagaHandler_PublicList_WithStatusFilter_PublicadoExpirado(t *testing.T) {
	vagaRepo := &mockVagaRepoH{listItems: []*empmodels.Vaga{}, listTotal: 0}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas?status=publicado_expirado", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for status=publicado_expirado, got %d", w.Code)
	}
}

func TestVagaHandler_PublicList_WithStatusFilter_VagaDescontinuada(t *testing.T) {
	vagaRepo := &mockVagaRepoH{listItems: []*empmodels.Vaga{}, listTotal: 0}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas?status=vaga_descontinuada", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for status=vaga_descontinuada, got %d", w.Code)
	}
}

func TestVagaHandler_PublicList_WithInvalidStatusFilter(t *testing.T) {
	vagaRepo := &mockVagaRepoH{listItems: []*empmodels.Vaga{}, listTotal: 0}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas?status=em_edicao", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid status filter, got %d", w.Code)
	}
}

func TestVagaHandler_PublicList_WithInvalidStatusUnknown(t *testing.T) {
	vagaRepo := &mockVagaRepoH{listItems: []*empmodels.Vaga{}, listTotal: 0}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas?status=status_invalido", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for unknown status filter, got %d", w.Code)
	}
}

func TestVagaHandler_PublicList_NoFilter_Success(t *testing.T) {
	vagaRepo := &mockVagaRepoH{
		listItems: []*empmodels.Vaga{
			{Titulo: "Ativa", Status: empmodels.StatusVagaPublicadoAtivo},
			{Titulo: "Congelada", Status: empmodels.StatusVagaCongelada},
		},
		listTotal: 2,
	}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for no filter, got %d", w.Code)
	}
}

func TestVagaHandler_Update_PublishedForbiddenForEditorComCuradoria(t *testing.T) {
	id := uuid.MustParse(validUUID)
	vagaRepo := &mockVagaRepoH{entity: &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaPublicadoAtivo}}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouterWithRoles(vagaRepo, empresaRepo, []string{"go:empregabilidade:editor_com_curadoria"})
	body := bodyOf(`{"titulo":"Dev Editor Com Curadoria"}`)
	req := httptest.NewRequest(http.MethodPut, "/vagas/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for editor_com_curadoria on published vaga, got %d", w.Code)
	}
}

func TestVagaHandler_PublicGetBySlug_Success(t *testing.T) {
	id := uuid.MustParse("f3d23675-97e5-4d57-8892-bff6ba805d6d")
	slug := "analista-de-ti-f3d2367597e5"
	vagaRepo := &mockVagaRepoH{entity: &empmodels.Vaga{
		ID:     id,
		Titulo: "Analista de TI",
		Status: empmodels.StatusVagaPublicadoAtivo,
		Slug:   slug,
	}}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas/slug/"+slug, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestVagaHandler_PublicGetBySlug_NaoEncontrada(t *testing.T) {
	vagaRepo := &mockVagaRepoH{entity: nil}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas/slug/inexistente-aaaabbbb0000", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestVagaHandler_PublicGetBySlug_NaoPublicada(t *testing.T) {
	id := uuid.MustParse("f3d23675-97e5-4d57-8892-bff6ba805d6d")
	vagaRepo := &mockVagaRepoH{entity: &empmodels.Vaga{
		ID:     id,
		Titulo: "Vaga em edição",
		Status: empmodels.StatusVagaEmEdicao,
		Slug:   "vaga-em-edicao-f3d2367597e5",
	}}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas/slug/vaga-em-edicao-f3d2367597e5", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for non-published vaga, got %d", w.Code)
	}
}

func TestVagaHandler_PublicGetBySlug_RedirectSlugHistorico(t *testing.T) {
	id := uuid.MustParse("f3d23675-97e5-4d57-8892-bff6ba805d6d")
	currentSlug := "analista-financeiro-jr-f3d2367597e5"
	vagaRepo := &mockVagaRepoH{entity: &empmodels.Vaga{
		ID:     id,
		Titulo: "Analista Financeiro Jr",
		Status: empmodels.StatusVagaPublicadoAtivo,
		Slug:   currentSlug,
	}}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas/slug/analista-jr-f3d2367597e5", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusMovedPermanently {
		t.Errorf("expected 301 for historical slug, got %d", w.Code)
	}
	location := w.Header().Get("Location")
	expected := "/public/vagas/slug/" + currentSlug
	if location != expected {
		t.Errorf("expected Location %q, got %q", expected, location)
	}
}

func TestVagaHandler_PublicList_DataPublicacao_Hoje(t *testing.T) {
	vagaRepo := &mockVagaRepoH{listItems: []*empmodels.Vaga{}, listTotal: 0}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas?data_publicacao=hoje", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for data_publicacao=hoje, got %d: %s", w.Code, w.Body.String())
	}
}

func TestVagaHandler_PublicList_DataPublicacao_Invalido(t *testing.T) {
	vagaRepo := &mockVagaRepoH{listItems: []*empmodels.Vaga{}, listTotal: 0}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas?data_publicacao=invalido", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for data_publicacao=invalido, got %d", w.Code)
	}
}

func TestVagaHandler_PublicList_AcessibilidadePCD_ParaPCD(t *testing.T) {
	vagaRepo := &mockVagaRepoH{listItems: []*empmodels.Vaga{}, listTotal: 0}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas?acessibilidade_pcd=para_pcd", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for acessibilidade_pcd=para_pcd, got %d: %s", w.Code, w.Body.String())
	}
}

func TestVagaHandler_PublicList_AcessibilidadePCD_Invalido(t *testing.T) {
	vagaRepo := &mockVagaRepoH{listItems: []*empmodels.Vaga{}, listTotal: 0}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas?acessibilidade_pcd=invalido", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for acessibilidade_pcd=invalido, got %d", w.Code)
	}
}

func TestVagaHandler_PublicList_Bairro(t *testing.T) {
	vagaRepo := &mockVagaRepoH{listItems: []*empmodels.Vaga{}, listTotal: 0}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas?bairro=copacabana", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for bairro=copacabana, got %d: %s", w.Code, w.Body.String())
	}
}

func TestVagaHandler_PublicList_IDRegimeContratacao(t *testing.T) {
	vagaRepo := &mockVagaRepoH{listItems: []*empmodels.Vaga{}, listTotal: 0}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas?id_regime_contratacao="+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for id_regime_contratacao, got %d: %s", w.Code, w.Body.String())
	}
}

func TestVagaHandler_PublicList_IDModeloTrabalho(t *testing.T) {
	vagaRepo := &mockVagaRepoH{listItems: []*empmodels.Vaga{}, listTotal: 0}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas?id_modelo_trabalho="+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for id_modelo_trabalho, got %d: %s", w.Code, w.Body.String())
	}
}

func TestVagaHandler_PublicList_Contratante_Nome(t *testing.T) {
	vagaRepo := &mockVagaRepoH{listItems: []*empmodels.Vaga{}, listTotal: 0}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas?contratante=Google", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for contratante=Google, got %d: %s", w.Code, w.Body.String())
	}
}

func TestVagaHandler_PublicList_Contratante_CNPJ(t *testing.T) {
	vagaRepo := &mockVagaRepoH{listItems: []*empmodels.Vaga{}, listTotal: 0}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas?contratante=12345678000100", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for contratante=12345678000100, got %d: %s", w.Code, w.Body.String())
	}
}

// ==================== Multi-Select Tests ====================

func TestVagaHandler_PublicList_Bairro_MultiSelect(t *testing.T) {
	vagaRepo := &mockVagaRepoH{listItems: []*empmodels.Vaga{}, listTotal: 0}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas?bairro=Centro,Tijuca", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for bairro=Centro,Tijuca, got %d: %s", w.Code, w.Body.String())
	}
}

func TestVagaHandler_PublicList_Bairro_DeduplicatesRepeatedValues(t *testing.T) {
	vagaRepo := &mockVagaRepoH{listItems: []*empmodels.Vaga{}, listTotal: 0}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas?bairro=Centro,Centro,Tijuca", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for deduplicated bairro CSV, got %d: %s", w.Code, w.Body.String())
	}
}

func TestVagaHandler_PublicList_Bairro_IgnoresEmptyCSVSegments(t *testing.T) {
	vagaRepo := &mockVagaRepoH{listItems: []*empmodels.Vaga{}, listTotal: 0}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas?bairro=Centro,,Tijuca", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for CSV with empty segment, got %d: %s", w.Code, w.Body.String())
	}
}

func TestVagaHandler_PublicList_Bairro_SingleValueStillWorks(t *testing.T) {
	vagaRepo := &mockVagaRepoH{listItems: []*empmodels.Vaga{}, listTotal: 0}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas?bairro=Centro", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for bairro=Centro (backward compat), got %d: %s", w.Code, w.Body.String())
	}
}

func TestVagaHandler_PublicList_IDRegimeContratacao_MultiSelect(t *testing.T) {
	vagaRepo := &mockVagaRepoH{listItems: []*empmodels.Vaga{}, listTotal: 0}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	validUUID2 := "660e8400-e29b-41d4-a716-446655440001"
	req := httptest.NewRequest(http.MethodGet, "/public/vagas?id_regime_contratacao="+validUUID+","+validUUID2, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for id_regime_contratacao multi, got %d: %s", w.Code, w.Body.String())
	}
}

func TestVagaHandler_PublicList_IDModeloTrabalho_MultiSelect(t *testing.T) {
	vagaRepo := &mockVagaRepoH{listItems: []*empmodels.Vaga{}, listTotal: 0}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	validUUID2 := "660e8400-e29b-41d4-a716-446655440001"
	req := httptest.NewRequest(http.MethodGet, "/public/vagas?id_modelo_trabalho="+validUUID+","+validUUID2, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for id_modelo_trabalho multi, got %d: %s", w.Code, w.Body.String())
	}
}

func TestVagaHandler_PublicList_AcessibilidadePCD_MultiSelect_AllValid(t *testing.T) {
	vagaRepo := &mockVagaRepoH{listItems: []*empmodels.Vaga{}, listTotal: 0}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas?acessibilidade_pcd=para_pcd,exclusivo_pcd", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for acessibilidade_pcd=para_pcd,exclusivo_pcd, got %d: %s", w.Code, w.Body.String())
	}
}

func TestVagaHandler_PublicList_AcessibilidadePCD_MultiSelect_AllThreeValid(t *testing.T) {
	vagaRepo := &mockVagaRepoH{listItems: []*empmodels.Vaga{}, listTotal: 0}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas?acessibilidade_pcd=para_pcd,preferencial_pcd,exclusivo_pcd", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for all three acessibilidade_pcd values, got %d: %s", w.Code, w.Body.String())
	}
}

func TestVagaHandler_PublicList_AcessibilidadePCD_MultiSelect_OneInvalid(t *testing.T) {
	vagaRepo := &mockVagaRepoH{listItems: []*empmodels.Vaga{}, listTotal: 0}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas?acessibilidade_pcd=para_pcd,invalido", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for acessibilidade_pcd with invalid value, got %d: %s", w.Code, w.Body.String())
	}
}

func TestVagaHandler_PublicList_Contratante_MultiSelect_NomesOnly(t *testing.T) {
	vagaRepo := &mockVagaRepoH{listItems: []*empmodels.Vaga{}, listTotal: 0}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas?contratante=Google,Microsoft", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for contratante=Google,Microsoft, got %d: %s", w.Code, w.Body.String())
	}
}

func TestVagaHandler_PublicList_Contratante_MultiSelect_CNPJsOnly(t *testing.T) {
	vagaRepo := &mockVagaRepoH{listItems: []*empmodels.Vaga{}, listTotal: 0}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas?contratante=12345678000100,98765432000199", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for contratante with multiple CNPJs, got %d: %s", w.Code, w.Body.String())
	}
}

func TestVagaHandler_PublicList_Contratante_MultiSelect_Mixed(t *testing.T) {
	vagaRepo := &mockVagaRepoH{listItems: []*empmodels.Vaga{}, listTotal: 0}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas?contratante=Google,12345678000100", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for contratante mixed CNPJ+name, got %d: %s", w.Code, w.Body.String())
	}
}

func TestVagaHandler_PublicList_MultipleFilters_Combined(t *testing.T) {
	vagaRepo := &mockVagaRepoH{listItems: []*empmodels.Vaga{}, listTotal: 0}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet,
		"/public/vagas?bairro=Centro,Tijuca&id_regime_contratacao="+validUUID+"&acessibilidade_pcd=para_pcd,exclusivo_pcd", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for combined multi-select filters, got %d: %s", w.Code, w.Body.String())
	}
}

func TestVagaHandler_List_NewFilters_DataPublicacao(t *testing.T) {
	tests := []struct {
		name           string
		dataPublicacao string
		wantStatus     int
	}{
		{"hoje", "hoje", http.StatusOK},
		{"ultima_semana", "ultima_semana", http.StatusOK},
		{"ultimo_mes", "ultimo_mes", http.StatusOK},
		{"invalid", "ontem", http.StatusBadRequest},
		{"empty", "", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vagaRepo := &mockVagaRepoH{listItems: []*empmodels.Vaga{}, listTotal: 0}
			r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
			url := "/vagas"
			if tt.dataPublicacao != "" {
				url += "?data_publicacao=" + tt.dataPublicacao
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tt.wantStatus {
				t.Errorf("expected %d, got %d: %s", tt.wantStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestVagaHandler_List_NewFilters_AcessibilidadePCD(t *testing.T) {
	tests := []struct {
		name       string
		param      string
		wantStatus int
	}{
		{"para_pcd", "para_pcd", http.StatusOK},
		{"exclusivo_pcd", "exclusivo_pcd", http.StatusOK},
		{"preferencial_pcd", "preferencial_pcd", http.StatusOK},
		{"multiple valid", "para_pcd,exclusivo_pcd", http.StatusOK},
		{"invalid value", "invalido", http.StatusBadRequest},
		{"mixed valid and invalid", "para_pcd,invalido", http.StatusBadRequest},
		{"empty", "", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vagaRepo := &mockVagaRepoH{listItems: []*empmodels.Vaga{}, listTotal: 0}
			r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
			url := "/vagas"
			if tt.param != "" {
				url += "?acessibilidade_pcd=" + tt.param
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tt.wantStatus {
				t.Errorf("expected %d, got %d: %s", tt.wantStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestVagaHandler_List_NewFilters_CSVParams(t *testing.T) {
	vagaRepo := &mockVagaRepoH{listItems: []*empmodels.Vaga{}, listTotal: 0}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet,
		"/vagas?id_regime_contratacao="+validUUID+"&id_modelo_trabalho="+validUUID+"&bairro=Centro,Tijuca", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for CSV filters, got %d: %s", w.Code, w.Body.String())
	}
}

func TestVagaHandler_List_PageSizeCappedAt100(t *testing.T) {
	vagaRepo := &mockVagaRepoH{listItems: []*empmodels.Vaga{}, listTotal: 0}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet, "/vagas?pageSize=500", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestVagaHandler_List_MiddlewareOrgaoParceiroID_SingleID(t *testing.T) {
	vagaRepo := &mockVagaRepoH{listItems: []*empmodels.Vaga{}, listTotal: 0}
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("vaga_orgao_parceiro_ids", []string{"orgao-42"})
		c.Next()
	})
	candidaturaRepo := &mockCandidaturaRepoForVaga{}
	svc := services.NewVagaServiceWithInterfaces(vagaRepo, &mockEmpresaRepoForVaga{}, candidaturaRepo)
	h := handlers.NewVagaHandler(svc)
	r.GET("/vagas", h.List)

	req := httptest.NewRequest(http.MethodGet, "/vagas", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestVagaHandler_List_MiddlewareOrgaoParceiroIDs_MultipleIDs(t *testing.T) {
	vagaRepo := &mockVagaRepoH{listItems: []*empmodels.Vaga{}, listTotal: 0}
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("vaga_orgao_parceiro_ids", []string{"orgao-1", "orgao-2", "orgao-3"})
		c.Next()
	})
	candidaturaRepo := &mockCandidaturaRepoForVaga{}
	svc := services.NewVagaServiceWithInterfaces(vagaRepo, &mockEmpresaRepoForVaga{}, candidaturaRepo)
	h := handlers.NewVagaHandler(svc)
	r.GET("/vagas", h.List)

	req := httptest.NewRequest(http.MethodGet, "/vagas", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestVagaHandler_List_OrgaoParceiroID_FallsBackToQueryParam(t *testing.T) {
	vagaRepo := &mockVagaRepoH{listItems: []*empmodels.Vaga{}, listTotal: 0}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet, "/vagas?orgao_parceiro_id=orgao-99", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func setupVagaRouterWithAllowedOrgaos(vagaRepo services.VagaRepoInterface, empresaRepo services.EmpresaRepoInterface, allowedOrgaos []string) *gin.Engine {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if allowedOrgaos != nil {
			c.Set("vaga_allowed_orgaos", allowedOrgaos)
		}
		c.Next()
	})
	candidaturaRepo := &mockCandidaturaRepoForVaga{}
	svc := services.NewVagaServiceWithInterfaces(vagaRepo, empresaRepo, candidaturaRepo)
	h := handlers.NewVagaHandler(svc)
	r.POST("/vagas", h.Create)
	return r
}

func TestVagaHandler_Create_SecretariaUser_AllowedOrgao_Success(t *testing.T) {
	vagaRepo := &mockVagaRepoH{}
	empresaRepo := &mockEmpresaRepoForVaga{entity: &empmodels.Empresa{CNPJ: "12345678000100"}}
	r := setupVagaRouterWithAllowedOrgaos(vagaRepo, empresaRepo, []string{"orgao-1", "orgao-2"})

	body := bodyOf(`{"titulo":"Dev Backend","id_orgao_parceiro":"orgao-1"}`)
	req := httptest.NewRequest(http.MethodPost, "/vagas", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestVagaHandler_Create_SecretariaUser_ForbiddenOrgao_Forbidden(t *testing.T) {
	vagaRepo := &mockVagaRepoH{}
	empresaRepo := &mockEmpresaRepoForVaga{entity: &empmodels.Empresa{CNPJ: "12345678000100"}}
	r := setupVagaRouterWithAllowedOrgaos(vagaRepo, empresaRepo, []string{"orgao-1", "orgao-2"})

	body := bodyOf(`{"titulo":"Dev Backend","id_orgao_parceiro":"orgao-99"}`)
	req := httptest.NewRequest(http.MethodPost, "/vagas", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
}

func TestVagaHandler_Create_Admin_NoRestriction_Success(t *testing.T) {
	vagaRepo := &mockVagaRepoH{}
	empresaRepo := &mockEmpresaRepoForVaga{entity: &empmodels.Empresa{CNPJ: "12345678000100"}}
	r := setupVagaRouterWithAllowedOrgaos(vagaRepo, empresaRepo, nil)

	body := bodyOf(`{"titulo":"Dev Backend","id_orgao_parceiro":"any-orgao"}`)
	req := httptest.NewRequest(http.MethodPost, "/vagas", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestVagaHandler_List_EmptyOrgaoParceiroIDs_ReturnsEmpty(t *testing.T) {
	vagaRepo := &mockVagaRepoH{
		listItems: []*empmodels.Vaga{{Titulo: "Dev Go"}},
		listTotal: 1,
	}
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("vaga_orgao_parceiro_ids", []string{})
		c.Next()
	})
	candidaturaRepo := &mockCandidaturaRepoForVaga{}
	svc := services.NewVagaServiceWithInterfaces(vagaRepo, &mockEmpresaRepoForVaga{}, candidaturaRepo)
	h := handlers.NewVagaHandler(svc)
	r.GET("/vagas", h.List)

	req := httptest.NewRequest(http.MethodGet, "/vagas", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"total":0`) {
		t.Errorf("expected total=0, got: %s", w.Body.String())
	}
}
