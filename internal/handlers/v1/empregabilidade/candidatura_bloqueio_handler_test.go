package empregabilidade_test

import (
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

type mockCandidaturaBloqueioRepoH struct {
	listItems []*empmodels.CandidaturaBloqueio
	listTotal int
	err       error
}

func (m *mockCandidaturaBloqueioRepoH) Create(_ context.Context, _ *empmodels.CandidaturaBloqueio) (uuid.UUID, error) {
	if m.err != nil {
		return uuid.Nil, m.err
	}
	return uuid.New(), nil
}

func (m *mockCandidaturaBloqueioRepoH) List(_ context.Context, _ map[string]interface{}, _, _ int) ([]*empmodels.CandidaturaBloqueio, int, error) {
	return m.listItems, m.listTotal, m.err
}

func setupCandidaturaBloqueioRouter(repo services.CandidaturaBloqueioRepositoryInterface) *gin.Engine {
	r := gin.New()
	svc := services.NewCandidaturaBloqueioServiceWithInterface(repo)
	h := handlers.NewCandidaturaBloqueioHandler(svc)
	r.POST("/candidatura-bloqueios", h.Create)
	r.GET("/candidatura-bloqueios", h.List)
	return r
}

func TestCandidaturaBloqueioHandler_Create_Success(t *testing.T) {
	repo := &mockCandidaturaBloqueioRepoH{}
	r := setupCandidaturaBloqueioRouter(repo)
	body := bodyOf(`{"cpf":"12345678900","id_vaga":"` + validUUID + `","criterios_nao_atendidos":["idade_minima"]}`)
	req := httptest.NewRequest(http.MethodPost, "/candidatura-bloqueios", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCandidaturaBloqueioHandler_Create_BadJSON(t *testing.T) {
	repo := &mockCandidaturaBloqueioRepoH{}
	r := setupCandidaturaBloqueioRouter(repo)
	body := bodyOf(`{bad}`)
	req := httptest.NewRequest(http.MethodPost, "/candidatura-bloqueios", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCandidaturaBloqueioHandler_Create_ServiceError(t *testing.T) {
	repo := &mockCandidaturaBloqueioRepoH{err: fmt.Errorf("db error")}
	r := setupCandidaturaBloqueioRouter(repo)
	body := bodyOf(`{"cpf":"12345678900","id_vaga":"` + validUUID + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/candidatura-bloqueios", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusCreated {
		t.Error("expected error status, got 201")
	}
}

func TestCandidaturaBloqueioHandler_List_Success(t *testing.T) {
	repo := &mockCandidaturaBloqueioRepoH{
		listItems: []*empmodels.CandidaturaBloqueio{{ID: uuid.New(), CPF: "12345678900", IDVaga: uuid.New()}},
		listTotal: 1,
	}
	r := setupCandidaturaBloqueioRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/candidatura-bloqueios", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCandidaturaBloqueioHandler_List_WithFilters(t *testing.T) {
	repo := &mockCandidaturaBloqueioRepoH{
		listItems: []*empmodels.CandidaturaBloqueio{{ID: uuid.New(), CPF: "12345678900", IDVaga: uuid.New()}},
		listTotal: 1,
	}
	r := setupCandidaturaBloqueioRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/candidatura-bloqueios?cpf=12345678900&id_vaga="+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCandidaturaBloqueioHandler_List_InvalidVagaID(t *testing.T) {
	repo := &mockCandidaturaBloqueioRepoH{}
	r := setupCandidaturaBloqueioRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/candidatura-bloqueios?id_vaga=bad-id", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCandidaturaBloqueioHandler_List_PaginationClamping(t *testing.T) {
	repo := &mockCandidaturaBloqueioRepoH{}
	r := setupCandidaturaBloqueioRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/candidatura-bloqueios?page=0&pageSize=99999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCandidaturaBloqueioHandler_List_ServiceError(t *testing.T) {
	repo := &mockCandidaturaBloqueioRepoH{err: fmt.Errorf("db error")}
	r := setupCandidaturaBloqueioRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/candidatura-bloqueios", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}
