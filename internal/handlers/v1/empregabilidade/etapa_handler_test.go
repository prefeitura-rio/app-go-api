package empregabilidade_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	handlers "github.com/prefeitura-rio/app-go-api/internal/handlers/v1/empregabilidade"
	empmodels "github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
)

// ──────────────────────────────────────────────────────────────────────────────
// Mock EtapaRepository
// ──────────────────────────────────────────────────────────────────────────────

type mockEtapaRepoH struct {
	entity    *empmodels.Etapa
	listItems []*empmodels.Etapa
	err       error
}

func (m *mockEtapaRepoH) Create(_ context.Context, e *empmodels.Etapa) (uuid.UUID, error) {
	if m.err != nil {
		return uuid.Nil, m.err
	}
	return uuid.New(), nil
}

func (m *mockEtapaRepoH) GetByID(_ context.Context, _ uuid.UUID) (*empmodels.Etapa, error) {
	return m.entity, m.err
}

func (m *mockEtapaRepoH) Update(_ context.Context, _ *empmodels.Etapa) error {
	return m.err
}

func (m *mockEtapaRepoH) Delete(_ context.Context, _ uuid.UUID) error {
	return m.err
}

func (m *mockEtapaRepoH) ListByVaga(_ context.Context, _ uuid.UUID) ([]*empmodels.Etapa, error) {
	return m.listItems, m.err
}

func (m *mockEtapaRepoH) DeleteByVaga(_ context.Context, _ uuid.UUID) error {
	return m.err
}

func setupEtapaRouter(repo services.EtapaRepositoryInterface) *gin.Engine {
	r := gin.New()
	svc := services.NewEtapaServiceWithInterface(repo)
	h := handlers.NewEtapaHandler(svc)
	r.POST("/vagas/:id/etapas", h.Create)
	r.GET("/vagas/:id/etapas", h.ListByVaga)
	r.GET("/vagas/:id/etapas/:etapaId", h.GetByID)
	r.PUT("/vagas/:id/etapas/:etapaId", h.Update)
	r.DELETE("/vagas/:id/etapas/:etapaId", h.Delete)
	return r
}

