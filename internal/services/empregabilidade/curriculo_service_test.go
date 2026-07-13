package empregabilidade_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
	"github.com/stretchr/testify/assert"
)

func TestNewCurriculoService(t *testing.T) {
	mockRepo := &mockCurriculoRepo{}
	service := services.NewCurriculoServiceWithInterface(mockRepo)
	assert.NotNil(t, service)
}

// ──────────────────────────────────────────────────────────────────────────────
// Mock Curriculo Repository
// ──────────────────────────────────────────────────────────────────────────────

type mockCurriculoRepo struct {
	formacoes    []*empregabilidade.CurriculoFormacao
	idiomas      []*empregabilidade.CurriculoIdioma
	cursos       []*empregabilidade.CurriculoCursoComplementar
	experiencias []*empregabilidade.CurriculoExperiencia
	conquistas   []*empregabilidade.CurriculoConquista
	situacao     *empregabilidade.CurriculoSituacaoInteresses
	perfil       *empregabilidade.CurriculoPerfil
	err          error
}

func (m *mockCurriculoRepo) CreateFormacao(_ context.Context, _ *empregabilidade.CurriculoFormacao) (uuid.UUID, error) {
	return uuid.New(), m.err
}

func (m *mockCurriculoRepo) GetFormacaoByID(_ context.Context, id uuid.UUID) (*empregabilidade.CurriculoFormacao, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, f := range m.formacoes {
		if f.ID == id {
			return f, nil
		}
	}
	return nil, nil
}

func (m *mockCurriculoRepo) UpdateFormacao(_ context.Context, _ *empregabilidade.CurriculoFormacao) error {
	return m.err
}

func (m *mockCurriculoRepo) DeleteFormacao(_ context.Context, _ uuid.UUID) error {
	return m.err
}

func (m *mockCurriculoRepo) ListFormacoesByCPF(_ context.Context, _ string) ([]*empregabilidade.CurriculoFormacao, error) {
	return m.formacoes, m.err
}

func (m *mockCurriculoRepo) CreateIdioma(_ context.Context, _ *empregabilidade.CurriculoIdioma) (uuid.UUID, error) {
	return uuid.New(), m.err
}

func (m *mockCurriculoRepo) GetIdiomaByID(_ context.Context, _ uuid.UUID) (*empregabilidade.CurriculoIdioma, error) {
	if len(m.idiomas) > 0 {
		return m.idiomas[0], m.err
	}
	return nil, m.err
}

func (m *mockCurriculoRepo) UpdateIdioma(_ context.Context, _ *empregabilidade.CurriculoIdioma) error {
	return m.err
}

func (m *mockCurriculoRepo) DeleteIdioma(_ context.Context, _ uuid.UUID) error {
	return m.err
}

func (m *mockCurriculoRepo) ListIdiomasByCPF(_ context.Context, _ string) ([]*empregabilidade.CurriculoIdioma, error) {
	return m.idiomas, m.err
}

func (m *mockCurriculoRepo) CreateCursoComplementar(_ context.Context, _ *empregabilidade.CurriculoCursoComplementar) (uuid.UUID, error) {
	return uuid.New(), m.err
}

func (m *mockCurriculoRepo) GetCursoComplementarByID(_ context.Context, _ uuid.UUID) (*empregabilidade.CurriculoCursoComplementar, error) {
	if len(m.cursos) > 0 {
		return m.cursos[0], m.err
	}
	return nil, m.err
}

func (m *mockCurriculoRepo) UpdateCursoComplementar(_ context.Context, _ *empregabilidade.CurriculoCursoComplementar) error {
	return m.err
}

func (m *mockCurriculoRepo) DeleteCursoComplementar(_ context.Context, _ uuid.UUID) error {
	return m.err
}

func (m *mockCurriculoRepo) ListCursosComplementaresByCPF(_ context.Context, _ string) ([]*empregabilidade.CurriculoCursoComplementar, error) {
	return m.cursos, m.err
}

func (m *mockCurriculoRepo) CreateExperiencia(_ context.Context, _ *empregabilidade.CurriculoExperiencia) (uuid.UUID, error) {
	return uuid.New(), m.err
}

func (m *mockCurriculoRepo) GetExperienciaByID(_ context.Context, _ uuid.UUID) (*empregabilidade.CurriculoExperiencia, error) {
	if len(m.experiencias) > 0 {
		return m.experiencias[0], m.err
	}
	return nil, m.err
}

func (m *mockCurriculoRepo) UpdateExperiencia(_ context.Context, _ *empregabilidade.CurriculoExperiencia) error {
	return m.err
}

