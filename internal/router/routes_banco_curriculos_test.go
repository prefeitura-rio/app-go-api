package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prefeitura-rio/app-go-api/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestRegisterEmpregabilidadeRoutes_BancoCurriculos(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// A autorização do banco de currículos vale mesmo com RBAC_ENABLED=false:
	// sem ela, qualquer usuário autenticado listaria todos os currículos.
	for _, rbacEnabled := range []bool{true, false} {
		router := gin.New()
		cfg := &config.AppConfig{App: config.AppSettings{RBACEnabled: rbacEnabled}}
		registerEmpregabilidadeRoutes(router.Group("/api/v1"), router.Group("/api/public"), createMockAppContainer(), cfg)

		routes := router.Routes()
		assert.True(t, routeExists(routes, "GET", "/api/v1/empregabilidade/banco-curriculos"))
		assert.True(t, routeExists(routes, "GET", "/api/v1/empregabilidade/banco-curriculos/:cpf"))

		for _, path := range []string{
			"/api/v1/empregabilidade/banco-curriculos",
			"/api/v1/empregabilidade/banco-curriculos/11111111111",
		} {
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
			assert.Equal(t, http.StatusForbidden, w.Code, "%s sem role com RBAC_ENABLED=%v", path, rbacEnabled)
		}
	}
}
