package empregabilidade_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	handlers "github.com/prefeitura-rio/app-go-api/internal/handlers/v1/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	empmodels "github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// ──────────────────────────────────────────────────────────────────────────────
// Mock CurriculoRepository (implements CurriculoRepositoryInterface)
// ──────────────────────────────────────────────────────────────────────────────

type mockCurriculoRepoH struct {
	formacao               *empmodels.CurriculoFormacao
	idioma                 *empmodels.CurriculoIdioma
	habilidade             *empmodels.CurriculoHabilidade
	comportamento_atitudes empmodels.ComportamentoAtitudes
	curso                  *empmodels.CurriculoCursoComplementar
	experiencia            *empmodels.CurriculoExperiencia
	conquista              *empmodels.CurriculoConquista
	situacao               *empmodels.CurriculoSituacaoInteresses
	err                    error
	mock.Mock
}

// Formação

func (m *mockCurriculoRepoH) CreateFormacao(_ context.Context, _ *empmodels.CurriculoFormacao) (uuid.UUID, error) {
	if m.err != nil {
		return uuid.Nil, m.err
	}
	return uuid.New(), nil
}

func (m *mockCurriculoRepoH) GetFormacaoByID(_ context.Context, _ uuid.UUID) (*empmodels.CurriculoFormacao, error) {
	return m.formacao, m.err
}

func (m *mockCurriculoRepoH) UpdateFormacao(_ context.Context, _ *empmodels.CurriculoFormacao) error {
	return m.err
}

func (m *mockCurriculoRepoH) DeleteFormacao(_ context.Context, _ uuid.UUID) error {
	return m.err
}

func (m *mockCurriculoRepoH) ListFormacoesByCPF(_ context.Context, _ string) ([]*empmodels.CurriculoFormacao, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.formacao != nil {
		return []*empmodels.CurriculoFormacao{m.formacao}, nil
	}
	return []*empmodels.CurriculoFormacao{}, nil
}

// Idioma

func (m *mockCurriculoRepoH) CreateIdioma(_ context.Context, _ *empmodels.CurriculoIdioma) (uuid.UUID, error) {
	if m.err != nil {
		return uuid.Nil, m.err
	}
	return uuid.New(), nil
}

func (m *mockCurriculoRepoH) GetIdiomaByID(_ context.Context, _ uuid.UUID) (*empmodels.CurriculoIdioma, error) {
	return m.idioma, m.err
}

func (m *mockCurriculoRepoH) UpdateIdioma(_ context.Context, _ *empmodels.CurriculoIdioma) error {
	return m.err
}

func (m *mockCurriculoRepoH) DeleteIdioma(_ context.Context, _ uuid.UUID) error {
	return m.err
}

func (m *mockCurriculoRepoH) ListIdiomasByCPF(_ context.Context, _ string) ([]*empmodels.CurriculoIdioma, error) {
	if m.err != nil {
		return nil, m.err
	}
	return []*empmodels.CurriculoIdioma{}, nil
}

// Curso Complementar

func (m *mockCurriculoRepoH) CreateCursoComplementar(_ context.Context, _ *empmodels.CurriculoCursoComplementar) (uuid.UUID, error) {
	if m.err != nil {
		return uuid.Nil, m.err
	}
	return uuid.New(), nil
}

func (m *mockCurriculoRepoH) GetCursoComplementarByID(_ context.Context, _ uuid.UUID) (*empmodels.CurriculoCursoComplementar, error) {
	return m.curso, m.err
}

func (m *mockCurriculoRepoH) UpdateCursoComplementar(_ context.Context, _ *empmodels.CurriculoCursoComplementar) error {
	return m.err
}

func (m *mockCurriculoRepoH) DeleteCursoComplementar(_ context.Context, _ uuid.UUID) error {
	return m.err
}

func (m *mockCurriculoRepoH) ListCursosComplementaresByCPF(_ context.Context, _ string) ([]*empmodels.CurriculoCursoComplementar, error) {
	if m.err != nil {
		return nil, m.err
	}
	return []*empmodels.CurriculoCursoComplementar{}, nil
}

// Experiência

func (m *mockCurriculoRepoH) CreateExperiencia(_ context.Context, _ *empmodels.CurriculoExperiencia) (uuid.UUID, error) {
	if m.err != nil {
		return uuid.Nil, m.err
	}
	return uuid.New(), nil
}

func (m *mockCurriculoRepoH) GetExperienciaByID(_ context.Context, _ uuid.UUID) (*empmodels.CurriculoExperiencia, error) {
	return m.experiencia, m.err
}

func (m *mockCurriculoRepoH) UpdateExperiencia(_ context.Context, _ *empmodels.CurriculoExperiencia) error {
	return m.err
}

func (m *mockCurriculoRepoH) DeleteExperiencia(_ context.Context, _ uuid.UUID) error {
	return m.err
}

func (m *mockCurriculoRepoH) ListExperienciasByCPF(_ context.Context, _ string) ([]*empmodels.CurriculoExperiencia, error) {
	if m.err != nil {
		return nil, m.err
	}
	return []*empmodels.CurriculoExperiencia{}, nil
}

// Conquista

func (m *mockCurriculoRepoH) CreateConquista(_ context.Context, _ *empmodels.CurriculoConquista) (uuid.UUID, error) {
	if m.err != nil {
		return uuid.Nil, m.err
	}
	return uuid.New(), nil
}

func (m *mockCurriculoRepoH) GetConquistaByID(_ context.Context, _ uuid.UUID) (*empmodels.CurriculoConquista, error) {
	return m.conquista, m.err
}

func (m *mockCurriculoRepoH) UpdateConquista(_ context.Context, _ *empmodels.CurriculoConquista) error {
	return m.err
}

func (m *mockCurriculoRepoH) DeleteConquista(_ context.Context, _ uuid.UUID) error {
	return m.err
}

func (m *mockCurriculoRepoH) ListConquistasByCPF(_ context.Context, _ string) ([]*empmodels.CurriculoConquista, error) {
	if m.err != nil {
		return nil, m.err
	}
	return []*empmodels.CurriculoConquista{}, nil
}

// ReplaceAll

func (m *mockCurriculoRepoH) ReplaceAllFormacoesByCPF(_ context.Context, _ string, _ []*empmodels.CurriculoFormacao) error {
	return m.err
}

func (m *mockCurriculoRepoH) ReplaceAllFormacaoAccordionByCPF(_ context.Context, _ string, _ []*empmodels.CurriculoFormacao, _ []*empmodels.CurriculoIdioma) error {
	return m.err
}

func (m *mockCurriculoRepoH) ReplaceAllExperienciasByCPF(_ context.Context, _ string, _ []*empmodels.CurriculoExperiencia) error {
	return m.err
}

func (m *mockCurriculoRepoH) ReplaceAllExperienciaProfissionalAccordionByCPF(_ context.Context, _ string, _ []*empmodels.CurriculoExperiencia, _ []*empmodels.CurriculoConquista, _ string) error {
	return m.err
}

func (m *mockCurriculoRepoH) ReplaceAllConquistasByCPF(_ context.Context, _ string, _ []*empmodels.CurriculoConquista) error {
	return m.err
}

func (m *mockCurriculoRepoH) ReplaceAllIdiomasByCPF(_ context.Context, _ string, _ []*empmodels.CurriculoIdioma) error {
	return m.err
}

func (m *mockCurriculoRepoH) ReplaceAllHabilidadesByCPF(_ context.Context, _ string, _ []*empmodels.CurriculoHabilidade) error {
	return m.err
}

func (m *mockCurriculoRepoH) ReplaceAllCursosComplementaresByCPF(_ context.Context, _ string, _ []*empmodels.CurriculoCursoComplementar) error {
	return m.err
}

func (m *mockCurriculoRepoH) ReplaceAllItensCurriculoByCPF(_ context.Context, _ string, _ *empmodels.CurriculoItensReplaceAll) error {
	return m.err
}

// Situação e Interesses

func (m *mockCurriculoRepoH) UpsertSituacaoInteresses(_ context.Context, _ *empmodels.CurriculoSituacaoInteresses) error {
	return m.err
}

func (m *mockCurriculoRepoH) GetSituacaoInteressesByCPF(_ context.Context, _ string) (*empmodels.CurriculoSituacaoInteresses, error) {
	return m.situacao, m.err
}

// Perfil

