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
	id := uuid.MustParse(validUUID)
	vagaRepo := &mockVagaRepoH{entity: &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaPublicadoAtivo}}
	empresaRepo := &mockEmpresaRepoForVaga{}
	r := setupVagaRouterWithRoles(vagaRepo, empresaRepo, []string{"go:empregabilidade:editor_sem_curadoria"})
	body := bodyOf(`{"titulo":"Dev Editor"}`)
	req := httptest.NewRequest(http.MethodPut, "/vagas/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for empregabilidade:editor_sem_curadoria, got %d: %s", w.Code, w.Body.String())
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
	vagaRepo := &mockVagaRepoH{entity: &empmodels.Vaga{
		ID:     id,
		Titulo: "Analista de TI",
		Status: empmodels.StatusVagaPublicadoAtivo,
	}}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas/slug/analista-de-ti-f3d2367597e5", nil)
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
	}}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)
	req := httptest.NewRequest(http.MethodGet, "/public/vagas/slug/vaga-em-edicao-f3d2367597e5", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for non-published vaga, got %d", w.Code)
	}
}
