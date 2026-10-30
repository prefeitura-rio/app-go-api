package empregabilidade_test

import (
	"context"
	"encoding/json"
	"fmt"
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

type mockCandidaturaBloqueioRepoH struct {
	listItems []*empmodels.CandidaturaBloqueio
	listTotal int
	err       error
	created   *empmodels.CandidaturaBloqueio
}

func (m *mockCandidaturaBloqueioRepoH) Create(_ context.Context, entity *empmodels.CandidaturaBloqueio) (uuid.UUID, error) {
	if m.err != nil {
		return uuid.Nil, m.err
	}
	m.created = entity
	return uuid.New(), nil
}

func (m *mockCandidaturaBloqueioRepoH) List(_ context.Context, _ map[string]interface{}, _, _ int) ([]*empmodels.CandidaturaBloqueio, int, error) {
	return m.listItems, m.listTotal, m.err
}

// Usuário cidadão do contexto padrão dos testes (CPF sintético).
const bloqueioTokenCPF = "12345678900"

func vagaExistenteBloqueio() *mockVagaForCandidaturaH {
	return &mockVagaForCandidaturaH{entity: &empmodels.Vaga{
		ID:     uuid.MustParse(validUUID),
		Status: empmodels.StatusVagaPublicadoAtivo,
	}}
}

// setupCandidaturaBloqueioRouterAs monta o router com o contexto de usuário que
// o middleware ExtractUserContext injetaria: CPF do token e papéis.
func setupCandidaturaBloqueioRouterAs(
	repo services.CandidaturaBloqueioRepositoryInterface,
	vagaRepo services.VagaRepositoryInterface,
	cpf, role string,
	roles []string,
) *gin.Engine {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if cpf != "" {
			c.Set(middlewares.UserCPFKey, cpf)
		}
		if role != "" {
			c.Set(middlewares.UserRoleKey, role)
		}
		if len(roles) > 0 {
			c.Set(middlewares.UserRolesKey, roles)
		}
		c.Next()
	})
	svc := services.NewCandidaturaBloqueioServiceWithInterface(repo, vagaRepo)
	h := handlers.NewCandidaturaBloqueioHandler(svc)
	r.POST("/candidatura-bloqueios", h.Create)
	r.GET("/candidatura-bloqueios", h.List)
	return r
}

func setupCandidaturaBloqueioRouter(repo services.CandidaturaBloqueioRepositoryInterface) *gin.Engine {
	return setupCandidaturaBloqueioRouterAs(repo, vagaExistenteBloqueio(), bloqueioTokenCPF, "", nil)
}

