package empregabilidade_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	handler "github.com/prefeitura-rio/app-go-api/internal/handlers/v1/empregabilidade"
	empmodels "github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
)

// ──────────────────────────────────────────────────────────────────────────────
// Mock para HabilidadeRepositoryInterface
// ──────────────────────────────────────────────────────────────────────────────

type mockHabilidadeRepo struct {
	err           error
	habilidade    *empmodels.Habilidade
	habilidades   []*empmodels.Habilidade
	totalHab      int64
	areaAtuacao   *empmodels.AreaAtuacao
	areasAtuacao  []*empmodels.AreaAtuacao
	totalArea     int64
	lastCreatedID int64
	notFound      bool
}

func (m *mockHabilidadeRepo) CreateHabilidade(_ context.Context, _ *empmodels.Habilidade) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	if m.lastCreatedID != 0 {
		return m.lastCreatedID, nil
	}
	return 10, nil
}

func (m *mockHabilidadeRepo) GetHabilidadeByID(_ context.Context, id int64) (*empmodels.Habilidade, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.notFound {
		return nil, nil // Retorna nil limpo
	}
	if m.habilidade != nil {
		return m.habilidade, nil
	}
	return &empmodels.Habilidade{ID: id, Nome: "Desenvolvimento Go"}, nil
}

func (m *mockHabilidadeRepo) UpdateHabilidade(_ context.Context, _ *empmodels.Habilidade) error {
	return m.err
}

func (m *mockHabilidadeRepo) DeleteHabilidade(_ context.Context, _ int64) error {
	return m.err
}

func (m *mockHabilidadeRepo) ListHabilidades(_ context.Context, _ empmodels.HabilidadeFilter, page, pageSize int) ([]*empmodels.Habilidade, int64, error) {
	if m.err != nil {
		return nil, 0, m.err
	}
	return m.habilidades, m.totalHab, nil
}

func (m *mockHabilidadeRepo) CreateAreaAtuacao(_ context.Context, _ *empmodels.AreaAtuacao) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	if m.lastCreatedID != 0 {
		return m.lastCreatedID, nil
	}
	return 20, nil
}

func (m *mockHabilidadeRepo) GetAreaAtuacaoByID(_ context.Context, id int64) (*empmodels.AreaAtuacao, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.areaAtuacao != nil {
		return m.areaAtuacao, nil
	}
	return &empmodels.AreaAtuacao{ID: id, Nome: "Tecnologia da Informação"}, nil
}

func (m *mockHabilidadeRepo) UpdateAreaAtuacao(_ context.Context, _ *empmodels.AreaAtuacao) error {
	return m.err
}

func (m *mockHabilidadeRepo) DeleteAreaAtuacao(_ context.Context, _ int64) error {
	return m.err
}

func (m *mockHabilidadeRepo) ListAreasAtuacao(_ context.Context, _ empmodels.AreaAtuacaoFilter, page, pageSize int) ([]*empmodels.AreaAtuacao, int64, error) {
	if m.err != nil {
		return nil, 0, m.err
	}
	return m.areasAtuacao, m.totalArea, nil
}

func (m *mockHabilidadeRepo) AttachAreaAtuacao(_ context.Context, _, _ int64) error {
	return m.err
}

func (m *mockHabilidadeRepo) DetachAreaAtuacao(_ context.Context, _, _ int64) error {
	return m.err
}

func (m *mockHabilidadeRepo) ReplaceAreasAtuacao(_ context.Context, _ int64, _ []int64) error {
	return m.err
}

// ──────────────────────────────────────────────────────────────────────────────
// Setup Router
// ──────────────────────────────────────────────────────────────────────────────