func (m *mockCurriculoRepo) DeleteExperiencia(_ context.Context, _ uuid.UUID) error {
	return m.err
}

func (m *mockCurriculoRepo) ListExperienciasByCPF(_ context.Context, _ string) ([]*empregabilidade.CurriculoExperiencia, error) {
	return m.experiencias, m.err
}

func (m *mockCurriculoRepo) CreateConquista(_ context.Context, _ *empregabilidade.CurriculoConquista) (uuid.UUID, error) {
	return uuid.New(), m.err
}

func (m *mockCurriculoRepo) GetConquistaByID(_ context.Context, _ uuid.UUID) (*empregabilidade.CurriculoConquista, error) {
	if len(m.conquistas) > 0 {
		return m.conquistas[0], m.err
	}
	return nil, m.err
}

func (m *mockCurriculoRepo) UpdateConquista(_ context.Context, _ *empregabilidade.CurriculoConquista) error {
	return m.err
}

func (m *mockCurriculoRepo) DeleteConquista(_ context.Context, _ uuid.UUID) error {
	return m.err
}

func (m *mockCurriculoRepo) ListConquistasByCPF(_ context.Context, _ string) ([]*empregabilidade.CurriculoConquista, error) {
	return m.conquistas, m.err
}

func (m *mockCurriculoRepo) ReplaceAllFormacoesByCPF(_ context.Context, _ string, _ []*empregabilidade.CurriculoFormacao) error {
	return m.err
}

func (m *mockCurriculoRepo) ReplaceAllFormacaoAccordionByCPF(_ context.Context, _ string, _ []*empregabilidade.CurriculoFormacao, _ []*empregabilidade.CurriculoIdioma) error {
	return m.err
}

func (m *mockCurriculoRepo) ReplaceAllExperienciasByCPF(_ context.Context, _ string, _ []*empregabilidade.CurriculoExperiencia) error {
	return m.err
}

func (m *mockCurriculoRepo) ReplaceAllExperienciaProfissionalAccordionByCPF(_ context.Context, _ string, _ []*empregabilidade.CurriculoExperiencia, _ []*empregabilidade.CurriculoConquista, _ string) error {
	return m.err
}

func (m *mockCurriculoRepo) ReplaceAllConquistasByCPF(_ context.Context, _ string, _ []*empregabilidade.CurriculoConquista) error {
	return m.err
}

func (m *mockCurriculoRepo) ReplaceAllIdiomasByCPF(_ context.Context, _ string, _ []*empregabilidade.CurriculoIdioma) error {
	return m.err
}

func (m *mockCurriculoRepo) ReplaceAllCursosComplementaresByCPF(_ context.Context, _ string, _ []*empregabilidade.CurriculoCursoComplementar) error {
	return m.err
}

func (m *mockCurriculoRepo) UpsertSituacaoInteresses(_ context.Context, _ *empregabilidade.CurriculoSituacaoInteresses) error {
	return m.err
}

func (m *mockCurriculoRepo) GetSituacaoInteressesByCPF(_ context.Context, _ string) (*empregabilidade.CurriculoSituacaoInteresses, error) {
	return m.situacao, m.err
}

