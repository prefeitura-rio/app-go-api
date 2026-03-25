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
// Mock CurriculoRepository (implements CurriculoRepositoryInterface)
// ──────────────────────────────────────────────────────────────────────────────

type mockCurriculoRepoH struct {
	formacao       *empmodels.CurriculoFormacao
	idioma         *empmodels.CurriculoIdioma
	curso          *empmodels.CurriculoCursoComplementar
	experiencia    *empmodels.CurriculoExperiencia
	conquista      *empmodels.CurriculoConquista
	situacao       *empmodels.CurriculoSituacaoInteresses
	err            error
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

func (m *mockCurriculoRepoH) ReplaceAllExperienciaProfissionalAccordionByCPF(_ context.Context, _ string, _ []*empmodels.CurriculoExperiencia, _ []*empmodels.CurriculoConquista) error {
	return m.err
}

func (m *mockCurriculoRepoH) ReplaceAllConquistasByCPF(_ context.Context, _ string, _ []*empmodels.CurriculoConquista) error {
	return m.err
}

func (m *mockCurriculoRepoH) ReplaceAllIdiomasByCPF(_ context.Context, _ string, _ []*empmodels.CurriculoIdioma) error {
	return m.err
}

func (m *mockCurriculoRepoH) ReplaceAllCursosComplementaresByCPF(_ context.Context, _ string, _ []*empmodels.CurriculoCursoComplementar) error {
	return m.err
}

// Situação e Interesses

func (m *mockCurriculoRepoH) UpsertSituacaoInteresses(_ context.Context, _ *empmodels.CurriculoSituacaoInteresses) error {
	return m.err
}

func (m *mockCurriculoRepoH) GetSituacaoInteressesByCPF(_ context.Context, _ string) (*empmodels.CurriculoSituacaoInteresses, error) {
	return m.situacao, m.err
}

// ──────────────────────────────────────────────────────────────────────────────
// Router setup
// ──────────────────────────────────────────────────────────────────────────────

func setupCurriculoRouter(repo services.CurriculoRepositoryInterface, cpf string, isAdmin bool) *gin.Engine {
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
	r.DELETE("/curriculo/cursos/:id", h.DeleteCursoComplementar)
	r.GET("/curriculo/:cpf/cursos", h.ListCursosComplementaresByCPF)

	r.POST("/curriculo/experiencias", h.CreateExperiencia)
	r.GET("/curriculo/experiencias/:id", h.GetExperienciaByID)
	r.PUT("/curriculo/experiencias/:id", h.UpdateExperiencia)
	r.DELETE("/curriculo/experiencias/:id", h.DeleteExperiencia)
	r.GET("/curriculo/:cpf/experiencias", h.ListExperienciasByCPF)

	r.POST("/curriculo/conquistas", h.CreateConquista)
	r.GET("/curriculo/conquistas/:id", h.GetConquistaByID)
	r.PUT("/curriculo/conquistas/:id", h.UpdateConquista)
	r.DELETE("/curriculo/conquistas/:id", h.DeleteConquista)
	r.GET("/curriculo/:cpf/conquistas", h.ListConquistasByCPF)

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
	if w.Code != http.StatusUnauthorized {
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