func setupHabilidadeRouter(hRepo services.HabilidadeRepositoryInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	hSvc := services.NewHabilidadeServiceWithInterface(hRepo)

	h := handler.NewHabilidadeHandler(hSvc, nil)

	r.POST("/habilidades", h.CreateHabilidade)
	r.GET("/habilidades", h.ListHabilidades)
	r.GET("/habilidades/:id", h.GetHabilidadeByID)
	r.PUT("/habilidades/:id", h.UpdateHabilidade)
	r.DELETE("/habilidades/:id", h.DeleteHabilidade)

	r.POST("/habilidades/areas-atuacao", h.CreateAreaAtuacao)
	r.POST("/habilidades/:id/areas-atuacao/:areaId", h.AttachAreaAtuacao)
	r.GET("/habilidades/areas-atuacao", h.ListAreasAtuacao)
	r.GET("/habilidades/areas-atuacao/:id", h.GetAreaAtuacaoByID)
	r.PUT("/habilidades/areas-atuacao/:id", h.UpdateAreaAtuacao)
	r.PUT("/habilidades/:id/areas-atuacao", h.ReplaceAreasAtuacao)
	r.DELETE("/habilidades/areas-atuacao/:id", h.DeleteAreaAtuacao)
	r.DELETE("/habilidades/:id/areas-atuacao/:areaId", h.DetachAreaAtuacao)

	return r
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests: Habilidades
// ──────────────────────────────────────────────────────────────────────────────

func TestCreateHabilidade_Success(t *testing.T) {
	repo := &mockHabilidadeRepo{lastCreatedID: 10}
	r := setupHabilidadeRouter(repo)

	body := bytes.NewBufferString(`{"nome":"Desenvolvimento Go"}`)
	req := httptest.NewRequest(http.MethodPost, "/habilidades", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCreateHabilidade_BadRequest(t *testing.T) {
	repo := &mockHabilidadeRepo{}
	r := setupHabilidadeRouter(repo)

	body := bytes.NewBufferString(`{"nome":""}`)
	req := httptest.NewRequest(http.MethodPost, "/habilidades", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
func TestCreateHabilidade_InternalServerError(t *testing.T) {
	// Configura o mock para retornar um erro proposital
	repo := &mockHabilidadeRepo{err: fmt.Errorf("erro de banco de dados")}
	r := setupHabilidadeRouter(repo)

	body := bytes.NewBufferString(`{"nome":"Desenvolvimento Go"}`)
	req := httptest.NewRequest(http.MethodPost, "/habilidades", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetHabilidadeByID_Success(t *testing.T) {
	repo := &mockHabilidadeRepo{habilidade: &empmodels.Habilidade{ID: 10, Nome: "Go"}}
	r := setupHabilidadeRouter(repo)

	req := httptest.NewRequest(http.MethodGet, "/habilidades/10", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetHabilidadeByID_InvalidID(t *testing.T) {
	repo := &mockHabilidadeRepo{}
	r := setupHabilidadeRouter(repo)

	req := httptest.NewRequest(http.MethodGet, "/habilidades/invalid-id", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetHabilidadeByID_NotFound(t *testing.T) {
	repo := &mockHabilidadeRepo{notFound: true}
	r := setupHabilidadeRouter(repo)

	// URL corrigida para cair no handler
	req := httptest.NewRequest(http.MethodGet, "/habilidades/999", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetHabilidadeByID_InternalServerError(t *testing.T) {
	// Retorna um erro no repositório para forçar HTTP 500
	repo := &mockHabilidadeRepo{err: errors.New("erro de banco")}
	r := setupHabilidadeRouter(repo)

	req := httptest.NewRequest(http.MethodGet, "/habilidades/10", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUpdateHabilidade(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		body       string
		repoErr    error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "success",
			id:         "10",
			body:       `{"nome":"Go Avançado"}`,
			wantStatus: http.StatusOK,
			wantBody:   "Habilidade atualizada com sucesso",
		},
		{
			name:       "invalid id",
			id:         "invalid-id",
			body:       `{"nome":"Go Avançado"}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   "ID inválido",
		},
		{
			name:       "invalid json",
			id:         "10",
			body:       `{"nome":`,
			wantStatus: http.StatusBadRequest,
			wantBody:   "Dados inválidos:",
		},
		{
			name:       "required nome missing",
			id:         "10",
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   "Dados inválidos:",
		},
		{
			name:       "service error",
			id:         "10",
			body:       `{"nome":"Go Avançado"}`,
			repoErr:    errors.New("erro ao atualizar habilidade"),
			wantStatus: http.StatusInternalServerError,
			wantBody:   "Erro ao atualizar habilidade",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockHabilidadeRepo{err: tt.repoErr}
			r := setupHabilidadeRouter(repo)

			body := bytes.NewBufferString(tt.body)
			req := httptest.NewRequest(http.MethodPut, "/habilidades/"+tt.id, body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.wantBody)
		})
	}
}

func TestDeleteHabilidade(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		repoErr    error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "success",
			id:         "10",
			wantStatus: http.StatusOK,
			wantBody:   "Habilidade excluída com sucesso",
		},
		{
			name:       "invalid id",
			id:         "invalid-id",
			wantStatus: http.StatusBadRequest,
			wantBody:   "ID inválido",
		},
		{
			name:       "service error",
			id:         "10",
			repoErr:    errors.New("erro ao excluir habilidade"),
			wantStatus: http.StatusInternalServerError,
			wantBody:   "Erro ao excluir habilidade",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockHabilidadeRepo{err: tt.repoErr}
			r := setupHabilidadeRouter(repo)

			req := httptest.NewRequest(http.MethodDelete, "/habilidades/"+tt.id, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.wantBody)
		})
	}
}

func TestListHabilidades(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		repoErr    error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "success with explicit pagination and search",
			url:        "/habilidades?q=go&page=1&pageSize=10",
			wantStatus: http.StatusOK,
			wantBody:   "Go",
		},
		{
			name:       "normalizes page below minimum and pageSize above maximum",
			url:        "/habilidades?page=0&pageSize=101",
			wantStatus: http.StatusOK,
			wantBody:   "Go",
		},
		{
			name:       "normalizes invalid numeric pagination",
			url:        "/habilidades?page=invalid&pageSize=invalid",
			wantStatus: http.StatusOK,
			wantBody:   "Go",
		},
		{
			name:       "uses default pagination when query params are omitted",
			url:        "/habilidades",
			wantStatus: http.StatusOK,
			wantBody:   "Go",
		},
		{
			name:       "internal server error from service",
			url:        "/habilidades?q=go&page=1&pageSize=10",
			repoErr:    errors.New("erro de banco"),
			wantStatus: http.StatusInternalServerError,
			wantBody:   "Erro ao buscar habilidades",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockHabilidadeRepo{
				err:         tt.repoErr,
				habilidades: []*empmodels.Habilidade{{ID: 1, Nome: "Go"}},
				totalHab:    1,
			}
			r := setupHabilidadeRouter(repo)

			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.wantBody)
		})
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests: Áreas de Atuação
// ──────────────────────────────────────────────────────────────────────────────

func TestCreateAreaAtuacao_Success(t *testing.T) {
	repo := &mockHabilidadeRepo{lastCreatedID: 5}
	r := setupHabilidadeRouter(repo)

	body := bytes.NewBufferString(`{"nome":"Tecnologia"}`)
	req := httptest.NewRequest(http.MethodPost, "/habilidades/areas-atuacao", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestGetAreaAtuacaoByID_Success(t *testing.T) {
	repo := &mockHabilidadeRepo{areaAtuacao: &empmodels.AreaAtuacao{ID: 5, Nome: "Tecnologia"}}
	r := setupHabilidadeRouter(repo)

	req := httptest.NewRequest(http.MethodGet, "/habilidades/areas-atuacao/5", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdateAreaAtuacao_Success(t *testing.T) {
	repo := &mockHabilidadeRepo{}
	r := setupHabilidadeRouter(repo)

	body := bytes.NewBufferString(`{"nome":"TI & Comunicação"}`)
	req := httptest.NewRequest(http.MethodPut, "/habilidades/areas-atuacao/5", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteAreaAtuacao_Success(t *testing.T) {
	repo := &mockHabilidadeRepo{}
	r := setupHabilidadeRouter(repo)

	req := httptest.NewRequest(http.MethodDelete, "/habilidades/areas-atuacao/5", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestListAreasAtuacao_Success(t *testing.T) {
	repo := &mockHabilidadeRepo{
		areasAtuacao: []*empmodels.AreaAtuacao{{ID: 1, Nome: "TI"}},
		totalArea:    1,
	}
	r := setupHabilidadeRouter(repo)

	req := httptest.NewRequest(http.MethodGet, "/habilidades/areas-atuacao?q=ti", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests: Relacionamentos (Habilidade <-> Área de Atuação)
// ──────────────────────────────────────────────────────────────────────────────

func TestAttachAreaAtuacao_Success(t *testing.T) {
	repo := &mockHabilidadeRepo{}
	r := setupHabilidadeRouter(repo)

	req := httptest.NewRequest(http.MethodPost, "/habilidades/10/areas-atuacao/2", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAttachAreaAtuacao_InvalidHabilidadeID(t *testing.T) {
	repo := &mockHabilidadeRepo{}
	r := setupHabilidadeRouter(repo)

	req := httptest.NewRequest(http.MethodPost, "/habilidades/abc/areas-atuacao/2", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDetachAreaAtuacao_Success(t *testing.T) {
	repo := &mockHabilidadeRepo{}
	r := setupHabilidadeRouter(repo)

	req := httptest.NewRequest(http.MethodDelete, "/habilidades/10/areas-atuacao/2", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReplaceAreasAtuacao_Success(t *testing.T) {
	repo := &mockHabilidadeRepo{}
	r := setupHabilidadeRouter(repo)

	body := bytes.NewBufferString(`{"area_ids": [1, 2, 3]}`)
	req := httptest.NewRequest(http.MethodPut, "/habilidades/10/areas-atuacao", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReplaceAreasAtuacao_Error(t *testing.T) {
	repo := &mockHabilidadeRepo{err: fmt.Errorf("db error")}
	r := setupHabilidadeRouter(repo)

	body := bytes.NewBufferString(`{"area_ids": [1, 2]}`)
	req := httptest.NewRequest(http.MethodPut, "/habilidades/10/areas-atuacao", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
