package empregabilidade_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	handlers "github.com/prefeitura-rio/app-go-api/internal/handlers/v1/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	empmodels "github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
)

// ──────────────────────────────────────────────────────────────────────────────
// Mock Onboarding repository
// ──────────────────────────────────────────────────────────────────────────────

type mockOnboardingRepoH struct {
	entity          *empmodels.Onboarding
	isFirstLogin    bool
	err             error
	upsertErr       error
	markCompleteErr error
}

func (m *mockOnboardingRepoH) GetByCPF(_ context.Context, _ string) (*empmodels.Onboarding, error) {
	return m.entity, m.err
}
func (m *mockOnboardingRepoH) Upsert(_ context.Context, _ *empmodels.Onboarding) error {
	return m.upsertErr
}
func (m *mockOnboardingRepoH) MarkFirstLoginCompleted(_ context.Context, _ string) error {
	return m.markCompleteErr
}

func setupOnboardingRouter(repo services.OnboardingRepositoryInterface, cpf string, isAdmin bool) *gin.Engine {
	r := gin.New()
	// Inject user context via middleware
	r.Use(func(c *gin.Context) {
		if cpf != "" {
			c.Set(middlewares.UserCPFKey, cpf)
		}
		if isAdmin {
			c.Set(middlewares.UserRoleKey, "ADMIN")
		}
		c.Next()
	})
	svc := services.NewOnboardingServiceWithInterface(repo)
	h := handlers.NewOnboardingHandler(svc)
	r.GET("/onboarding/:cpf", h.IsFirstLogin)
	r.PUT("/onboarding/:cpf/complete", h.MarkFirstLoginCompleted)
	return r
}

