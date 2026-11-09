package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/config"
	empHandlers "github.com/prefeitura-rio/app-go-api/internal/handlers/v1/empregabilidade"
	empModels "github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	empRepository "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
	empServices "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
	"github.com/stretchr/testify/assert"
)

// candidaturaRepoStub atende as rotas de gestão de candidaturas e conta as
// chamadas, para provar que um chamador barrado nunca chega ao repositório.
// Não há candidatura cadastrada: as rotas por id respondem 404.
type candidaturaRepoStub struct {
	calls int
}

func (s *candidaturaRepoStub) Create(_ context.Context, _ *empModels.Candidatura) (uuid.UUID, error) {
	s.calls++
	return uuid.New(), nil
}

func (s *candidaturaRepoStub) GetByID(_ context.Context, _ uuid.UUID) (*empModels.Candidatura, error) {
	s.calls++
	return nil, nil
}

func (s *candidaturaRepoStub) Update(_ context.Context, _ *empModels.Candidatura) error {
	s.calls++
	return nil
}

func (s *candidaturaRepoStub) Delete(_ context.Context, _ uuid.UUID) error {
	s.calls++
	return nil
}

func (s *candidaturaRepoStub) List(_ context.Context, _ empModels.CandidaturaFilter, _, _ int) ([]*empModels.Candidatura, int, error) {
	s.calls++
	return []*empModels.Candidatura{}, 0, nil
}

func (s *candidaturaRepoStub) CheckExistingCandidatura(_ context.Context, _ string, _ uuid.UUID) (bool, error) {
	s.calls++
	return false, nil
}

func (s *candidaturaRepoStub) UpdateStatus(_ context.Context, _ uuid.UUID, _ empModels.StatusCandidatura) error {
	s.calls++
	return nil
}

func (s *candidaturaRepoStub) UpdateEtapa(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
	s.calls++
	return nil
}

func (s *candidaturaRepoStub) BulkUpdateStatus(_ context.Context, _ uuid.UUID, _ []string, _ empModels.StatusCandidatura) (empRepository.BulkUpdateResult, error) {
	s.calls++
	return empRepository.BulkUpdateResult{}, nil
}

func (s *candidaturaRepoStub) BulkGetByCPFs(_ context.Context, _ uuid.UUID, _ []string) ([]*empModels.Candidatura, error) {
	s.calls++
	return []*empModels.Candidatura{}, nil
}

func (s *candidaturaRepoStub) BulkUpdateEtapa(_ context.Context, _ []uuid.UUID, _ uuid.UUID) error {
	s.calls++
	return nil
}

func (s *candidaturaRepoStub) BulkSaveAndUpdateStatusByVagaID(_ context.Context, _ uuid.UUID, _ empModels.StatusCandidatura) error {
	s.calls++
	return nil
}

func (s *candidaturaRepoStub) BulkRestoreStatusByVagaID(_ context.Context, _ uuid.UUID) error {
	s.calls++
	return nil
}

func (s *candidaturaRepoStub) CountByStatus(_ context.Context, _ empModels.CandidaturaFilter) (map[empModels.StatusCandidatura]int64, error) {
	s.calls++
	return map[empModels.StatusCandidatura]int64{}, nil
}

// candidaturaVagaRepoStub não tem vaga cadastrada e soma as chamadas no mesmo
// contador do repositório de candidaturas.
type candidaturaVagaRepoStub struct {
	repo *candidaturaRepoStub
}

func (s candidaturaVagaRepoStub) GetByID(_ context.Context, _ uuid.UUID) (*empModels.Vaga, error) {
	s.repo.calls++
	return nil, nil
}