func (m *mockCurriculoRepoH) GetPerfilByCPF(_ context.Context, _ string) (*empmodels.CurriculoPerfil, error) {
	return nil, m.err
}

// Habilidade

func (m *mockCurriculoRepoH) ListHabilidadesByCPF(ctx context.Context, cpf string) ([]*empmodels.CurriculoHabilidade, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.habilidade != nil {
		return []*empmodels.CurriculoHabilidade{m.habilidade}, nil
	}

	return []*empmodels.CurriculoHabilidade{}, nil
}

func (m *mockCurriculoRepoH) AddHabilidadeAoCurriculo(_ context.Context, _ *empmodels.CurriculoHabilidade) error {
	if m.err != nil {
		return m.err
	}
	return nil
}

func (m *mockCurriculoRepoH) DetachHabilidadeDoCurriculo(ctx context.Context, id int64) error {
	return m.err
}

// Comportamentos e Atitudes

func (m *mockCurriculoRepoH) ListComportamentoAtitudesByCPF(_ context.Context, _ string) ([]*empmodels.CurriculoComportamentoAtitudes, error) {
	if m.err != nil {
		return nil, m.err
	}
	return []*empmodels.CurriculoComportamentoAtitudes{}, nil
}

func (m *mockCurriculoRepoH) AddComportamentoAtitudesAoCurriculo(ctx context.Context, vinculo *empmodels.CurriculoComportamentoAtitudes) error {
	if m.err != nil {
		return m.err
	}
	return nil
}

func (m *mockCurriculoRepoH) DetachComportamentoAtitudesDoCurriculo(_ context.Context, _ int64) error {
	return m.err
}

// ──────────────────────────────────────────────────────────────────────────────
// Router setup
// ──────────────────────────────────────────────────────────────────────────────

func setupCurriculoRouter(repo services.CurriculoRepositoryInterface, cpf string, isAdmin bool) *gin.Engine {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if cpf == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		c.Set(middlewares.UserCPFKey, cpf)
		if isAdmin {
			c.Set(middlewares.UserRoleKey, "ADMIN")
		}
		c.Next()
	})

	svc := services.NewCurriculoServiceWithInterface(repo)
	h := handlers.NewCurriculoHandler(svc)

	r.GET("/curriculo/:cpf", h.GetCurriculoCompleto)

	r.POST("/curriculo/formacoes", h.CreateFormacao)
	r.GET("/curriculo/formacoes/:id", h.GetFormacaoByID)
	r.PUT("/curriculo/formacoes/:id", h.UpdateFormacao)
	r.DELETE("/curriculo/formacoes/:id", h.DeleteFormacao)
	r.GET("/curriculo/:cpf/formacoes", h.ListFormacoesByCPF)

	r.POST("/curriculo/idiomas", h.CreateIdioma)
	r.GET("/curriculo/idiomas/:id", h.GetIdiomaByID)
	r.PUT("/curriculo/idiomas/:id", h.UpdateIdioma)
	r.DELETE("/curriculo/idiomas/:id", h.DeleteIdioma)
	r.GET("/curriculo/:cpf/idiomas", h.ListIdiomasByCPF)

	r.POST("/curriculo/cursos", h.CreateCursoComplementar)
	r.GET("/curriculo/cursos/:id", h.GetCursoComplementarByID)
	r.PUT("/curriculo/cursos/:id", h.UpdateCursoComplementar)
	r.PUT("/cursos/:id", h.UpdateCursoComplementar) // Alternative path for tests
	r.DELETE("/curriculo/cursos/:id", h.DeleteCursoComplementar)
	r.GET("/curriculo/:cpf/cursos", h.ListCursosComplementaresByCPF)

	r.POST("/curriculo/experiencias", h.CreateExperiencia)
	r.GET("/curriculo/experiencias/:id", h.GetExperienciaByID)
	r.PUT("/curriculo/experiencias/:id", h.UpdateExperiencia)
	r.PUT("/experiencias/:id", h.UpdateExperiencia) // Alternative path for tests
	r.DELETE("/curriculo/experiencias/:id", h.DeleteExperiencia)
	r.GET("/curriculo/:cpf/experiencias", h.ListExperienciasByCPF)

	r.POST("/curriculo/conquistas", h.CreateConquista)
	r.GET("/curriculo/conquistas/:id", h.GetConquistaByID)
	r.PUT("/curriculo/conquistas/:id", h.UpdateConquista)
	r.PUT("/conquistas/:id", h.UpdateConquista) // Alternative path for tests
	r.DELETE("/curriculo/conquistas/:id", h.DeleteConquista)
	r.GET("/curriculo/:cpf/conquistas", h.ListConquistasByCPF)

	// Habilidades
	r.GET("/curriculo/habilidades", h.ListHabilidadesDoCurriculo)
	r.DELETE("/curriculo/habilidades/:id", h.DeleteHabilidadeDoCurriculo)
	r.POST("/curriculo/habilidades", h.AddHabilidadeAoCurriculo)

	// Comportamento/atitudes
	r.GET("/curriculo/comportamentos-atitudes", h.ListComportamentoAtitudesDoCurriculo)
	r.DELETE("/curriculo/comportamentos-atitudes/:id", h.DeleteComportamentoAtitudesDoCurriculo)
	r.POST("/curriculo/comportamentos-atitudes", h.AddComportamentoAtitudesAoCurriculo)

	// ReplaceAll endpoints
	r.PUT("/curriculo/:cpf/formacoes", h.ReplaceAllFormacoesByCPF)
	r.PUT("/curriculo/:cpf/experiencias", h.ReplaceAllExperienciasByCPF)
	r.PUT("/curriculo/:cpf/conquistas", h.ReplaceAllConquistasByCPF)
	r.PUT("/curriculo/:cpf/idiomas", h.ReplaceAllIdiomasByCPF)
	r.PUT("/curriculo/:cpf/cursos-complementares", h.ReplaceAllCursosComplementaresByCPF)
	r.PUT("/curriculo", h.ReplaceAllItensCurriculoByCPF)

	r.PUT("/curriculo/:cpf/situacao-interesses", h.UpsertSituacaoInteresses)
	r.GET("/curriculo/:cpf/situacao-interesses", h.GetSituacaoInteressesByCPF)

	return r
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests: GetCurriculoCompleto
// ──────────────────────────────────────────────────────────────────────────────

