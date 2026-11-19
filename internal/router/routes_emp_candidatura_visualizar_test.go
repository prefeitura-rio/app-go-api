package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prefeitura-rio/app-go-api/internal/config"
	empHandlers "github.com/prefeitura-rio/app-go-api/internal/handlers/v1/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	empServices "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
	"github.com/stretchr/testify/assert"
)

// Marcar as atualizações como vistas é do próprio candidato: a rota não passa
// pela autorização de gestão de candidaturas, e o cidadão sem role chega ao
// repositório com o próprio CPF.
func TestRegisterEmpregabilidadeRoutes_CandidaturasVisualizar_Cidadao(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &candidaturaRepoStub{}
	app := createMockAppContainer()
	app.EmpCandidaturaHandler = empHandlers.NewCandidaturaHandler(
		empServices.NewCandidaturaService(repo, candidaturaVagaRepoStub{repo: repo}, nil, nil, nil, nil),
	)

	router := gin.New()
	apiV1 := router.Group("/api/v1", injectCaller("USER", nil, ""), func(c *gin.Context) {
		c.Set(middlewares.UserCPFKey, "12345678900")
		c.Next()
	})
	registerEmpregabilidadeRoutes(apiV1, router.Group("/api/public"), app, &config.AppConfig{})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/v1/empregabilidade/candidaturas/usuario/12345678900/visualizar", nil))

	assert.Equal(t, http.StatusNoContent, w.Code, w.Body.String())
	assert.Equal(t, 1, repo.calls)
}