// As rotas de gestão de candidaturas (listagem com CPF do candidato e mudança
// de status e etapa) só aceitam ADMIN, go:empregabilidade:admin ou editor de
// Empregabilidade, mesmo com RBAC_ENABLED=false (DEV-435). Com o RBAC ligado,
// o editor também precisa de escopo de órgão/secretaria (VagaAuthorization).
func TestRegisterEmpregabilidadeRoutes_CandidaturasGestao_Authorization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const base = "/api/v1/empregabilidade/candidaturas"
	id := "55555555-5555-4555-8555-555555555555"
	vagaID := "66666666-6666-4666-8666-666666666666"
	etapaID := "77777777-7777-4777-8777-777777777777"

	routes := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodGet, base, ""},
		{http.MethodPut, base + "/bulk-status", `{"cpfs":["00000000000"],"vaga_id":"` + vagaID + `","status":"aprovada"}`},
		{http.MethodPut, base + "/bulk-etapa", `{"cpfs":["00000000000"],"vaga_id":"` + vagaID + `","id_etapa":"` + etapaID + `"}`},
		{http.MethodPut, base + "/" + id + "/status", `{"status":"aprovada"}`},
		{http.MethodPut, base + "/" + id + "/approve", ""},
		{http.MethodPut, base + "/" + id + "/reject", ""},
		{http.MethodPut, base + "/" + id + "/etapa", `{"id_etapa_atual":"` + etapaID + `"}`},
	}

	callers := []struct {
		name        string
		rbacEnabled bool
		role        string
		roles       []string
		orgaoID     string
		allowed     bool
	}{
		{"sem role, RBAC desligado", false, "USER", nil, "", false},
		{"sem role, RBAC ligado", true, "USER", nil, "", false},
		{"editor de cursos com órgão, RBAC desligado", false, "USER", []string{"go:cursos:editor"}, "orgao-1", false},
		{"editor de cursos com órgão, RBAC ligado", true, "USER", []string{"go:cursos:editor"}, "orgao-1", false},
		{"admin de cursos, RBAC desligado", false, "USER", []string{"go:cursos:admin"}, "", false},
		{"admin de serviços, RBAC desligado", false, "USER", []string{"busca:services:admin"}, "", false},
		{"admin do banco de currículos, RBAC desligado", false, "USER", []string{"go:curriculos:admin"}, "", false},
		{"role desconhecida de empregabilidade, RBAC desligado", false, "USER", []string{"go:empregabilidade:leitor"}, "orgao-1", false},
		{"ADMIN, RBAC desligado", false, "ADMIN", []string{"go:admin"}, "", true},
		{"ADMIN, RBAC ligado", true, "ADMIN", []string{"go:admin"}, "", true},
		{"admin de empregabilidade, RBAC desligado", false, "USER", []string{"go:empregabilidade:admin"}, "", true},
		{"admin de empregabilidade, RBAC ligado", true, "USER", []string{"go:empregabilidade:admin"}, "", true},
		{"editor de empregabilidade sem órgão, RBAC desligado", false, "USER", []string{"go:empregabilidade:editor"}, "", true},
		{"editor sem curadoria sem órgão, RBAC desligado", false, "USER", []string{"go:empregabilidade:editor_sem_curadoria"}, "", true},
		{"editor com curadoria sem órgão, RBAC desligado", false, "USER", []string{"go:empregabilidade:editor_com_curadoria"}, "", true},
		{"editor de empregabilidade sem órgão, RBAC ligado", true, "USER", []string{"go:empregabilidade:editor"}, "", false},
		{"editor com curadoria com órgão, RBAC ligado", true, "USER", []string{"go:empregabilidade:editor_com_curadoria"}, "orgao-1", true},
	}

	for _, rt := range routes {
		for _, tt := range callers {
			t.Run(rt.method+" "+strings.TrimPrefix(rt.path, base)+" "+tt.name, func(t *testing.T) {
				repo := &candidaturaRepoStub{}
				app := createMockAppContainer()
				app.EmpCandidaturaHandler = empHandlers.NewCandidaturaHandler(
					empServices.NewCandidaturaService(repo, candidaturaVagaRepoStub{repo: repo}, nil, nil, nil, nil),
				)

				router := gin.New()
				apiV1 := router.Group("/api/v1", injectCaller(tt.role, tt.roles, tt.orgaoID))
				cfg := &config.AppConfig{App: config.AppSettings{RBACEnabled: tt.rbacEnabled}}
				registerEmpregabilidadeRoutes(apiV1, router.Group("/api/public"), app, cfg)

				req := httptest.NewRequest(rt.method, rt.path, strings.NewReader(rt.body))
				if rt.body != "" {
					req.Header.Set("Content-Type", "application/json")
				}
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				if tt.allowed {
					assert.NotEqual(t, http.StatusForbidden, w.Code, w.Body.String())
					assert.Positive(t, repo.calls, "chamador autorizado tem de chegar ao repositório")
				} else {
					assert.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
					assert.Equal(t, 0, repo.calls, "chamador barrado não pode chegar ao repositório")
				}
			})
		}
	}
}
