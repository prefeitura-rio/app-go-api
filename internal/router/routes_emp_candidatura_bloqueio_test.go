package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/config"
	empHandlers "github.com/prefeitura-rio/app-go-api/internal/handlers/v1/empregabilidade"
	empModels "github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	empServices "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
	"github.com/stretchr/testify/assert"
)

// bloqueioRepoStub atende o List da rota GET /candidatura-bloqueios e conta as
// chamadas, para provar que um chamador barrado nunca chega ao repositório.
type bloqueioRepoStub struct {
	listCalls int
}

func (s *bloqueioRepoStub) Create(_ context.Context, _ *empModels.CandidaturaBloqueio) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (s *bloqueioRepoStub) List(_ context.Context, _ map[string]interface{}, _, _ int) ([]*empModels.CandidaturaBloqueio, int, error) {
	s.listCalls++
	return []*empModels.CandidaturaBloqueio{}, 0, nil
}

type vagaExistenciaStub struct{}

func (vagaExistenciaStub) ExistsByID(_ context.Context, _ uuid.UUID) (bool, error) { return true, nil }

// GET /api/v1/empregabilidade/candidatura-bloqueios devolve CPFs de cidadãos:
// só ADMIN, go:empregabilidade:admin ou editor de Empregabilidade podem listar,
// mesmo com RBAC_ENABLED=false (DEV-434). Com o RBAC ligado, o editor também
// precisa de escopo de órgão/secretaria (VagaAuthorization).
func TestRegisterEmpregabilidadeRoutes_CandidaturaBloqueiosList_Authorization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const path = "/api/v1/empregabilidade/candidatura-bloqueios"

	tests := []struct {
		name        string
		rbacEnabled bool
		role        string
		roles       []string
		orgaoID     string
		wantStatus  int
	}{
		{"sem role, RBAC desligado", false, "USER", nil, "", http.StatusForbidden},
		{"sem role, RBAC ligado", true, "USER", nil, "", http.StatusForbidden},
		{"editor de cursos com órgão, RBAC desligado", false, "USER", []string{"go:cursos:editor"}, "orgao-1", http.StatusForbidden},
		{"editor de cursos com órgão, RBAC ligado", true, "USER", []string{"go:cursos:editor"}, "orgao-1", http.StatusForbidden},
		{"admin de cursos, RBAC desligado", false, "USER", []string{"go:cursos:admin"}, "", http.StatusForbidden},
		{"admin de serviços, RBAC desligado", false, "USER", []string{"go:servicos:admin"}, "", http.StatusForbidden},
		{"admin do banco de currículos, RBAC desligado", false, "USER", []string{"go:curriculos:admin"}, "", http.StatusForbidden},
		{"ADMIN, RBAC desligado", false, "ADMIN", []string{"go:admin"}, "", http.StatusOK},
		{"ADMIN, RBAC ligado", true, "ADMIN", []string{"go:admin"}, "", http.StatusOK},
		{"admin de empregabilidade, RBAC desligado", false, "USER", []string{"go:empregabilidade:admin"}, "", http.StatusOK},
		{"admin de empregabilidade, RBAC ligado", true, "USER", []string{"go:empregabilidade:admin"}, "", http.StatusOK},
		{"editor de empregabilidade sem órgão, RBAC desligado", false, "USER", []string{"go:empregabilidade:editor"}, "", http.StatusOK},
		{"editor sem curadoria sem órgão, RBAC desligado", false, "USER", []string{"go:empregabilidade:editor_sem_curadoria"}, "", http.StatusOK},
		{"editor com curadoria sem órgão, RBAC desligado", false, "USER", []string{"go:empregabilidade:editor_com_curadoria"}, "", http.StatusOK},
		{"editor de empregabilidade sem órgão, RBAC ligado", true, "USER", []string{"go:empregabilidade:editor"}, "", http.StatusForbidden},
		{"editor de empregabilidade com órgão, RBAC ligado", true, "USER", []string{"go:empregabilidade:editor"}, "orgao-1", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &bloqueioRepoStub{}
			app := createMockAppContainer()
			app.EmpCandidaturaBloqueioHandler = empHandlers.NewCandidaturaBloqueioHandler(
				empServices.NewCandidaturaBloqueioServiceWithInterface(repo, vagaExistenciaStub{}),
			)

			router := gin.New()
			apiV1 := router.Group("/api/v1", injectCaller(tt.role, tt.roles, tt.orgaoID))
			cfg := &config.AppConfig{App: config.AppSettings{RBACEnabled: tt.rbacEnabled}}
			registerEmpregabilidadeRoutes(apiV1, router.Group("/api/public"), app, cfg)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus == http.StatusForbidden {
				assert.Equal(t, 0, repo.listCalls, "chamador barrado não pode chegar ao repositório")
			} else {
				assert.Equal(t, 1, repo.listCalls)
			}
		})
	}
}
