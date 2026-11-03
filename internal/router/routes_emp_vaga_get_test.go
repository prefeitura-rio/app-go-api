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
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	empModels "github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	empServices "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
	"github.com/stretchr/testify/assert"
)

// vagaGetRepoStub atende só o GetByID; os demais métodos da interface não são
// usados pela rota GET /vagas/:id (chamá-los derruba o teste com panic).
type vagaGetRepoStub struct {
	empServices.VagaRepoInterface
	vagas map[uuid.UUID]*empModels.Vaga
}

func (s *vagaGetRepoStub) GetByID(_ context.Context, id uuid.UUID) (*empModels.Vaga, error) {
	v, ok := s.vagas[id]
	if !ok {
		return nil, nil
	}
	copied := *v
	return &copied, nil
}

func injectCaller(role string, roles []string, orgaoID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if role != "" {
			c.Set(middlewares.UserRoleKey, role)
		}
		if roles != nil {
			c.Set(middlewares.UserRolesKey, roles)
		}
		if orgaoID != "" {
			c.Set(middlewares.UserGroupsKey, []string{"go:orgao:" + orgaoID})
		}
		c.Next()
	}
}

// GET /api/v1/empregabilidade/vagas/:id exige vagaAuth e, para vaga ainda não
// publicada, que o órgão parceiro esteja no escopo do chamador (DEV-135).
func TestRegisterEmpregabilidadeRoutes_VagaGetByID_Authorization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rascunhoID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	aprovacaoID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	publicadaID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	inexistenteID := uuid.MustParse("44444444-4444-4444-8444-444444444444")

	repo := &vagaGetRepoStub{vagas: map[uuid.UUID]*empModels.Vaga{
		rascunhoID:  {ID: rascunhoID, Titulo: "Rascunho", IDOrgaoParceiro: "orgao-b", Status: empModels.StatusVagaEmEdicao},
		aprovacaoID: {ID: aprovacaoID, Titulo: "Em aprovação", IDOrgaoParceiro: "orgao-b", Status: empModels.StatusVagaEmAprovacao},
		publicadaID: {ID: publicadaID, Titulo: "Publicada", IDOrgaoParceiro: "orgao-b", Status: empModels.StatusVagaPublicadoAtivo},
	}}
	svc := empServices.NewVagaServiceWithInterfaces(repo, nil, nil)

	editor := []string{"go:empregabilidade:editor"}
	tests := []struct {
		name   string
		caller gin.HandlerFunc
		id     uuid.UUID
		want   int
	}{
		{"cidadão sem role, vaga em edição", injectCaller("", nil, ""), rascunhoID, http.StatusForbidden},
		{"cidadão sem role, vaga em aprovação", injectCaller("", nil, ""), aprovacaoID, http.StatusForbidden},
		{"cidadão sem role, vaga publicada", injectCaller("", nil, ""), publicadaID, http.StatusForbidden},
		{"editor de outro órgão, vaga em edição", injectCaller("", editor, "orgao-a"), rascunhoID, http.StatusForbidden},
		{"editor de outro órgão, vaga em aprovação", injectCaller("", editor, "orgao-a"), aprovacaoID, http.StatusForbidden},
		{"editor de outro órgão, vaga publicada", injectCaller("", editor, "orgao-a"), publicadaID, http.StatusOK},
		{"editor do órgão, vaga em edição", injectCaller("", editor, "orgao-b"), rascunhoID, http.StatusOK},
		{"editor do órgão, vaga em aprovação", injectCaller("", editor, "orgao-b"), aprovacaoID, http.StatusOK},
		{"emp admin, vaga em edição", injectCaller("", []string{"go:empregabilidade:admin"}, ""), rascunhoID, http.StatusOK},
		{"admin, vaga em aprovação", injectCaller("ADMIN", nil, ""), aprovacaoID, http.StatusOK},
		{"editor do órgão, vaga inexistente", injectCaller("", editor, "orgao-b"), inexistenteID, http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(tt.caller)
			app := createMockAppContainer()
			app.EmpVagaService = svc
			app.EmpVagaHandler = empHandlers.NewVagaHandler(svc)
			cfg := &config.AppConfig{App: config.AppSettings{RBACEnabled: true}}
			registerEmpregabilidadeRoutes(router.Group("/api/v1"), router.Group("/api/public"), app, cfg)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/empregabilidade/vagas/"+tt.id.String(), nil))
			assert.Equal(t, tt.want, w.Code, w.Body.String())
		})
	}
}