func TestEtapaHandler_Create_Success(t *testing.T) {
	repo := &mockEtapaRepoH{}
	r := setupEtapaRouter(repo)
	body := bodyOf(`{"nome":"Entrevista Técnica"}`)
	req := httptest.NewRequest(http.MethodPost, "/vagas/"+validUUID+"/etapas", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestEtapaHandler_Create_InvalidVagaID(t *testing.T) {
	repo := &mockEtapaRepoH{}
	r := setupEtapaRouter(repo)
	body := bodyOf(`{"nome":"Entrevista"}`)
	req := httptest.NewRequest(http.MethodPost, "/vagas/not-a-uuid/etapas", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestEtapaHandler_Create_BadJSON(t *testing.T) {
	repo := &mockEtapaRepoH{}
	r := setupEtapaRouter(repo)
	// ShouldBindJSON with strict mode won't fail on empty body for struct without required tags
	// Use a body that is definitely invalid JSON
	body := bodyOf(`{invalid}`)
	req := httptest.NewRequest(http.MethodPost, "/vagas/"+validUUID+"/etapas", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestEtapaHandler_Create_ServiceError(t *testing.T) {
	repo := &mockEtapaRepoH{err: errTest}
	r := setupEtapaRouter(repo)
	body := bodyOf(`{"nome":"Entrevista"}`)
	req := httptest.NewRequest(http.MethodPost, "/vagas/"+validUUID+"/etapas", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestEtapaHandler_Create_UniqueViolation_Returns409(t *testing.T) {
	repo := &mockEtapaRepoH{err: uniqueViolationErr("etapas_id_vaga_ordem_key")}
	r := setupEtapaRouter(repo)
	body := bodyOf(`{"nome":"Entrevista"}`)
	req := httptest.NewRequest(http.MethodPost, "/vagas/"+validUUID+"/etapas", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d: %s", w.Code, w.Body.String())
	}
}

func TestEtapaHandler_Create_NotNullViolation_Returns400(t *testing.T) {
	repo := &mockEtapaRepoH{err: notNullViolationErr("titulo")}
	r := setupEtapaRouter(repo)
	body := bodyOf(`{"nome":"Entrevista"}`)
	req := httptest.NewRequest(http.MethodPost, "/vagas/"+validUUID+"/etapas", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestEtapaHandler_ListByVaga_Success(t *testing.T) {
	repo := &mockEtapaRepoH{
		listItems: []*empmodels.Etapa{{Titulo: "Triagem"}},
	}
	r := setupEtapaRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/vagas/"+validUUID+"/etapas", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestEtapaHandler_ListByVaga_InvalidVagaID(t *testing.T) {
	repo := &mockEtapaRepoH{}
	r := setupEtapaRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/vagas/bad-uuid/etapas", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestEtapaHandler_ListByVaga_ServiceError(t *testing.T) {
	repo := &mockEtapaRepoH{err: errTest}
	r := setupEtapaRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/vagas/"+validUUID+"/etapas", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestEtapaHandler_GetByID_Found(t *testing.T) {
	id := uuid.MustParse(validUUID)
	repo := &mockEtapaRepoH{entity: &empmodels.Etapa{ID: id, Titulo: "Triagem"}}
	r := setupEtapaRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/vagas/"+validUUID+"/etapas/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestEtapaHandler_GetByID_NotFound(t *testing.T) {
	repo := &mockEtapaRepoH{entity: nil}
	r := setupEtapaRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/vagas/"+validUUID+"/etapas/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestEtapaHandler_GetByID_InvalidID(t *testing.T) {
	repo := &mockEtapaRepoH{}
	r := setupEtapaRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/vagas/"+validUUID+"/etapas/bad-uuid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestEtapaHandler_GetByID_ServiceError(t *testing.T) {
	repo := &mockEtapaRepoH{err: errTest}
	r := setupEtapaRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/vagas/"+validUUID+"/etapas/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestEtapaHandler_Update_Success(t *testing.T) {
	repo := &mockEtapaRepoH{}
	r := setupEtapaRouter(repo)
	body := bodyOf(`{"nome":"Etapa Final"}`)
	req := httptest.NewRequest(http.MethodPut, "/vagas/"+validUUID+"/etapas/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestEtapaHandler_Update_InvalidVagaID(t *testing.T) {
	repo := &mockEtapaRepoH{}
	r := setupEtapaRouter(repo)
	body := bodyOf(`{"nome":"Etapa"}`)
	req := httptest.NewRequest(http.MethodPut, "/vagas/bad-id/etapas/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestEtapaHandler_Update_InvalidEtapaID(t *testing.T) {
	repo := &mockEtapaRepoH{}
	r := setupEtapaRouter(repo)
	body := bodyOf(`{"nome":"Etapa"}`)
	req := httptest.NewRequest(http.MethodPut, "/vagas/"+validUUID+"/etapas/bad-id", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestEtapaHandler_Update_ServiceError(t *testing.T) {
	repo := &mockEtapaRepoH{err: errTest}
	r := setupEtapaRouter(repo)
	body := bodyOf(`{"nome":"Etapa"}`)
	req := httptest.NewRequest(http.MethodPut, "/vagas/"+validUUID+"/etapas/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestEtapaHandler_Delete_Success(t *testing.T) {
	repo := &mockEtapaRepoH{}
	r := setupEtapaRouter(repo)
	req := httptest.NewRequest(http.MethodDelete, "/vagas/"+validUUID+"/etapas/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestEtapaHandler_Delete_InvalidID(t *testing.T) {
	repo := &mockEtapaRepoH{}
	r := setupEtapaRouter(repo)
	req := httptest.NewRequest(http.MethodDelete, "/vagas/"+validUUID+"/etapas/bad-id", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestEtapaHandler_Delete_ServiceError(t *testing.T) {
	repo := &mockEtapaRepoH{err: errTest}
	r := setupEtapaRouter(repo)
	req := httptest.NewRequest(http.MethodDelete, "/vagas/"+validUUID+"/etapas/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}