func (m *mockCurriculoRepo) GetPerfilByCPF(_ context.Context, _ string) (*empregabilidade.CurriculoPerfil, error) {
	return m.perfil, m.err
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests: Formação
// ──────────────────────────────────────────────────────────────────────────────

func TestCurriculoService_CreateFormacao_ValidData(t *testing.T) {
	repo := &mockCurriculoRepo{}
	svc := services.NewCurriculoServiceWithInterface(repo)
	entity := &empregabilidade.CurriculoFormacao{
		CPF:             "12345678900",
		NomeInstituicao: "UFRJ",
		NomeCurso:       "Engenharia",
		Status:          empregabilidade.StatusFormacaoEmAndamento,
	}
	id, err := svc.CreateFormacao(context.Background(), entity)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if id == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
}

func TestCurriculoService_CreateFormacao_InvalidStatus(t *testing.T) {
	repo := &mockCurriculoRepo{}
	svc := services.NewCurriculoServiceWithInterface(repo)
	// Invalid status value
	entity := &empregabilidade.CurriculoFormacao{
		CPF:    "12345678900",
		Status: "INVALIDO",
	}
	_, err := svc.CreateFormacao(context.Background(), entity)
	if err == nil {
		t.Error("expected validation error for invalid status")
	}
}

func TestCurriculoService_GetFormacaoByID_Found(t *testing.T) {
	id := uuid.New()
	formacao := &empregabilidade.CurriculoFormacao{ID: id, CPF: "12345678900"}
	repo := &mockCurriculoRepo{formacoes: []*empregabilidade.CurriculoFormacao{formacao}}
	svc := services.NewCurriculoServiceWithInterface(repo)
	result, err := svc.GetFormacaoByID(context.Background(), id)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil {
		t.Error("expected formacao, got nil")
	}
}

func TestCurriculoService_GetFormacaoByID_NotFound(t *testing.T) {
	repo := &mockCurriculoRepo{formacoes: nil}
	svc := services.NewCurriculoServiceWithInterface(repo)
	result, err := svc.GetFormacaoByID(context.Background(), uuid.New())
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result != nil {
		t.Error("expected nil, got formacao")
	}
}

func TestCurriculoService_UpdateFormacao_ValidData(t *testing.T) {
	repo := &mockCurriculoRepo{}
	svc := services.NewCurriculoServiceWithInterface(repo)
	entity := &empregabilidade.CurriculoFormacao{
		CPF:             "12345678900",
		NomeInstituicao: "PUC",
		NomeCurso:       "Direito",
		Status:          empregabilidade.StatusFormacaoCompleto,
	}
	err := svc.UpdateFormacao(context.Background(), entity)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestCurriculoService_UpdateFormacao_InvalidStatus(t *testing.T) {
	repo := &mockCurriculoRepo{}
	svc := services.NewCurriculoServiceWithInterface(repo)
	entity := &empregabilidade.CurriculoFormacao{
		CPF:    "12345678900",
		Status: "INVALIDO",
	}
	err := svc.UpdateFormacao(context.Background(), entity)
	if err == nil {
		t.Error("expected validation error for invalid status")
	}
}

func TestCurriculoService_DeleteFormacao(t *testing.T) {
	repo := &mockCurriculoRepo{}
	svc := services.NewCurriculoServiceWithInterface(repo)
	err := svc.DeleteFormacao(context.Background(), uuid.New())
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestCurriculoService_ListFormacoesByCPF(t *testing.T) {
	formacoes := []*empregabilidade.CurriculoFormacao{
		{CPF: "12345678900", NomeInstituicao: "UFRJ"},
	}
	repo := &mockCurriculoRepo{formacoes: formacoes}
	svc := services.NewCurriculoServiceWithInterface(repo)
	result, err := svc.ListFormacoesByCPF(context.Background(), "12345678900")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 formacao, got %d", len(result))
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests: Idioma
// ──────────────────────────────────────────────────────────────────────────────

func TestCurriculoService_CreateIdioma(t *testing.T) {
	repo := &mockCurriculoRepo{}
	svc := services.NewCurriculoServiceWithInterface(repo)
	entity := &empregabilidade.CurriculoIdioma{CPF: "12345678900"}
	id, err := svc.CreateIdioma(context.Background(), entity)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if id == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
}

func TestCurriculoService_GetIdiomaByID_Found(t *testing.T) {
	id := uuid.New()
	repo := &mockCurriculoRepo{idiomas: []*empregabilidade.CurriculoIdioma{{ID: id}}}
	svc := services.NewCurriculoServiceWithInterface(repo)
	result, err := svc.GetIdiomaByID(context.Background(), id)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil {
		t.Error("expected idioma, got nil")
	}
}

func TestCurriculoService_UpdateIdioma(t *testing.T) {
	repo := &mockCurriculoRepo{}
	svc := services.NewCurriculoServiceWithInterface(repo)
	entity := &empregabilidade.CurriculoIdioma{CPF: "12345678900"}
	err := svc.UpdateIdioma(context.Background(), entity)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestCurriculoService_DeleteIdioma(t *testing.T) {
	repo := &mockCurriculoRepo{}
	svc := services.NewCurriculoServiceWithInterface(repo)
	err := svc.DeleteIdioma(context.Background(), uuid.New())
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestCurriculoService_ListIdiomasByCPF(t *testing.T) {
	idiomas := []*empregabilidade.CurriculoIdioma{{CPF: "12345678900"}}
	repo := &mockCurriculoRepo{idiomas: idiomas}
	svc := services.NewCurriculoServiceWithInterface(repo)
	result, err := svc.ListIdiomasByCPF(context.Background(), "12345678900")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1, got %d", len(result))
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests: Curso Complementar
// ──────────────────────────────────────────────────────────────────────────────

func TestCurriculoService_CreateCursoComplementar(t *testing.T) {
	repo := &mockCurriculoRepo{}
	svc := services.NewCurriculoServiceWithInterface(repo)
	entity := &empregabilidade.CurriculoCursoComplementar{CPF: "12345678900", NomeCurso: "Go Avancado"}
	id, err := svc.CreateCursoComplementar(context.Background(), entity)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if id == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
}

func TestCurriculoService_GetCursoComplementarByID_Found(t *testing.T) {
	id := uuid.New()
	repo := &mockCurriculoRepo{cursos: []*empregabilidade.CurriculoCursoComplementar{{ID: id}}}
	svc := services.NewCurriculoServiceWithInterface(repo)
	result, err := svc.GetCursoComplementarByID(context.Background(), id)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil {
		t.Error("expected curso, got nil")
	}
}

func TestCurriculoService_UpdateCursoComplementar(t *testing.T) {
	repo := &mockCurriculoRepo{}
	svc := services.NewCurriculoServiceWithInterface(repo)
	entity := &empregabilidade.CurriculoCursoComplementar{CPF: "12345678900"}
	err := svc.UpdateCursoComplementar(context.Background(), entity)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestCurriculoService_DeleteCursoComplementar(t *testing.T) {
	repo := &mockCurriculoRepo{}
	svc := services.NewCurriculoServiceWithInterface(repo)
	err := svc.DeleteCursoComplementar(context.Background(), uuid.New())
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestCurriculoService_ListCursosComplementaresByCPF(t *testing.T) {
	cursos := []*empregabilidade.CurriculoCursoComplementar{{CPF: "12345678900"}}
	repo := &mockCurriculoRepo{cursos: cursos}
	svc := services.NewCurriculoServiceWithInterface(repo)
	result, err := svc.ListCursosComplementaresByCPF(context.Background(), "12345678900")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1, got %d", len(result))
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests: Experiência
// ──────────────────────────────────────────────────────────────────────────────

func TestCurriculoService_CreateExperiencia(t *testing.T) {
	repo := &mockCurriculoRepo{}
	svc := services.NewCurriculoServiceWithInterface(repo)
	entity := &empregabilidade.CurriculoExperiencia{CPF: "12345678900"}
	id, err := svc.CreateExperiencia(context.Background(), entity)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if id == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
}

func TestCurriculoService_GetExperienciaByID_Found(t *testing.T) {
	id := uuid.New()
	repo := &mockCurriculoRepo{experiencias: []*empregabilidade.CurriculoExperiencia{{ID: id}}}
	svc := services.NewCurriculoServiceWithInterface(repo)
	result, err := svc.GetExperienciaByID(context.Background(), id)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil {
		t.Error("expected experiencia, got nil")
	}
}

func TestCurriculoService_UpdateExperiencia(t *testing.T) {
	repo := &mockCurriculoRepo{}
	svc := services.NewCurriculoServiceWithInterface(repo)
	entity := &empregabilidade.CurriculoExperiencia{CPF: "12345678900"}
	err := svc.UpdateExperiencia(context.Background(), entity)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestCurriculoService_DeleteExperiencia(t *testing.T) {
	repo := &mockCurriculoRepo{}
	svc := services.NewCurriculoServiceWithInterface(repo)
	err := svc.DeleteExperiencia(context.Background(), uuid.New())
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestCurriculoService_ListExperienciasByCPF(t *testing.T) {
	experiencias := []*empregabilidade.CurriculoExperiencia{{CPF: "12345678900"}}
	repo := &mockCurriculoRepo{experiencias: experiencias}
	svc := services.NewCurriculoServiceWithInterface(repo)
	result, err := svc.ListExperienciasByCPF(context.Background(), "12345678900")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1, got %d", len(result))
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests: Conquista
// ──────────────────────────────────────────────────────────────────────────────

func TestCurriculoService_CreateConquista(t *testing.T) {
	repo := &mockCurriculoRepo{}
	svc := services.NewCurriculoServiceWithInterface(repo)
	entity := &empregabilidade.CurriculoConquista{CPF: "12345678900"}
	id, err := svc.CreateConquista(context.Background(), entity)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if id == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
}

func TestCurriculoService_GetConquistaByID_Found(t *testing.T) {
	id := uuid.New()
	repo := &mockCurriculoRepo{conquistas: []*empregabilidade.CurriculoConquista{{ID: id}}}
	svc := services.NewCurriculoServiceWithInterface(repo)
	result, err := svc.GetConquistaByID(context.Background(), id)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil {
		t.Error("expected conquista, got nil")
	}
}

func TestCurriculoService_UpdateConquista(t *testing.T) {
	repo := &mockCurriculoRepo{}
	svc := services.NewCurriculoServiceWithInterface(repo)
	entity := &empregabilidade.CurriculoConquista{CPF: "12345678900"}
	err := svc.UpdateConquista(context.Background(), entity)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestCurriculoService_DeleteConquista(t *testing.T) {
	repo := &mockCurriculoRepo{}
	svc := services.NewCurriculoServiceWithInterface(repo)
	err := svc.DeleteConquista(context.Background(), uuid.New())
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestCurriculoService_ListConquistasByCPF(t *testing.T) {
	conquistas := []*empregabilidade.CurriculoConquista{{CPF: "12345678900"}}
	repo := &mockCurriculoRepo{conquistas: conquistas}
	svc := services.NewCurriculoServiceWithInterface(repo)
	result, err := svc.ListConquistasByCPF(context.Background(), "12345678900")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1, got %d", len(result))
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests: ReplaceAll
// ──────────────────────────────────────────────────────────────────────────────

func TestCurriculoService_ReplaceAllFormacoesByCPF(t *testing.T) {
	repo := &mockCurriculoRepo{}
	svc := services.NewCurriculoServiceWithInterface(repo)
	items := []*empregabilidade.CurriculoFormacao{}
	err := svc.ReplaceAllFormacoesByCPF(context.Background(), "12345678900", items)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestCurriculoService_ReplaceAllExperienciasByCPF(t *testing.T) {
	repo := &mockCurriculoRepo{}
	svc := services.NewCurriculoServiceWithInterface(repo)
	items := []*empregabilidade.CurriculoExperiencia{}
	err := svc.ReplaceAllExperienciasByCPF(context.Background(), "12345678900", items)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestCurriculoService_ReplaceAllConquistasByCPF(t *testing.T) {
	repo := &mockCurriculoRepo{}
	svc := services.NewCurriculoServiceWithInterface(repo)
	items := []*empregabilidade.CurriculoConquista{}
	err := svc.ReplaceAllConquistasByCPF(context.Background(), "12345678900", items)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestCurriculoService_ReplaceAllIdiomasByCPF(t *testing.T) {
	repo := &mockCurriculoRepo{}
	svc := services.NewCurriculoServiceWithInterface(repo)
	items := []*empregabilidade.CurriculoIdioma{}
	err := svc.ReplaceAllIdiomasByCPF(context.Background(), "12345678900", items)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestCurriculoService_ReplaceAllCursosComplementaresByCPF(t *testing.T) {
	repo := &mockCurriculoRepo{}
	svc := services.NewCurriculoServiceWithInterface(repo)
	items := []*empregabilidade.CurriculoCursoComplementar{}
	err := svc.ReplaceAllCursosComplementaresByCPF(context.Background(), "12345678900", items)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestCurriculoService_ReplaceAllFormacaoAccordionByCPF(t *testing.T) {
	repo := &mockCurriculoRepo{}
	svc := services.NewCurriculoServiceWithInterface(repo)
	formacoes := []*empregabilidade.CurriculoFormacao{}
	idiomas := []*empregabilidade.CurriculoIdioma{}
	err := svc.ReplaceAllFormacaoAccordionByCPF(context.Background(), "12345678900", formacoes, idiomas)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestCurriculoService_ReplaceAllExperienciaProfissionalAccordionByCPF(t *testing.T) {
	repo := &mockCurriculoRepo{}
	svc := services.NewCurriculoServiceWithInterface(repo)
	experiencias := []*empregabilidade.CurriculoExperiencia{}
	conquistas := []*empregabilidade.CurriculoConquista{}
	err := svc.ReplaceAllExperienciaProfissionalAccordionByCPF(context.Background(), "12345678900", experiencias, conquistas, "")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests: Situação e Interesses
// ──────────────────────────────────────────────────────────────────────────────

func TestCurriculoService_UpsertSituacaoInteresses(t *testing.T) {
	repo := &mockCurriculoRepo{}
	svc := services.NewCurriculoServiceWithInterface(repo)
	entity := &empregabilidade.CurriculoSituacaoInteresses{CPF: "12345678900"}
	err := svc.UpsertSituacaoInteresses(context.Background(), entity)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestCurriculoService_GetSituacaoInteressesByCPF_Found(t *testing.T) {
	situacao := &empregabilidade.CurriculoSituacaoInteresses{CPF: "12345678900"}
	repo := &mockCurriculoRepo{situacao: situacao}
	svc := services.NewCurriculoServiceWithInterface(repo)
	result, err := svc.GetSituacaoInteressesByCPF(context.Background(), "12345678900")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil {
		t.Error("expected situacao, got nil")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests: GetCurriculoCompleto
// ──────────────────────────────────────────────────────────────────────────────

func TestCurriculoService_GetCurriculoCompleto_Success(t *testing.T) {
	repo := &mockCurriculoRepo{
		formacoes:    []*empregabilidade.CurriculoFormacao{},
		idiomas:      []*empregabilidade.CurriculoIdioma{},
		cursos:       []*empregabilidade.CurriculoCursoComplementar{},
		experiencias: []*empregabilidade.CurriculoExperiencia{},
		conquistas:   []*empregabilidade.CurriculoConquista{},
		situacao:     nil,
	}
	svc := services.NewCurriculoServiceWithInterface(repo)
	result, err := svc.GetCurriculoCompleto(context.Background(), "12345678900")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil {
		t.Error("expected curriculo completo, got nil")
	}
}

func TestCurriculoService_GetCurriculoCompleto_WithData(t *testing.T) {
	repo := &mockCurriculoRepo{
		formacoes: []*empregabilidade.CurriculoFormacao{
			{CPF: "12345678900", NomeInstituicao: "UFRJ"},
		},
		idiomas: []*empregabilidade.CurriculoIdioma{
			{CPF: "12345678900"},
		},
		cursos: []*empregabilidade.CurriculoCursoComplementar{
			{CPF: "12345678900", NomeCurso: "Go Avancado"},
		},
		experiencias: []*empregabilidade.CurriculoExperiencia{
			{CPF: "12345678900", Cargo: "Desenvolvedor"},
		},
		conquistas: []*empregabilidade.CurriculoConquista{
			{CPF: "12345678900", Titulo: "Prêmio Excelência"},
		},
		situacao: &empregabilidade.CurriculoSituacaoInteresses{
			CPF: "12345678900",
		},
	}
	svc := services.NewCurriculoServiceWithInterface(repo)
	result, err := svc.GetCurriculoCompleto(context.Background(), "12345678900")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil {
		t.Error("expected curriculo completo, got nil")
	}
	if len(result.Formacoes) != 1 {
		t.Errorf("expected 1 formacao, got %d", len(result.Formacoes))
	}
	if len(result.Idiomas) != 1 {
		t.Errorf("expected 1 idioma, got %d", len(result.Idiomas))
	}
	if len(result.CursosComplementares) != 1 {
		t.Errorf("expected 1 curso, got %d", len(result.CursosComplementares))
	}
	if len(result.Experiencias) != 1 {
		t.Errorf("expected 1 experiencia, got %d", len(result.Experiencias))
	}
	if len(result.Conquistas) != 1 {
		t.Errorf("expected 1 conquista, got %d", len(result.Conquistas))
	}
	if result.SituacaoInteresses == nil {
		t.Error("expected situacao, got nil")
	}
}

func TestCurriculoService_GetCurriculoCompleto_ErrorFormacoes(t *testing.T) {
	repo := &mockCurriculoRepo{
		err: errors.New("formacoes error"),
	}
	svc := services.NewCurriculoServiceWithInterface(repo)
	result, err := svc.GetCurriculoCompleto(context.Background(), "12345678900")
	if err == nil {
		t.Error("expected error from formacoes, got nil")
	}
	if result != nil {
		t.Error("expected nil result when formacoes fails")
	}
	if err.Error() != "formacoes error" {
		t.Errorf("expected 'formacoes error', got '%s'", err.Error())
	}
}

func TestCurriculoService_GetCurriculoCompleto_ErrorIdiomas(t *testing.T) {
	customRepo := &mockCurriculoRepoWithSequentialErrors{
		step: 2, // Will fail on idiomas (step 2)
	}
	svc := services.NewCurriculoServiceWithInterface(customRepo)
	result, err := svc.GetCurriculoCompleto(context.Background(), "12345678900")
	if err == nil {
		t.Error("expected error from idiomas, got nil")
	}
	if result != nil {
		t.Error("expected nil result when idiomas fails")
	}
}

func TestCurriculoService_GetCurriculoCompleto_ErrorCursos(t *testing.T) {
	customRepo := &mockCurriculoRepoWithSequentialErrors{
		step: 3, // Will fail on cursos (step 3)
	}
	svc := services.NewCurriculoServiceWithInterface(customRepo)
	result, err := svc.GetCurriculoCompleto(context.Background(), "12345678900")
	if err == nil {
		t.Error("expected error from cursos, got nil")
	}
	if result != nil {
		t.Error("expected nil result when cursos fails")
	}
}

func TestCurriculoService_GetCurriculoCompleto_ErrorExperiencias(t *testing.T) {
	customRepo := &mockCurriculoRepoWithSequentialErrors{
		step: 4, // Will fail on experiencias (step 4)
	}
	svc := services.NewCurriculoServiceWithInterface(customRepo)
	result, err := svc.GetCurriculoCompleto(context.Background(), "12345678900")
	if err == nil {
		t.Error("expected error from experiencias, got nil")
	}
	if result != nil {
		t.Error("expected nil result when experiencias fails")
	}
}

func TestCurriculoService_GetCurriculoCompleto_ErrorConquistas(t *testing.T) {
	customRepo := &mockCurriculoRepoWithSequentialErrors{
		step: 5, // Will fail on conquistas (step 5)
	}
	svc := services.NewCurriculoServiceWithInterface(customRepo)
	result, err := svc.GetCurriculoCompleto(context.Background(), "12345678900")
	if err == nil {
		t.Error("expected error from conquistas, got nil")
	}
	if result != nil {
		t.Error("expected nil result when conquistas fails")
	}
}

func TestCurriculoService_GetCurriculoCompleto_ErrorSituacao(t *testing.T) {
	customRepo := &mockCurriculoRepoWithSequentialErrors{
		step: 5, // Will fail on situacao (step 6)
	}
	svc := services.NewCurriculoServiceWithInterface(customRepo)
	result, err := svc.GetCurriculoCompleto(context.Background(), "12345678900")
	if err == nil {
		t.Error("expected error from situacao, got nil")
	}
	if result != nil {
		t.Error("expected nil result when situacao fails")
	}
}

// mockCurriculoRepoWithSequentialErrors allows specific steps to fail
type mockCurriculoRepoWithSequentialErrors struct {
	step int // 0=success, 1=fail formacoes, 2=fail idiomas, etc.
}

func (m *mockCurriculoRepoWithSequentialErrors) CreateFormacao(_ context.Context, _ *empregabilidade.CurriculoFormacao) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (m *mockCurriculoRepoWithSequentialErrors) GetFormacaoByID(_ context.Context, _ uuid.UUID) (*empregabilidade.CurriculoFormacao, error) {
	return nil, nil
}

func (m *mockCurriculoRepoWithSequentialErrors) UpdateFormacao(_ context.Context, _ *empregabilidade.CurriculoFormacao) error {
	return nil
}

func (m *mockCurriculoRepoWithSequentialErrors) DeleteFormacao(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockCurriculoRepoWithSequentialErrors) ListFormacoesByCPF(_ context.Context, _ string) ([]*empregabilidade.CurriculoFormacao, error) {
	if m.step == 1 {
		return nil, errors.New("formacoes error")
	}
	return []*empregabilidade.CurriculoFormacao{}, nil
}

func (m *mockCurriculoRepoWithSequentialErrors) CreateIdioma(_ context.Context, _ *empregabilidade.CurriculoIdioma) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (m *mockCurriculoRepoWithSequentialErrors) GetIdiomaByID(_ context.Context, _ uuid.UUID) (*empregabilidade.CurriculoIdioma, error) {
	return nil, nil
}

func (m *mockCurriculoRepoWithSequentialErrors) UpdateIdioma(_ context.Context, _ *empregabilidade.CurriculoIdioma) error {
	return nil
}

func (m *mockCurriculoRepoWithSequentialErrors) DeleteIdioma(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockCurriculoRepoWithSequentialErrors) ListIdiomasByCPF(_ context.Context, _ string) ([]*empregabilidade.CurriculoIdioma, error) {
	if m.step == 2 {
		return nil, errors.New("idiomas error")
	}
	return []*empregabilidade.CurriculoIdioma{}, nil
}

func (m *mockCurriculoRepoWithSequentialErrors) CreateCursoComplementar(_ context.Context, _ *empregabilidade.CurriculoCursoComplementar) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (m *mockCurriculoRepoWithSequentialErrors) GetCursoComplementarByID(_ context.Context, _ uuid.UUID) (*empregabilidade.CurriculoCursoComplementar, error) {
	return nil, nil
}

func (m *mockCurriculoRepoWithSequentialErrors) UpdateCursoComplementar(_ context.Context, _ *empregabilidade.CurriculoCursoComplementar) error {
	return nil
}

func (m *mockCurriculoRepoWithSequentialErrors) DeleteCursoComplementar(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockCurriculoRepoWithSequentialErrors) ListCursosComplementaresByCPF(_ context.Context, _ string) ([]*empregabilidade.CurriculoCursoComplementar, error) {
	if m.step == 3 {
		return nil, errors.New("cursos error")
	}
	return []*empregabilidade.CurriculoCursoComplementar{}, nil
}

func (m *mockCurriculoRepoWithSequentialErrors) CreateExperiencia(_ context.Context, _ *empregabilidade.CurriculoExperiencia) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (m *mockCurriculoRepoWithSequentialErrors) GetExperienciaByID(_ context.Context, _ uuid.UUID) (*empregabilidade.CurriculoExperiencia, error) {
	return nil, nil
}

func (m *mockCurriculoRepoWithSequentialErrors) UpdateExperiencia(_ context.Context, _ *empregabilidade.CurriculoExperiencia) error {
	return nil
}

func (m *mockCurriculoRepoWithSequentialErrors) DeleteExperiencia(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockCurriculoRepoWithSequentialErrors) ListExperienciasByCPF(_ context.Context, _ string) ([]*empregabilidade.CurriculoExperiencia, error) {
	if m.step == 4 {
		return nil, errors.New("experiencias error")
	}
	return []*empregabilidade.CurriculoExperiencia{}, nil
}

func (m *mockCurriculoRepoWithSequentialErrors) CreateConquista(_ context.Context, _ *empregabilidade.CurriculoConquista) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (m *mockCurriculoRepoWithSequentialErrors) GetConquistaByID(_ context.Context, _ uuid.UUID) (*empregabilidade.CurriculoConquista, error) {
	return nil, nil
}

func (m *mockCurriculoRepoWithSequentialErrors) UpdateConquista(_ context.Context, _ *empregabilidade.CurriculoConquista) error {
	return nil
}

func (m *mockCurriculoRepoWithSequentialErrors) DeleteConquista(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockCurriculoRepoWithSequentialErrors) ListConquistasByCPF(_ context.Context, _ string) ([]*empregabilidade.CurriculoConquista, error) {
	if m.step == 5 {
		return nil, errors.New("conquistas error")
	}
	return []*empregabilidade.CurriculoConquista{}, nil
}

func (m *mockCurriculoRepoWithSequentialErrors) ReplaceAllFormacoesByCPF(_ context.Context, _ string, _ []*empregabilidade.CurriculoFormacao) error {
	return nil
}

func (m *mockCurriculoRepoWithSequentialErrors) ReplaceAllFormacaoAccordionByCPF(_ context.Context, _ string, _ []*empregabilidade.CurriculoFormacao, _ []*empregabilidade.CurriculoIdioma) error {
	return nil
}

func (m *mockCurriculoRepoWithSequentialErrors) ReplaceAllExperienciasByCPF(_ context.Context, _ string, _ []*empregabilidade.CurriculoExperiencia) error {
	return nil
}

func (m *mockCurriculoRepoWithSequentialErrors) ReplaceAllExperienciaProfissionalAccordionByCPF(_ context.Context, _ string, _ []*empregabilidade.CurriculoExperiencia, _ []*empregabilidade.CurriculoConquista, _ string) error {
	return nil
}

func (m *mockCurriculoRepoWithSequentialErrors) ReplaceAllConquistasByCPF(_ context.Context, _ string, _ []*empregabilidade.CurriculoConquista) error {
	return nil
}

func (m *mockCurriculoRepoWithSequentialErrors) ReplaceAllIdiomasByCPF(_ context.Context, _ string, _ []*empregabilidade.CurriculoIdioma) error {
	return nil
}

func (m *mockCurriculoRepoWithSequentialErrors) ReplaceAllCursosComplementaresByCPF(_ context.Context, _ string, _ []*empregabilidade.CurriculoCursoComplementar) error {
	return nil
}

func (m *mockCurriculoRepoWithSequentialErrors) UpsertSituacaoInteresses(_ context.Context, _ *empregabilidade.CurriculoSituacaoInteresses) error {
	return nil
}

func (m *mockCurriculoRepoWithSequentialErrors) GetSituacaoInteressesByCPF(_ context.Context, _ string) (*empregabilidade.CurriculoSituacaoInteresses, error) {
	if m.step == 6 {
		return nil, errors.New("situacao error")
	}
	return nil, nil
}

func (m *mockCurriculoRepoWithSequentialErrors) GetPerfilByCPF(_ context.Context, _ string) (*empregabilidade.CurriculoPerfil, error) {
	if m.step == 7 {
		return nil, errors.New("perfil error")
	}
	return nil, nil
}