func TestOnboardingHandler_IsFirstLogin_AsAdmin(t *testing.T) {
	entity := &empmodels.Onboarding{CPF: "12345678900"}
	repo := &mockOnboardingRepoH{entity: entity}
	r := setupOnboardingRouter(repo, "12345678900", true)
	req := httptest.NewRequest(http.MethodGet, "/onboarding/12345678900", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestOnboardingHandler_IsFirstLogin_AsOwner(t *testing.T) {
	entity := &empmodels.Onboarding{CPF: "12345678900", IsEmpregabilidadeFirstLogin: false}
	repo := &mockOnboardingRepoH{entity: entity}
	r := setupOnboardingRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/onboarding/12345678900", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestOnboardingHandler_IsFirstLogin_Forbidden(t *testing.T) {
	repo := &mockOnboardingRepoH{}
	r := setupOnboardingRouter(repo, "99999999999", false) // different CPF
	req := httptest.NewRequest(http.MethodGet, "/onboarding/12345678900", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestOnboardingHandler_IsFirstLogin_Unauthorized(t *testing.T) {
	repo := &mockOnboardingRepoH{}
	r := setupOnboardingRouter(repo, "", false) // no CPF in context
	req := httptest.NewRequest(http.MethodGet, "/onboarding/12345678900", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestOnboardingHandler_IsFirstLogin_ServiceError(t *testing.T) {
	repo := &mockOnboardingRepoH{err: fmt.Errorf("db error")}
	r := setupOnboardingRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/onboarding/12345678900", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestOnboardingHandler_MarkCompleted_AsOwner(t *testing.T) {
	repo := &mockOnboardingRepoH{}
	r := setupOnboardingRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodPut, "/onboarding/12345678900/complete", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestOnboardingHandler_MarkCompleted_Forbidden(t *testing.T) {
	repo := &mockOnboardingRepoH{}
	r := setupOnboardingRouter(repo, "99999999999", false)
	req := httptest.NewRequest(http.MethodPut, "/onboarding/12345678900/complete", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestOnboardingHandler_MarkCompleted_Unauthorized(t *testing.T) {
	repo := &mockOnboardingRepoH{}
	r := setupOnboardingRouter(repo, "", false)
	req := httptest.NewRequest(http.MethodPut, "/onboarding/12345678900/complete", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestOnboardingHandler_MarkCompleted_ServiceError(t *testing.T) {
	repo := &mockOnboardingRepoH{markCompleteErr: fmt.Errorf("db error")}
	r := setupOnboardingRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodPut, "/onboarding/12345678900/complete", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Mock TermosUso repository
// ──────────────────────────────────────────────────────────────────────────────

type mockTermosUsoRepoH struct {
	entity      *empmodels.TermosUso
	hasAccepted bool
	err         error
	upsertErr   error
	acceptErr   error
}

func (m *mockTermosUsoRepoH) GetByCPF(_ context.Context, _ string) (*empmodels.TermosUso, error) {
	return m.entity, m.err
}
func (m *mockTermosUsoRepoH) Upsert(_ context.Context, _ *empmodels.TermosUso) error {
	return m.upsertErr
}
func (m *mockTermosUsoRepoH) AcceptTerms(_ context.Context, _ string) error {
	return m.acceptErr
}

func setupTermosUsoRouter(repo services.TermosUsoRepositoryInterface, cpf string, isAdmin bool) *gin.Engine {
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
	svc := services.NewTermosUsoServiceWithInterface(repo)
	h := handlers.NewTermosUsoHandler(svc)
	r.GET("/termos/:cpf", h.HasAcceptedTerms)
	r.PUT("/termos/:cpf/accept", h.AcceptTerms)
	r.GET("/termos/:cpf/details", h.GetDetails)
	return r
}

func TestTermosUsoHandler_HasAcceptedTerms_AsOwner(t *testing.T) {
	repo := &mockTermosUsoRepoH{entity: &empmodels.TermosUso{CPF: "12345678900"}}
	r := setupTermosUsoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/termos/12345678900", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTermosUsoHandler_HasAcceptedTerms_AsAdmin(t *testing.T) {
	repo := &mockTermosUsoRepoH{entity: &empmodels.TermosUso{CPF: "12345678900"}}
	r := setupTermosUsoRouter(repo, "admin", true)
	req := httptest.NewRequest(http.MethodGet, "/termos/12345678900", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestTermosUsoHandler_HasAcceptedTerms_Forbidden(t *testing.T) {
	repo := &mockTermosUsoRepoH{}
	r := setupTermosUsoRouter(repo, "99999999999", false)
	req := httptest.NewRequest(http.MethodGet, "/termos/12345678900", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestTermosUsoHandler_HasAcceptedTerms_Unauthorized(t *testing.T) {
	repo := &mockTermosUsoRepoH{}
	r := setupTermosUsoRouter(repo, "", false)
	req := httptest.NewRequest(http.MethodGet, "/termos/12345678900", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestTermosUsoHandler_HasAcceptedTerms_ServiceError(t *testing.T) {
	repo := &mockTermosUsoRepoH{err: fmt.Errorf("db error")}
	r := setupTermosUsoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/termos/12345678900", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestTermosUsoHandler_AcceptTerms_Success(t *testing.T) {
	repo := &mockTermosUsoRepoH{}
	r := setupTermosUsoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodPut, "/termos/12345678900/accept", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTermosUsoHandler_AcceptTerms_Forbidden(t *testing.T) {
	repo := &mockTermosUsoRepoH{}
	r := setupTermosUsoRouter(repo, "99999999999", false)
	req := httptest.NewRequest(http.MethodPut, "/termos/12345678900/accept", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestTermosUsoHandler_AcceptTerms_ServiceError(t *testing.T) {
	repo := &mockTermosUsoRepoH{acceptErr: fmt.Errorf("db error")}
	r := setupTermosUsoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodPut, "/termos/12345678900/accept", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestTermosUsoHandler_GetDetails_Found(t *testing.T) {
	entity := &empmodels.TermosUso{CPF: "12345678900"}
	repo := &mockTermosUsoRepoH{entity: entity}
	r := setupTermosUsoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/termos/12345678900/details", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTermosUsoHandler_GetDetails_NotFound(t *testing.T) {
	repo := &mockTermosUsoRepoH{entity: nil}
	r := setupTermosUsoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/termos/12345678900/details", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestTermosUsoHandler_GetDetails_ServiceError(t *testing.T) {
	repo := &mockTermosUsoRepoH{err: fmt.Errorf("db error")}
	r := setupTermosUsoRouter(repo, "12345678900", false)
	req := httptest.NewRequest(http.MethodGet, "/termos/12345678900/details", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// DEV-38: erros internos não vazam no corpo do 500
// ──────────────────────────────────────────────────────────────────────────────

// driverErr imita o texto que o GORM/pq devolve numa falha de banco: nome de
// tabela, constraint e o valor da chave em DETAIL. Nada disso pode chegar ao cliente.
const driverErr = `ERROR: duplicate key value violates unique constraint "uni_termos_uso_cpf" (SQLSTATE 23505) DETAIL: Key (cpf)=(00000000000) already exists. relation "empregabilidade.termos_uso"`

func TestOnboardingTermosHandlers_ServiceError_DoesNotLeakInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const cpf = "12345678900"
	dbErr := fmt.Errorf("erro ao buscar: %w", fmt.Errorf("%s", driverErr))

	tests := []struct {
		name        string
		router      *gin.Engine
		method      string
		path        string
		wantMessage string
	}{
		{
			name:        "termos HasAcceptedTerms",
			router:      setupTermosUsoRouter(&mockTermosUsoRepoH{err: dbErr}, cpf, false),
			method:      http.MethodGet,
			path:        "/termos/" + cpf,
			wantMessage: "Erro ao verificar aceite dos termos de uso",
		},
		{
			name:        "termos AcceptTerms",
			router:      setupTermosUsoRouter(&mockTermosUsoRepoH{acceptErr: dbErr}, cpf, false),
			method:      http.MethodPut,
			path:        "/termos/" + cpf + "/accept",
			wantMessage: "Erro ao registrar aceite dos termos de uso",
		},
		{
			name:        "termos GetDetails",
			router:      setupTermosUsoRouter(&mockTermosUsoRepoH{err: dbErr}, cpf, false),
			method:      http.MethodGet,
			path:        "/termos/" + cpf + "/details",
			wantMessage: "Erro ao buscar termos de uso",
		},
		{
			name:        "onboarding IsFirstLogin",
			router:      setupOnboardingRouter(&mockOnboardingRepoH{err: dbErr}, cpf, false),
			method:      http.MethodGet,
			path:        "/onboarding/" + cpf,
			wantMessage: "Erro ao verificar onboarding",
		},
		{
			name:        "onboarding MarkFirstLoginCompleted",
			router:      setupOnboardingRouter(&mockOnboardingRepoH{markCompleteErr: dbErr}, cpf, false),
			method:      http.MethodPut,
			path:        "/onboarding/" + cpf + "/complete",
			wantMessage: "Erro ao concluir onboarding",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			tt.router.ServeHTTP(w, req)

			if w.Code != http.StatusInternalServerError {
				t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
			}
			var body map[string]string
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("invalid JSON body %q: %v", w.Body.String(), err)
			}
			if body["error"] != tt.wantMessage {
				t.Errorf("expected error %q, got %q", tt.wantMessage, body["error"])
			}
			for _, leaked := range []string{"SQLSTATE", "constraint", "DETAIL", "termos_uso", "00000000000", "erro ao buscar"} {
				if strings.Contains(w.Body.String(), leaked) {
					t.Errorf("response body leaks internal error detail %q: %s", leaked, w.Body.String())
				}
			}
		})
	}
}