func postBloqueio(r *gin.Engine, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/candidatura-bloqueios", bodyOf(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestCandidaturaBloqueioHandler_Create_Success(t *testing.T) {
	repo := &mockCandidaturaBloqueioRepoH{}
	r := setupCandidaturaBloqueioRouter(repo)
	w := postBloqueio(r, `{"cpf":"12345678900","id_vaga":"`+validUUID+`","criterios_nao_atendidos":["idade_minima"]}`)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCandidaturaBloqueioHandler_Create_CidadaoComCPFAlheio_GravaCPFDoToken(t *testing.T) {
	repo := &mockCandidaturaBloqueioRepoH{}
	r := setupCandidaturaBloqueioRouter(repo)
	w := postBloqueio(r, `{"cpf":"98765432100","id_vaga":"`+validUUID+`","criterios_nao_atendidos":["idade_minima"]}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	if repo.created == nil || repo.created.CPF != bloqueioTokenCPF {
		t.Fatalf("expected persisted CPF from token %q, got %+v", bloqueioTokenCPF, repo.created)
	}
	var resp empmodels.CandidaturaBloqueio
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if resp.CPF != bloqueioTokenCPF {
		t.Errorf("expected response CPF %q, got %q", bloqueioTokenCPF, resp.CPF)
	}
}

func TestCandidaturaBloqueioHandler_Create_CidadaoSemCPFNoCorpo_UsaToken(t *testing.T) {
	repo := &mockCandidaturaBloqueioRepoH{}
	r := setupCandidaturaBloqueioRouter(repo)
	w := postBloqueio(r, `{"id_vaga":"`+validUUID+`"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	if repo.created == nil || repo.created.CPF != bloqueioTokenCPF {
		t.Fatalf("expected persisted CPF from token %q, got %+v", bloqueioTokenCPF, repo.created)
	}
}

func TestCandidaturaBloqueioHandler_Create_SemCPFNoToken_Unauthorized(t *testing.T) {
	repo := &mockCandidaturaBloqueioRepoH{}
	r := setupCandidaturaBloqueioRouterAs(repo, vagaExistenteBloqueio(), "", "", nil)
	w := postBloqueio(r, `{"cpf":"98765432100","id_vaga":"`+validUUID+`"}`)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
	if repo.created != nil {
		t.Errorf("expected nothing persisted, got %+v", repo.created)
	}
}

func TestCandidaturaBloqueioHandler_Create_PapelNaoAdmin_NaoUsaCPFDoCorpo(t *testing.T) {
	// Papéis de outros módulos (ex.: editor de cursos) não autorizam registrar
	// em nome de outro CPF.
	repo := &mockCandidaturaBloqueioRepoH{}
	r := setupCandidaturaBloqueioRouterAs(repo, vagaExistenteBloqueio(), bloqueioTokenCPF, "EDITOR", []string{"go:cursos:editor", "go:empregabilidade:editor"})
	w := postBloqueio(r, `{"cpf":"98765432100","id_vaga":"`+validUUID+`"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	if repo.created == nil || repo.created.CPF != bloqueioTokenCPF {
		t.Fatalf("expected persisted CPF from token %q, got %+v", bloqueioTokenCPF, repo.created)
	}
}

func TestCandidaturaBloqueioHandler_Create_Admin_UsaCPFDoCorpo(t *testing.T) {
	tests := []struct {
		name  string
		role  string
		roles []string
	}{
		{"ADMIN", "ADMIN", nil},
		{"go:empregabilidade:admin", "", []string{"go:empregabilidade:admin"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCandidaturaBloqueioRepoH{}
			r := setupCandidaturaBloqueioRouterAs(repo, vagaExistenteBloqueio(), "11122233344", tt.role, tt.roles)
			w := postBloqueio(r, `{"cpf":"98765432100","id_vaga":"`+validUUID+`"}`)
			if w.Code != http.StatusCreated {
				t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
			}
			if repo.created == nil || repo.created.CPF != "98765432100" {
				t.Fatalf("expected persisted CPF from body, got %+v", repo.created)
			}
		})
	}
}

func TestCandidaturaBloqueioHandler_Create_AdminSemCPFNoCorpo_UsaToken(t *testing.T) {
	repo := &mockCandidaturaBloqueioRepoH{}
	r := setupCandidaturaBloqueioRouterAs(repo, vagaExistenteBloqueio(), "11122233344", "ADMIN", nil)
	w := postBloqueio(r, `{"id_vaga":"`+validUUID+`"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	if repo.created == nil || repo.created.CPF != "11122233344" {
		t.Fatalf("expected persisted CPF from token, got %+v", repo.created)
	}
}

func TestCandidaturaBloqueioHandler_Create_VagaInexistente_NotFound(t *testing.T) {
	repo := &mockCandidaturaBloqueioRepoH{}
	r := setupCandidaturaBloqueioRouterAs(repo, &mockVagaForCandidaturaH{}, bloqueioTokenCPF, "", nil)
	w := postBloqueio(r, `{"id_vaga":"`+validUUID+`"}`)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
	if repo.created != nil {
		t.Errorf("expected nothing persisted, got %+v", repo.created)
	}
}

func TestCandidaturaBloqueioHandler_Create_SemIDVaga_NotFound(t *testing.T) {
	repo := &mockCandidaturaBloqueioRepoH{}
	r := setupCandidaturaBloqueioRouterAs(repo, &mockVagaForCandidaturaH{}, bloqueioTokenCPF, "", nil)
	w := postBloqueio(r, `{"criterios_nao_atendidos":["idade_minima"]}`)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
	if repo.created != nil {
		t.Errorf("expected nothing persisted, got %+v", repo.created)
	}
}

func TestCandidaturaBloqueioHandler_Create_ErroAoBuscarVaga_NaoGrava(t *testing.T) {
	repo := &mockCandidaturaBloqueioRepoH{}
	r := setupCandidaturaBloqueioRouterAs(repo, &mockVagaForCandidaturaH{err: fmt.Errorf("db down")}, bloqueioTokenCPF, "", nil)
	w := postBloqueio(r, `{"id_vaga":"`+validUUID+`"}`)
	if w.Code < http.StatusInternalServerError {
		t.Fatalf("expected 5xx, got %d: %s", w.Code, w.Body.String())
	}
	if repo.created != nil {
		t.Errorf("expected nothing persisted, got %+v", repo.created)
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