func TestCurriculoHandler_GetCurriculoCompleto_AsOwner(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/12345678900", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCurriculoHandler_GetCurriculoCompleto_Forbidden(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "99999999999", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/12345678900", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestCurriculoHandler_GetCurriculoCompleto_Unauthorized(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/12345678900", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusForbidden {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests: Formação
// ──────────────────────────────────────────────────────────────────────────────

func TestCurriculoHandler_CreateFormacao_Success(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"instituicao":"UFRJ","curso":"Engenharia","status_formacao":"em_andamento"}`)
	req := httptest.NewRequest(http.MethodPost, "/curriculo/formacoes", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCurriculoHandler_CreateFormacao_Unauthorized(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "", false)
	body := bodyOf(`{"instituicao":"UFRJ"}`)
	req := httptest.NewRequest(http.MethodPost, "/curriculo/formacoes", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestCurriculoHandler_GetFormacaoByID_Found(t *testing.T) {
	id := uuid.MustParse(validUUID)
	repo := &mockCurriculoRepoH{formacao: &empmodels.CurriculoFormacao{ID: id, CPF: "12345678900"}}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/formacoes/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestCurriculoHandler_GetFormacaoByID_NotFound(t *testing.T) {
	repo := &mockCurriculoRepoH{formacao: nil}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/formacoes/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestCurriculoHandler_GetFormacaoByID_InvalidID(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/formacoes/bad-id", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCurriculoHandler_UpdateFormacao_AsOwner(t *testing.T) {
	id := uuid.MustParse(validUUID)
	repo := &mockCurriculoRepoH{formacao: &empmodels.CurriculoFormacao{ID: id, CPF: "12345678900"}}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"instituicao":"PUC","status_formacao":"em_andamento"}`)
	req := httptest.NewRequest(http.MethodPut, "/curriculo/formacoes/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCurriculoHandler_UpdateFormacao_NotFound(t *testing.T) {
	repo := &mockCurriculoRepoH{formacao: nil}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"instituicao":"PUC","status_formacao":"em_andamento"}`)
	req := httptest.NewRequest(http.MethodPut, "/curriculo/formacoes/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestCurriculoHandler_UpdateFormacao_InvalidID(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"instituicao":"PUC"}`)
	req := httptest.NewRequest(http.MethodPut, "/curriculo/formacoes/bad-id", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCurriculoHandler_DeleteFormacao_AsOwner(t *testing.T) {
	id := uuid.MustParse(validUUID)
	repo := &mockCurriculoRepoH{formacao: &empmodels.CurriculoFormacao{ID: id, CPF: "12345678900"}}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodDelete, "/curriculo/formacoes/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestCurriculoHandler_DeleteFormacao_NotFound(t *testing.T) {
	repo := &mockCurriculoRepoH{formacao: nil}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodDelete, "/curriculo/formacoes/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestCurriculoHandler_ListFormacoesByCPF_AsOwner(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/12345678900/formacoes", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests: Idioma
// ──────────────────────────────────────────────────────────────────────────────

func TestCurriculoHandler_CreateIdioma_Success(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"id_idioma":"` + validUUID + `","id_nivel_idioma":"` + validUUID + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/curriculo/idiomas", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCurriculoHandler_CreateIdioma_Unauthorized(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "", false)
	body := bodyOf(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/curriculo/idiomas", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestCurriculoHandler_GetIdiomaByID_Found(t *testing.T) {
	id := uuid.MustParse(validUUID)
	repo := &mockCurriculoRepoH{idioma: &empmodels.CurriculoIdioma{ID: id, CPF: "12345678900"}}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/idiomas/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestCurriculoHandler_GetIdiomaByID_NotFound(t *testing.T) {
	repo := &mockCurriculoRepoH{idioma: nil}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/idiomas/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestCurriculoHandler_GetIdiomaByID_InvalidID(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/idiomas/bad-id", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCurriculoHandler_UpdateIdioma_AsOwner(t *testing.T) {
	id := uuid.MustParse(validUUID)
	repo := &mockCurriculoRepoH{idioma: &empmodels.CurriculoIdioma{ID: id, CPF: "12345678900"}}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"id_idioma":"` + validUUID + `","id_nivel_idioma":"` + validUUID + `"}`)
	req := httptest.NewRequest(http.MethodPut, "/curriculo/idiomas/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCurriculoHandler_DeleteIdioma_AsOwner(t *testing.T) {
	id := uuid.MustParse(validUUID)
	repo := &mockCurriculoRepoH{idioma: &empmodels.CurriculoIdioma{ID: id, CPF: "12345678900"}}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodDelete, "/curriculo/idiomas/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestCurriculoHandler_ListIdiomasByCPF_AsOwner(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/12345678900/idiomas", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests: Curso Complementar
// ──────────────────────────────────────────────────────────────────────────────

func TestCurriculoHandler_CreateCurso_Success(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"nome":"Curso Go","carga_horaria":40}`)
	req := httptest.NewRequest(http.MethodPost, "/curriculo/cursos", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCurriculoHandler_GetCursoByID_Found(t *testing.T) {
	id := uuid.MustParse(validUUID)
	repo := &mockCurriculoRepoH{curso: &empmodels.CurriculoCursoComplementar{ID: id, CPF: "12345678900"}}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/cursos/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestCurriculoHandler_GetCursoByID_NotFound(t *testing.T) {
	repo := &mockCurriculoRepoH{curso: nil}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/cursos/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestCurriculoHandler_DeleteCurso_AsOwner(t *testing.T) {
	id := uuid.MustParse(validUUID)
	repo := &mockCurriculoRepoH{curso: &empmodels.CurriculoCursoComplementar{ID: id, CPF: "12345678900"}}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodDelete, "/curriculo/cursos/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestCurriculoHandler_ListCursosByCPF_AsOwner(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/12345678900/cursos", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests: Experiência
// ──────────────────────────────────────────────────────────────────────────────

func TestCurriculoHandler_CreateExperiencia_Success(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"cargo":"Dev","empresa":"Tech","emprego_atual":true}`)
	req := httptest.NewRequest(http.MethodPost, "/curriculo/experiencias", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCurriculoHandler_GetExperienciaByID_Found(t *testing.T) {
	id := uuid.MustParse(validUUID)
	repo := &mockCurriculoRepoH{experiencia: &empmodels.CurriculoExperiencia{ID: id, CPF: "12345678900"}}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/experiencias/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestCurriculoHandler_GetExperienciaByID_NotFound(t *testing.T) {
	repo := &mockCurriculoRepoH{experiencia: nil}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/experiencias/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestCurriculoHandler_DeleteExperiencia_AsOwner(t *testing.T) {
	id := uuid.MustParse(validUUID)
	repo := &mockCurriculoRepoH{experiencia: &empmodels.CurriculoExperiencia{ID: id, CPF: "12345678900"}}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodDelete, "/curriculo/experiencias/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestCurriculoHandler_ListExperienciasByCPF_AsOwner(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/12345678900/experiencias", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests: Conquista
// ──────────────────────────────────────────────────────────────────────────────

func TestCurriculoHandler_CreateConquista_Success(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"titulo":"Premio Inovacao","id_tipo_conquista":"` + validUUID + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/curriculo/conquistas", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCurriculoHandler_GetConquistaByID_Found(t *testing.T) {
	id := uuid.MustParse(validUUID)
	repo := &mockCurriculoRepoH{conquista: &empmodels.CurriculoConquista{ID: id, CPF: "12345678900"}}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/conquistas/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestCurriculoHandler_GetConquistaByID_NotFound(t *testing.T) {
	repo := &mockCurriculoRepoH{conquista: nil}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/conquistas/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestCurriculoHandler_DeleteConquista_AsOwner(t *testing.T) {
	id := uuid.MustParse(validUUID)
	repo := &mockCurriculoRepoH{conquista: &empmodels.CurriculoConquista{ID: id, CPF: "12345678900"}}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodDelete, "/curriculo/conquistas/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestCurriculoHandler_ListConquistasByCPF_AsOwner(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/12345678900/conquistas", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests: SituacaoInteresses
// ──────────────────────────────────────────────────────────────────────────────

func TestCurriculoHandler_UpsertSituacaoInteresses_AsOwner(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{}`)
	req := httptest.NewRequest(http.MethodPut, "/curriculo/12345678900/situacao-interesses", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCurriculoHandler_GetSituacaoInteressesByCPF_AsOwner(t *testing.T) {
	repo := &mockCurriculoRepoH{situacao: &empmodels.CurriculoSituacaoInteresses{CPF: "12345678900"}}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/12345678900/situacao-interesses", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// ──────────────────────────────────────────────────────────────────────────────ƒ
// Tests: UpdateCursoComplementar (0% coverage)
// ──────────────────────────────────────────────────────────────────────────────

func TestCurriculoHandler_UpdateCursoComplementar_InvalidID(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"nome":"Curso Atualizado"}`)
	req := httptest.NewRequest(http.MethodPut, "/cursos/bad-id", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCurriculoHandler_UpdateCursoComplementar_NotFound(t *testing.T) {
	repo := &mockCurriculoRepoH{curso: nil}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"nome":"Curso Atualizado"}`)
	req := httptest.NewRequest(http.MethodPut, "/cursos/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestCurriculoHandler_UpdateCursoComplementar_Forbidden(t *testing.T) {
	repo := &mockCurriculoRepoH{
		curso: &empmodels.CurriculoCursoComplementar{
			ID:  uuid.MustParse(validUUID),
			CPF: "12345678900",
		},
	}
	r := setupCurriculoRouter(repo, "99999999999", false)
	body := bodyOf(`{"nome":"Curso Atualizado"}`)
	req := httptest.NewRequest(http.MethodPut, "/cursos/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestCurriculoHandler_UpdateCursoComplementar_Success(t *testing.T) {
	repo := &mockCurriculoRepoH{
		curso: &empmodels.CurriculoCursoComplementar{
			ID:  uuid.MustParse(validUUID),
			CPF: "12345678900",
		},
	}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"nome":"Curso Atualizado"}`)
	req := httptest.NewRequest(http.MethodPut, "/cursos/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCurriculoHandler_UpdateCursoComplementar_BadJSON(t *testing.T) {
	repo := &mockCurriculoRepoH{
		curso: &empmodels.CurriculoCursoComplementar{
			ID:  uuid.MustParse(validUUID),
			CPF: "12345678900",
		},
	}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{bad}`)
	req := httptest.NewRequest(http.MethodPut, "/cursos/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// --- HABILIDADES ---

// 1. Sucesso na listagem de habilidades (HTTP 200)
func TestCurriculoHandler_ListHabilidadesDoCurriculo_Success(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)

	req := httptest.NewRequest(http.MethodGet, "/curriculo/habilidades", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// 2. Não Autenticado / CPF ausente no contexto (HTTP 401)
func TestCurriculoHandler_ListHabilidadesDoCurriculo_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &mockCurriculoRepoH{}
	svc := services.NewCurriculoServiceWithInterface(repo)
	h := handlers.NewCurriculoHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(
		http.MethodGet,
		"/curriculo/habilidades",
		nil,
	)

	h.ListHabilidadesDoCurriculo(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// 3. Erro na camada de serviço/banco de dados (HTTP 500)
func TestCurriculoHandler_ListHabilidadesDoCurriculo_ServiceError(t *testing.T) {
	repo := &mockCurriculoRepoH{err: errors.New("erro ao consultar habilidades")}
	r := setupCurriculoRouter(repo, "12345678900", false)

	req := httptest.NewRequest(http.MethodGet, "/curriculo/habilidades", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// 1. Sucesso (HTTP 201)
func TestCurriculoHandler_AddHabilidadeAoCurriculo_Success(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)

	body := bodyOf(`{"id_habilidade": 1}`)
	req := httptest.NewRequest(http.MethodPost, "/curriculo/habilidades", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

// 2. Não Autenticado / CPF Vazio no Middleware (HTTP 401)
func TestCurriculoHandler_AddHabilidadeAoCurriculo_Unauthorized(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	// Certifique-se de que o router mock registra a rota exata POST /curriculo/habilidades
	r := setupCurriculoRouter(repo, "", false)

	body := bodyOf(`{"id_habilidade": 1}`)
	req := httptest.NewRequest(http.MethodPost, "/curriculo/habilidades", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code, "Deveria retornar 401 quando GetUserCPF retorna string vazia")
}

// 3. JSON Inválido / ShouldBindJSON Error (HTTP 400)
func TestCurriculoHandler_AddHabilidadeAoCurriculo_InvalidJSON(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)

	body := bodyOf(`{"id_habilidade": "texto_em_vez_de_int"}`)
	req := httptest.NewRequest(http.MethodPost, "/curriculo/habilidades", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// 4. Erro de Serviço/Banco (HTTP 500)
func TestCurriculoHandler_AddHabilidadeAoCurriculo_ServiceError(t *testing.T) {
	repo := &mockCurriculoRepoH{err: errors.New("erro interno")}
	r := setupCurriculoRouter(repo, "12345678900", false)

	body := bodyOf(`{"id_habilidade": 1}`)
	req := httptest.NewRequest(http.MethodPost, "/curriculo/habilidades", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCurriculoHandler_AddHabilidadeAoCurriculo_SemCPFNoContexto(t *testing.T) {
	repo := &mockCurriculoRepoH{}

	// Passa string vazia "" para simular ausência do CPF no contexto/claims do middleware
	r := setupCurriculoRouter(repo, "", false)

	body := bodyOf(`{"id_habilidade": 1}`)
	req := httptest.NewRequest(http.MethodPost, "/curriculo/habilidades", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCurriculoHandler_AddHabilidadeAoCurriculo_ValidationError(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{}`) // Sem o id_habilidade obrigatório
	req := httptest.NewRequest(http.MethodPost, "/curriculo/habilidades", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest && w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 400 or 422, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCurriculoHandler_AddHabilidadeAoCurriculo_RepoError(t *testing.T) {
	repo := &mockCurriculoRepoH{err: errors.New("erro de banco")}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"id_habilidade": 1}`)
	req := httptest.NewRequest(http.MethodPost, "/curriculo/habilidades", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d: %s", http.StatusInternalServerError, w.Code, w.Body.String())
	}
}

func TestCurriculoHandler_AddHabilidadeAoCurriculo_Conflict(t *testing.T) {
	repo := &mockCurriculoRepoH{err: errors.New("habilidade ja vinculada")}
	r := setupCurriculoRouter(repo, "12345678900", false)

	body := bodyOf(`{"id_habilidade": 1}`)
	req := httptest.NewRequest(http.MethodPost, "/curriculo/habilidades", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// O handler retorna 500 para qualquer erro do repositório/serviço
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d: %s", http.StatusInternalServerError, w.Code, w.Body.String())
	}
}

func TestCurriculoHandler_AddHabilidadeAoCurriculo_IdZeroOuNegativo(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)

	body := bodyOf(`{"id_habilidade": 0}`)
	req := httptest.NewRequest(http.MethodPost, "/curriculo/habilidades", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest && w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 400 or 422, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCurriculoHandler_DeleteHabilidadeDoCurriculo_NotFound(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)

	// Sem o ID no final, a rota não existe no Gin
	req := httptest.NewRequest(http.MethodDelete, "/curriculo/habilidades", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCurriculoHandler_DeleteHabilidadeDoCurriculo_Success(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodDelete, "/curriculo/habilidades/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent && w.Code != http.StatusOK {
		t.Errorf("expected 204 or 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCurriculoHandler_DeleteHabilidadeDoCurriculo_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &mockCurriculoRepoH{}
	svc := services.NewCurriculoServiceWithInterface(repo)
	h := handlers.NewCurriculoHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(
		http.MethodDelete,
		"/curriculo/habilidades/1",
		nil,
	)
	c.Params = gin.Params{
		{Key: "id", Value: "1"},
	}

	h.DeleteHabilidadeDoCurriculo(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCurriculoHandler_DeleteHabilidadeDoCurriculo_InvalidID(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/curriculo/habilidades/abc",
		nil,
	)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCurriculoHandler_DeleteHabilidadeDoCurriculo_RecordNotFound(t *testing.T) {
	repo := &mockCurriculoRepoH{
		err: gorm.ErrRecordNotFound,
	}
	r := setupCurriculoRouter(repo, "12345678900", false)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/curriculo/habilidades/1",
		nil,
	)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCurriculoHandler_DeleteHabilidadeDoCurriculo_ServiceError(t *testing.T) {
	repo := &mockCurriculoRepoH{
		err: errors.New("erro ao remover habilidade"),
	}
	r := setupCurriculoRouter(repo, "12345678900", false)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/curriculo/habilidades/1",
		nil,
	)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- COMPORTAMENTOS E ATITUDES ---

func TestCurriculoHandler_ListComportamentoAtitudesDoCurriculo_Success(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/comportamentos-atitudes", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `[]`, w.Body.String())
}

func TestCurriculoHandler_ListComportamentoAtitudesDoCurriculo_Unauthorized(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	svc := services.NewCurriculoServiceWithInterface(repo)
	h := handlers.NewCurriculoHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/curriculo/comportamentos-atitudes", nil)

	h.ListComportamentoAtitudesDoCurriculo(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Usuário não autenticado")
}

func TestCurriculoHandler_ListComportamentoAtitudesDoCurriculo_ServiceError(t *testing.T) {
	repo := &mockCurriculoRepoH{err: errors.New("erro ao buscar comportamentos")}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/comportamentos-atitudes", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Erro ao buscar comportamentos/atitudes do currículo")
}

func TestCurriculoHandler_AddComportamentoAtitudesAoCurriculo_Success(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"id_comportamento_atitudes": 1}`)
	req := httptest.NewRequest(http.MethodPost, "/curriculo/comportamentos-atitudes", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "12345678900", response["cpf"])
	assert.Equal(t, float64(1), response["id_comportamento_atitudes"])
}

func TestCurriculoHandler_AddComportamentoAtitudesAoCurriculo_ErrorPaths(t *testing.T) {
	tests := []struct {
		name           string
		cpf            string
		body           string
		repoErr        error
		wantStatus     int
		wantBodySubstr string
	}{
		{
			name:           "sem CPF no contexto",
			cpf:            "",
			body:           `{"id_comportamento_atitudes":1}`,
			wantStatus:     http.StatusUnauthorized,
			wantBodySubstr: "Usuário não autenticado",
		},
		{
			name:           "JSON inválido",
			cpf:            "12345678900",
			body:           `{"id_comportamento_atitudes":"invalido"}`,
			wantStatus:     http.StatusBadRequest,
			wantBodySubstr: "Dados inválidos:",
		},
		{
			name:           "erro do serviço",
			cpf:            "12345678900",
			body:           `{"id_comportamento_atitudes":1}`,
			repoErr:        errors.New("erro ao persistir vínculo"),
			wantStatus:     http.StatusInternalServerError,
			wantBodySubstr: "Erro ao adicionar comportamento/atitude ao currículo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCurriculoRepoH{err: tt.repoErr}
			svc := services.NewCurriculoServiceWithInterface(repo)
			h := handlers.NewCurriculoHandler(svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/curriculo/comportamentos-atitudes", bodyOf(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")
			if tt.cpf != "" {
				c.Set(middlewares.UserCPFKey, tt.cpf)
			}

			h.AddComportamentoAtitudesAoCurriculo(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.wantBodySubstr)
		})
	}
}

func TestCurriculoHandler_DeleteComportamentoAtitudesDoCurriculo(t *testing.T) {
	tests := []struct {
		name       string
		cpf        string
		id         string
		repoErr    error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "ID inválido",
			cpf:        "12345678900",
			id:         "invalido",
			wantStatus: http.StatusBadRequest,
			wantBody:   "ID inválido",
		},
		{
			name:       "vínculo não encontrado",
			cpf:        "12345678900",
			id:         "99",
			repoErr:    gorm.ErrRecordNotFound,
			wantStatus: http.StatusNotFound,
			wantBody:   "Vínculo não encontrado ou não pertence ao usuário",
		},
		{
			name:       "erro interno",
			cpf:        "12345678900",
			id:         "10",
			repoErr:    errors.New("erro de banco"),
			wantStatus: http.StatusInternalServerError,
			wantBody:   "Erro ao remover comportamento/atitude do currículo",
		},
		{
			name:       "sucesso",
			cpf:        "12345678900",
			id:         "1",
			wantStatus: http.StatusOK,
			wantBody:   "Comportamento/atitude desvinculado com sucesso",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCurriculoRepoH{err: tt.repoErr}
			r := setupCurriculoRouter(repo, tt.cpf, false)
			req := httptest.NewRequest(http.MethodDelete, "/curriculo/comportamentos-atitudes/"+tt.id, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.wantBody)
		})
	}
}

func TestCurriculoHandler_DeleteComportamentoAtitudesDoCurriculo_Unauthorized(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	svc := services.NewCurriculoServiceWithInterface(repo)
	h := handlers.NewCurriculoHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/curriculo/comportamentos-atitudes/1", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	h.DeleteComportamentoAtitudesDoCurriculo(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Usuário não autenticado")
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests: ReplaceAllItensCurriculoByCPF
// ──────────────────────────────────────────────────────────────────────────────

func TestCurriculoHandler_ReplaceAllItensCurriculoByCPF_Success(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"habilidades":[{"descricao":"Go"}],"comportamentos":[{"descricao":"Liderança"}]}`)
	req := httptest.NewRequest(http.MethodPut, "/curriculo", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests: UpdateExperiencia (0% coverage)
// ──────────────────────────────────────────────────────────────────────────────

func TestCurriculoHandler_UpdateExperiencia_InvalidID(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"cargo":"Desenvolvedor Senior"}`)
	req := httptest.NewRequest(http.MethodPut, "/experiencias/bad-id", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCurriculoHandler_UpdateExperiencia_NotFound(t *testing.T) {
	repo := &mockCurriculoRepoH{experiencia: nil}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"cargo":"Desenvolvedor Senior"}`)
	req := httptest.NewRequest(http.MethodPut, "/experiencias/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestCurriculoHandler_UpdateExperiencia_Forbidden(t *testing.T) {
	repo := &mockCurriculoRepoH{
		experiencia: &empmodels.CurriculoExperiencia{
			ID:  uuid.MustParse(validUUID),
			CPF: "12345678900",
		},
	}
	r := setupCurriculoRouter(repo, "99999999999", false)
	body := bodyOf(`{"cargo":"Desenvolvedor Senior"}`)
	req := httptest.NewRequest(http.MethodPut, "/experiencias/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestCurriculoHandler_UpdateExperiencia_Success(t *testing.T) {
	repo := &mockCurriculoRepoH{
		experiencia: &empmodels.CurriculoExperiencia{
			ID:  uuid.MustParse(validUUID),
			CPF: "12345678900",
		},
	}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"cargo":"Desenvolvedor Senior"}`)
	req := httptest.NewRequest(http.MethodPut, "/experiencias/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCurriculoHandler_UpdateExperiencia_BadJSON(t *testing.T) {
	repo := &mockCurriculoRepoH{
		experiencia: &empmodels.CurriculoExperiencia{
			ID:  uuid.MustParse(validUUID),
			CPF: "12345678900",
		},
	}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{bad}`)
	req := httptest.NewRequest(http.MethodPut, "/experiencias/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests: UpdateConquista (0% coverage)
// ──────────────────────────────────────────────────────────────────────────────

func TestCurriculoHandler_UpdateConquista_InvalidID(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"titulo":"Prêmio Atualizado"}`)
	req := httptest.NewRequest(http.MethodPut, "/conquistas/bad-id", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCurriculoHandler_UpdateConquista_NotFound(t *testing.T) {
	repo := &mockCurriculoRepoH{conquista: nil}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"titulo":"Prêmio Atualizado"}`)
	req := httptest.NewRequest(http.MethodPut, "/conquistas/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestCurriculoHandler_UpdateConquista_Forbidden(t *testing.T) {
	repo := &mockCurriculoRepoH{
		conquista: &empmodels.CurriculoConquista{
			ID:  uuid.MustParse(validUUID),
			CPF: "12345678900",
		},
	}
	r := setupCurriculoRouter(repo, "99999999999", false)
	body := bodyOf(`{"titulo":"Prêmio Atualizado"}`)
	req := httptest.NewRequest(http.MethodPut, "/conquistas/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestCurriculoHandler_UpdateConquista_Success(t *testing.T) {
	repo := &mockCurriculoRepoH{
		conquista: &empmodels.CurriculoConquista{
			ID:  uuid.MustParse(validUUID),
			CPF: "12345678900",
		},
	}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"titulo":"Prêmio Atualizado"}`)
	req := httptest.NewRequest(http.MethodPut, "/conquistas/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCurriculoHandler_UpdateConquista_BadJSON(t *testing.T) {
	repo := &mockCurriculoRepoH{
		conquista: &empmodels.CurriculoConquista{
			ID:  uuid.MustParse(validUUID),
			CPF: "12345678900",
		},
	}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{bad}`)
	req := httptest.NewRequest(http.MethodPut, "/conquistas/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests: ReplaceAllFormacoesByCPF (0% coverage)
// ──────────────────────────────────────────────────────────────────────────────

func TestCurriculoHandler_ReplaceAllFormacoesByCPF_Forbidden(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "99999999999", false)
	body := bodyOf(`{"formacoes":[],"idiomas":[]}`)
	req := httptest.NewRequest(http.MethodPut, "/curriculo/12345678900/formacoes", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestCurriculoHandler_ReplaceAllFormacoesByCPF_BadJSON(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{bad}`)
	req := httptest.NewRequest(http.MethodPut, "/curriculo/12345678900/formacoes", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCurriculoHandler_ReplaceAllFormacoesByCPF_Success(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"formacoes":[],"idiomas":[]}`)
	req := httptest.NewRequest(http.MethodPut, "/curriculo/12345678900/formacoes", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests: ReplaceAllExperienciasByCPF (0% coverage)
// ──────────────────────────────────────────────────────────────────────────────

func TestCurriculoHandler_ReplaceAllExperienciasByCPF_Forbidden(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "99999999999", false)
	body := bodyOf(`{"experiencias":[],"conquistas":[]}`)
	req := httptest.NewRequest(http.MethodPut, "/curriculo/12345678900/experiencias", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestCurriculoHandler_ReplaceAllExperienciasByCPF_BadJSON(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{bad}`)
	req := httptest.NewRequest(http.MethodPut, "/curriculo/12345678900/experiencias", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCurriculoHandler_ReplaceAllExperienciasByCPF_Success(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"experiencias":[],"conquistas":[]}`)
	req := httptest.NewRequest(http.MethodPut, "/curriculo/12345678900/experiencias", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests: ReplaceAllConquistasByCPF (0% coverage)
// ──────────────────────────────────────────────────────────────────────────────

func TestCurriculoHandler_ReplaceAllConquistasByCPF_Forbidden(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "99999999999", false)
	body := bodyOf(`[]`)
	req := httptest.NewRequest(http.MethodPut, "/curriculo/12345678900/conquistas", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestCurriculoHandler_ReplaceAllConquistasByCPF_BadJSON(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{bad}`)
	req := httptest.NewRequest(http.MethodPut, "/curriculo/12345678900/conquistas", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCurriculoHandler_ReplaceAllConquistasByCPF_Success(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`[]`)
	req := httptest.NewRequest(http.MethodPut, "/curriculo/12345678900/conquistas", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests: ReplaceAllIdiomasByCPF (0% coverage)
// ──────────────────────────────────────────────────────────────────────────────

func TestCurriculoHandler_ReplaceAllIdiomasByCPF_Forbidden(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "99999999999", false)
	body := bodyOf(`[]`)
	req := httptest.NewRequest(http.MethodPut, "/curriculo/12345678900/idiomas", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestCurriculoHandler_ReplaceAllIdiomasByCPF_BadJSON(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{bad}`)
	req := httptest.NewRequest(http.MethodPut, "/curriculo/12345678900/idiomas", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCurriculoHandler_ReplaceAllIdiomasByCPF_Success(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`[]`)
	req := httptest.NewRequest(http.MethodPut, "/curriculo/12345678900/idiomas", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests: ReplaceAllCursosComplementaresByCPF (0% coverage)
// ──────────────────────────────────────────────────────────────────────────────

func TestCurriculoHandler_ReplaceAllCursosComplementaresByCPF_Forbidden(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "99999999999", false)
	body := bodyOf(`[]`)
	req := httptest.NewRequest(http.MethodPut, "/curriculo/12345678900/cursos-complementares", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestCurriculoHandler_ReplaceAllCursosComplementaresByCPF_BadJSON(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{bad}`)
	req := httptest.NewRequest(http.MethodPut, "/curriculo/12345678900/cursos-complementares", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCurriculoHandler_ReplaceAllCursosComplementaresByCPF_Success(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`[]`)
	req := httptest.NewRequest(http.MethodPut, "/curriculo/12345678900/cursos-complementares", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Additional CurriculoHandler Edge Case Tests
// ──────────────────────────────────────────────────────────────────────────────

func TestCurriculoHandler_GetCurriculoCompleto_ServiceError(t *testing.T) {
	repo := &mockCurriculoRepoH{err: errTest}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/12345678900", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestCurriculoHandler_CreateFormacao_ServiceError(t *testing.T) {
	repo := &mockCurriculoRepoH{err: errTest}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"instituicao":"UFRJ"}`)
	req := httptest.NewRequest(http.MethodPost, "/curriculo/formacoes", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusCreated {
		t.Error("expected error status")
	}
}

func TestCurriculoHandler_CreateFormacao_BadJSON(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{bad}`)
	req := httptest.NewRequest(http.MethodPost, "/curriculo/formacoes", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCurriculoHandler_GetFormacaoByID_ServiceError(t *testing.T) {
	repo := &mockCurriculoRepoH{err: errTest}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/formacoes/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestCurriculoHandler_UpdateFormacao_ServiceError(t *testing.T) {
	id := uuid.MustParse(validUUID)
	repo := &mockCurriculoRepoH{
		formacao: &empmodels.CurriculoFormacao{ID: id, CPF: "12345678900"},
		err:      errTest,
	}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"instituicao":"PUC"}`)
	req := httptest.NewRequest(http.MethodPut, "/curriculo/formacoes/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestCurriculoHandler_UpdateFormacao_Forbidden(t *testing.T) {
	id := uuid.MustParse(validUUID)
	repo := &mockCurriculoRepoH{formacao: &empmodels.CurriculoFormacao{ID: id, CPF: "12345678900"}}
	r := setupCurriculoRouter(repo, "99999999999", false)
	body := bodyOf(`{"instituicao":"PUC"}`)
	req := httptest.NewRequest(http.MethodPut, "/curriculo/formacoes/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestCurriculoHandler_UpdateFormacao_BadJSON(t *testing.T) {
	id := uuid.MustParse(validUUID)
	repo := &mockCurriculoRepoH{formacao: &empmodels.CurriculoFormacao{ID: id, CPF: "12345678900"}}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{bad}`)
	req := httptest.NewRequest(http.MethodPut, "/curriculo/formacoes/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCurriculoHandler_UpdateFormacao_GetByIDError(t *testing.T) {
	repo := &mockCurriculoRepoH{err: errTest}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"instituicao":"PUC"}`)
	req := httptest.NewRequest(http.MethodPut, "/curriculo/formacoes/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestCurriculoHandler_DeleteFormacao_ServiceError(t *testing.T) {
	id := uuid.MustParse(validUUID)
	repo := &mockCurriculoRepoH{
		formacao: &empmodels.CurriculoFormacao{ID: id, CPF: "12345678900"},
		err:      errTest,
	}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodDelete, "/curriculo/formacoes/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestCurriculoHandler_DeleteFormacao_Forbidden(t *testing.T) {
	id := uuid.MustParse(validUUID)
	repo := &mockCurriculoRepoH{formacao: &empmodels.CurriculoFormacao{ID: id, CPF: "12345678900"}}
	r := setupCurriculoRouter(repo, "99999999999", false)
	req := httptest.NewRequest(http.MethodDelete, "/curriculo/formacoes/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestCurriculoHandler_DeleteFormacao_GetByIDError(t *testing.T) {
	repo := &mockCurriculoRepoH{err: errTest}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodDelete, "/curriculo/formacoes/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestCurriculoHandler_DeleteFormacao_InvalidID(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodDelete, "/curriculo/formacoes/bad-id", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCurriculoHandler_ListFormacoesByCPF_ServiceError(t *testing.T) {
	repo := &mockCurriculoRepoH{err: errTest}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/12345678900/formacoes", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestCurriculoHandler_ListFormacoesByCPF_Forbidden(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "99999999999", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/12345678900/formacoes", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestCurriculoHandler_CreateIdioma_BadJSON(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{bad}`)
	req := httptest.NewRequest(http.MethodPost, "/curriculo/idiomas", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCurriculoHandler_CreateIdioma_ServiceError(t *testing.T) {
	repo := &mockCurriculoRepoH{err: errTest}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"id_idioma":"` + validUUID + `","id_nivel_idioma":"` + validUUID + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/curriculo/idiomas", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusCreated {
		t.Error("expected error status")
	}
}

func TestCurriculoHandler_GetIdiomaByID_ServiceError(t *testing.T) {
	repo := &mockCurriculoRepoH{err: errTest}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/idiomas/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestCurriculoHandler_UpdateIdioma_Forbidden(t *testing.T) {
	id := uuid.MustParse(validUUID)
	repo := &mockCurriculoRepoH{idioma: &empmodels.CurriculoIdioma{ID: id, CPF: "12345678900"}}
	r := setupCurriculoRouter(repo, "99999999999", false)
	body := bodyOf(`{"id_idioma":"` + validUUID + `"}`)
	req := httptest.NewRequest(http.MethodPut, "/curriculo/idiomas/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestCurriculoHandler_DeleteIdioma_Forbidden(t *testing.T) {
	id := uuid.MustParse(validUUID)
	repo := &mockCurriculoRepoH{idioma: &empmodels.CurriculoIdioma{ID: id, CPF: "12345678900"}}
	r := setupCurriculoRouter(repo, "99999999999", false)
	req := httptest.NewRequest(http.MethodDelete, "/curriculo/idiomas/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestCurriculoHandler_CreateCurso_Unauthorized(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "", false)
	body := bodyOf(`{"nome":"Curso"}`)
	req := httptest.NewRequest(http.MethodPost, "/curriculo/cursos", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestCurriculoHandler_CreateCurso_BadJSON(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{bad}`)
	req := httptest.NewRequest(http.MethodPost, "/curriculo/cursos", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCurriculoHandler_GetCursoByID_ServiceError(t *testing.T) {
	repo := &mockCurriculoRepoH{err: errTest}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/cursos/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestCurriculoHandler_GetCursoByID_InvalidID(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/cursos/bad-id", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCurriculoHandler_UpdateCursoComplementar_ServiceError(t *testing.T) {
	repo := &mockCurriculoRepoH{
		curso: &empmodels.CurriculoCursoComplementar{
			ID:  uuid.MustParse(validUUID),
			CPF: "12345678900",
		},
		err: errTest,
	}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"nome":"Curso"}`)
	req := httptest.NewRequest(http.MethodPut, "/cursos/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestCurriculoHandler_UpdateCursoComplementar_GetByIDError(t *testing.T) {
	repo := &mockCurriculoRepoH{err: errTest}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"nome":"Curso"}`)
	req := httptest.NewRequest(http.MethodPut, "/cursos/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestCurriculoHandler_DeleteCurso_Forbidden(t *testing.T) {
	id := uuid.MustParse(validUUID)
	repo := &mockCurriculoRepoH{curso: &empmodels.CurriculoCursoComplementar{ID: id, CPF: "12345678900"}}
	r := setupCurriculoRouter(repo, "99999999999", false)
	req := httptest.NewRequest(http.MethodDelete, "/curriculo/cursos/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestCurriculoHandler_CreateExperiencia_Unauthorized(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "", false)
	body := bodyOf(`{"cargo":"Dev"}`)
	req := httptest.NewRequest(http.MethodPost, "/curriculo/experiencias", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestCurriculoHandler_CreateExperiencia_BadJSON(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{bad}`)
	req := httptest.NewRequest(http.MethodPost, "/curriculo/experiencias", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCurriculoHandler_GetExperienciaByID_ServiceError(t *testing.T) {
	repo := &mockCurriculoRepoH{err: errTest}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/experiencias/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestCurriculoHandler_GetExperienciaByID_InvalidID(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/experiencias/bad-id", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCurriculoHandler_UpdateExperiencia_ServiceError(t *testing.T) {
	repo := &mockCurriculoRepoH{
		experiencia: &empmodels.CurriculoExperiencia{
			ID:  uuid.MustParse(validUUID),
			CPF: "12345678900",
		},
		err: errTest,
	}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"cargo":"Senior"}`)
	req := httptest.NewRequest(http.MethodPut, "/experiencias/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestCurriculoHandler_UpdateExperiencia_GetByIDError(t *testing.T) {
	repo := &mockCurriculoRepoH{err: errTest}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"cargo":"Senior"}`)
	req := httptest.NewRequest(http.MethodPut, "/experiencias/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestCurriculoHandler_DeleteExperiencia_Forbidden(t *testing.T) {
	id := uuid.MustParse(validUUID)
	repo := &mockCurriculoRepoH{experiencia: &empmodels.CurriculoExperiencia{ID: id, CPF: "12345678900"}}
	r := setupCurriculoRouter(repo, "99999999999", false)
	req := httptest.NewRequest(http.MethodDelete, "/curriculo/experiencias/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestCurriculoHandler_CreateConquista_Unauthorized(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "", false)
	body := bodyOf(`{"titulo":"Premio"}`)
	req := httptest.NewRequest(http.MethodPost, "/curriculo/conquistas", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestCurriculoHandler_CreateConquista_BadJSON(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{bad}`)
	req := httptest.NewRequest(http.MethodPost, "/curriculo/conquistas", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCurriculoHandler_GetConquistaByID_ServiceError(t *testing.T) {
	repo := &mockCurriculoRepoH{err: errTest}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/conquistas/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestCurriculoHandler_GetConquistaByID_InvalidID(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/conquistas/bad-id", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCurriculoHandler_UpdateConquista_ServiceError(t *testing.T) {
	repo := &mockCurriculoRepoH{
		conquista: &empmodels.CurriculoConquista{
			ID:  uuid.MustParse(validUUID),
			CPF: "12345678900",
		},
		err: errTest,
	}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"titulo":"Premio"}`)
	req := httptest.NewRequest(http.MethodPut, "/conquistas/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestCurriculoHandler_UpdateConquista_GetByIDError(t *testing.T) {
	repo := &mockCurriculoRepoH{err: errTest}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"titulo":"Premio"}`)
	req := httptest.NewRequest(http.MethodPut, "/conquistas/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestCurriculoHandler_DeleteConquista_Forbidden(t *testing.T) {
	id := uuid.MustParse(validUUID)
	repo := &mockCurriculoRepoH{conquista: &empmodels.CurriculoConquista{ID: id, CPF: "12345678900"}}
	r := setupCurriculoRouter(repo, "99999999999", false)
	req := httptest.NewRequest(http.MethodDelete, "/curriculo/conquistas/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestCurriculoHandler_ReplaceAllFormacoesByCPF_ServiceError(t *testing.T) {
	repo := &mockCurriculoRepoH{err: errTest}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"formacoes":[],"idiomas":[]}`)
	req := httptest.NewRequest(http.MethodPut, "/curriculo/12345678900/formacoes", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestCurriculoHandler_ReplaceAllExperienciasByCPF_ServiceError(t *testing.T) {
	repo := &mockCurriculoRepoH{err: errTest}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"experiencias":[],"conquistas":[]}`)
	req := httptest.NewRequest(http.MethodPut, "/curriculo/12345678900/experiencias", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestCurriculoHandler_ReplaceAllConquistasByCPF_ServiceError(t *testing.T) {
	repo := &mockCurriculoRepoH{err: errTest}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`[]`)
	req := httptest.NewRequest(http.MethodPut, "/curriculo/12345678900/conquistas", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestCurriculoHandler_ReplaceAllIdiomasByCPF_ServiceError(t *testing.T) {
	repo := &mockCurriculoRepoH{err: errTest}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`[]`)
	req := httptest.NewRequest(http.MethodPut, "/curriculo/12345678900/idiomas", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestCurriculoHandler_ReplaceAllCursosComplementaresByCPF_ServiceError(t *testing.T) {
	repo := &mockCurriculoRepoH{err: errTest}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`[]`)
	req := httptest.NewRequest(http.MethodPut, "/curriculo/12345678900/cursos-complementares", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestCurriculoHandler_UpsertSituacaoInteresses_BadJSON(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{bad}`)
	req := httptest.NewRequest(http.MethodPut, "/curriculo/12345678900/situacao-interesses", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCurriculoHandler_GetSituacaoInteressesByCPF_Forbidden(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "99999999999", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/12345678900/situacao-interesses", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestCurriculoHandler_GetSituacaoInteressesByCPF_ServiceError(t *testing.T) {
	repo := &mockCurriculoRepoH{err: errTest}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/12345678900/situacao-interesses", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Additional Delete Method Tests for Better Coverage
// ──────────────────────────────────────────────────────────────────────────────

func TestCurriculoHandler_DeleteIdioma_InvalidID(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodDelete, "/curriculo/idiomas/bad-id", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCurriculoHandler_DeleteIdioma_NotFound(t *testing.T) {
	repo := &mockCurriculoRepoH{idioma: nil}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodDelete, "/curriculo/idiomas/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestCurriculoHandler_DeleteIdioma_ServiceError(t *testing.T) {
	id := uuid.MustParse(validUUID)
	repo := &mockCurriculoRepoH{
		idioma: &empmodels.CurriculoIdioma{ID: id, CPF: "12345678900"},
		err:    errTest,
	}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodDelete, "/curriculo/idiomas/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestCurriculoHandler_DeleteIdioma_GetByIDError(t *testing.T) {
	repo := &mockCurriculoRepoH{err: errTest}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodDelete, "/curriculo/idiomas/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestCurriculoHandler_DeleteCurso_InvalidID(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodDelete, "/curriculo/cursos/bad-id", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCurriculoHandler_DeleteCurso_NotFound(t *testing.T) {
	repo := &mockCurriculoRepoH{curso: nil}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodDelete, "/curriculo/cursos/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestCurriculoHandler_DeleteCurso_ServiceError(t *testing.T) {
	id := uuid.MustParse(validUUID)
	repo := &mockCurriculoRepoH{
		curso: &empmodels.CurriculoCursoComplementar{ID: id, CPF: "12345678900"},
		err:   errTest,
	}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodDelete, "/curriculo/cursos/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestCurriculoHandler_DeleteCurso_GetByIDError(t *testing.T) {
	repo := &mockCurriculoRepoH{err: errTest}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodDelete, "/curriculo/cursos/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestCurriculoHandler_DeleteExperiencia_InvalidID(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodDelete, "/curriculo/experiencias/bad-id", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCurriculoHandler_DeleteExperiencia_NotFound(t *testing.T) {
	repo := &mockCurriculoRepoH{experiencia: nil}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodDelete, "/curriculo/experiencias/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestCurriculoHandler_DeleteExperiencia_ServiceError(t *testing.T) {
	id := uuid.MustParse(validUUID)
	repo := &mockCurriculoRepoH{
		experiencia: &empmodels.CurriculoExperiencia{ID: id, CPF: "12345678900"},
		err:         errTest,
	}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodDelete, "/curriculo/experiencias/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestCurriculoHandler_DeleteExperiencia_GetByIDError(t *testing.T) {
	repo := &mockCurriculoRepoH{err: errTest}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodDelete, "/curriculo/experiencias/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestCurriculoHandler_DeleteConquista_InvalidID(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodDelete, "/curriculo/conquistas/bad-id", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCurriculoHandler_DeleteConquista_NotFound(t *testing.T) {
	repo := &mockCurriculoRepoH{conquista: nil}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodDelete, "/curriculo/conquistas/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestCurriculoHandler_DeleteConquista_ServiceError(t *testing.T) {
	id := uuid.MustParse(validUUID)
	repo := &mockCurriculoRepoH{
		conquista: &empmodels.CurriculoConquista{ID: id, CPF: "12345678900"},
		err:       errTest,
	}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodDelete, "/curriculo/conquistas/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestCurriculoHandler_DeleteConquista_GetByIDError(t *testing.T) {
	repo := &mockCurriculoRepoH{err: errTest}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodDelete, "/curriculo/conquistas/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestCurriculoHandler_UpdateIdioma_ServiceError(t *testing.T) {
	id := uuid.MustParse(validUUID)
	repo := &mockCurriculoRepoH{
		idioma: &empmodels.CurriculoIdioma{ID: id, CPF: "12345678900"},
		err:    errTest,
	}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"id_idioma":"` + validUUID + `"}`)
	req := httptest.NewRequest(http.MethodPut, "/curriculo/idiomas/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestCurriculoHandler_UpdateIdioma_NotFound(t *testing.T) {
	repo := &mockCurriculoRepoH{idioma: nil}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"id_idioma":"` + validUUID + `"}`)
	req := httptest.NewRequest(http.MethodPut, "/curriculo/idiomas/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestCurriculoHandler_UpdateIdioma_InvalidID(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"id_idioma":"` + validUUID + `"}`)
	req := httptest.NewRequest(http.MethodPut, "/curriculo/idiomas/bad-id", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCurriculoHandler_UpdateIdioma_BadJSON(t *testing.T) {
	id := uuid.MustParse(validUUID)
	repo := &mockCurriculoRepoH{idioma: &empmodels.CurriculoIdioma{ID: id, CPF: "12345678900"}}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{bad}`)
	req := httptest.NewRequest(http.MethodPut, "/curriculo/idiomas/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCurriculoHandler_UpdateIdioma_GetByIDError(t *testing.T) {
	repo := &mockCurriculoRepoH{err: errTest}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"id_idioma":"` + validUUID + `"}`)
	req := httptest.NewRequest(http.MethodPut, "/curriculo/idiomas/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestCurriculoHandler_ListIdiomasByCPF_Forbidden(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "99999999999", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/12345678900/idiomas", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestCurriculoHandler_ListIdiomasByCPF_ServiceError(t *testing.T) {
	repo := &mockCurriculoRepoH{err: errTest}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/12345678900/idiomas", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestCurriculoHandler_ListCursosByCPF_Forbidden(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "99999999999", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/12345678900/cursos", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestCurriculoHandler_ListCursosByCPF_ServiceError(t *testing.T) {
	repo := &mockCurriculoRepoH{err: errTest}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/12345678900/cursos", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestCurriculoHandler_ListExperienciasByCPF_Forbidden(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "99999999999", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/12345678900/experiencias", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestCurriculoHandler_ListExperienciasByCPF_ServiceError(t *testing.T) {
	repo := &mockCurriculoRepoH{err: errTest}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/12345678900/experiencias", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestCurriculoHandler_ListConquistasByCPF_Forbidden(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "99999999999", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/12345678900/conquistas", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestCurriculoHandler_ListConquistasByCPF_ServiceError(t *testing.T) {
	repo := &mockCurriculoRepoH{err: errTest}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/curriculo/12345678900/conquistas", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestCurriculoHandler_UpsertSituacaoInteresses_ServiceError(t *testing.T) {
	repo := &mockCurriculoRepoH{err: errTest}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{}`)
	req := httptest.NewRequest(http.MethodPut, "/curriculo/12345678900/situacao-interesses", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestCurriculoHandler_UpsertSituacaoInteresses_Unauthorized(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "", false)
	body := bodyOf(`{}`)
	req := httptest.NewRequest(http.MethodPut, "/curriculo/12345678900/situacao-interesses", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Additional error path tests for nested operations
// ──────────────────────────────────────────────────────────────────────────────

func TestCurriculoHandler_CreateConquista_ServiceError(t *testing.T) {
	repo := &mockCurriculoRepoH{err: errTest}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"titulo":"Premio Inovacao","id_tipo_conquista":"` + validUUID + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/curriculo/conquistas", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusCreated {
		t.Error("expected error status, got 201")
	}
}

func TestCurriculoHandler_CreateExperiencia_ServiceError(t *testing.T) {
	repo := &mockCurriculoRepoH{err: errTest}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"cargo":"Desenvolvedor","empresa":"Acme Inc"}`)
	req := httptest.NewRequest(http.MethodPost, "/curriculo/experiencias", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusCreated {
		t.Error("expected error status, got 201")
	}
}

func TestCurriculoHandler_CreateCursoComplementar_Success(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"nome":"Curso de Go","instituicao":"Online Academy"}`)
	req := httptest.NewRequest(http.MethodPost, "/curriculo/cursos", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCurriculoHandler_CreateCursoComplementar_Unauthorized(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "", false)
	body := bodyOf(`{"nome":"Curso de Go"}`)
	req := httptest.NewRequest(http.MethodPost, "/curriculo/cursos", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestCurriculoHandler_CreateCursoComplementar_BadJSON(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{bad}`)
	req := httptest.NewRequest(http.MethodPost, "/curriculo/cursos", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCurriculoHandler_CreateCursoComplementar_ServiceError(t *testing.T) {
	repo := &mockCurriculoRepoH{err: errTest}
	r := setupCurriculoRouter(repo, "12345678900", false)
	body := bodyOf(`{"nome":"Curso de Go","instituicao":"Online Academy"}`)
	req := httptest.NewRequest(http.MethodPost, "/curriculo/cursos", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusCreated {
		t.Error("expected error status, got 201")
	}
}

func TestCurriculoHandler_DeleteCursoComplementar_AsOwner(t *testing.T) {
	id := uuid.MustParse(validUUID)
	repo := &mockCurriculoRepoH{curso: &empmodels.CurriculoCursoComplementar{ID: id, CPF: "12345678900"}}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodDelete, "/curriculo/cursos/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestCurriculoHandler_DeleteCursoComplementar_Forbidden(t *testing.T) {
	id := uuid.MustParse(validUUID)
	repo := &mockCurriculoRepoH{curso: &empmodels.CurriculoCursoComplementar{ID: id, CPF: "00000000000"}}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodDelete, "/curriculo/cursos/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestCurriculoHandler_DeleteCursoComplementar_InvalidID(t *testing.T) {
	repo := &mockCurriculoRepoH{}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodDelete, "/curriculo/cursos/bad-id", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCurriculoHandler_DeleteCursoComplementar_NotFound(t *testing.T) {
	repo := &mockCurriculoRepoH{curso: nil}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodDelete, "/curriculo/cursos/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestCurriculoHandler_DeleteCursoComplementar_ServiceError(t *testing.T) {
	id := uuid.MustParse(validUUID)
	repo := &mockCurriculoRepoH{
		curso: &empmodels.CurriculoCursoComplementar{ID: id, CPF: "12345678900"},
		err:   errTest,
	}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodDelete, "/curriculo/cursos/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status, got 200")
	}
}

func TestCurriculoHandler_DeleteCursoComplementar_GetByIDError(t *testing.T) {
	repo := &mockCurriculoRepoH{err: errTest}
	r := setupCurriculoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodDelete, "/curriculo/cursos/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status, got 200")
	}
}
