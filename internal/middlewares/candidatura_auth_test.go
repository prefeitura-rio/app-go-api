package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	"github.com/stretchr/testify/assert"
)

func TestCandidaturaGestaoAuthorization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		role       string
		roles      []string
		wantStatus int
	}{
		{"sem contexto de usuário", "", nil, http.StatusForbidden},
		{"cidadão sem role", "USER", nil, http.StatusForbidden},
		{"editor de cursos", "USER", []string{"go:cursos:editor"}, http.StatusForbidden},
		{"admin de cursos", "USER", []string{"go:cursos:admin"}, http.StatusForbidden},
		{"editor de serviços", "USER", []string{"busca:services:editor"}, http.StatusForbidden},
		{"admin do banco de currículos", "USER", []string{"go:curriculos:admin"}, http.StatusForbidden},
		{"editor do banco de currículos", "USER", []string{"go:curriculos:editor"}, http.StatusForbidden},
		{"role desconhecida de empregabilidade", "USER", []string{"go:empregabilidade:leitor"}, http.StatusForbidden},
		{"role parecida, mas não exata", "USER", []string{"go:empregabilidade:editor:x"}, http.StatusForbidden},
		{"ADMIN", "ADMIN", []string{"go:admin"}, http.StatusOK},
		{"admin de empregabilidade", "USER", []string{"go:empregabilidade:admin"}, http.StatusOK},
		{"editor de empregabilidade", "USER", []string{"go:empregabilidade:editor"}, http.StatusOK},
		{"editor sem curadoria", "USER", []string{"go:empregabilidade:editor_sem_curadoria"}, http.StatusOK},
		{"editor com curadoria", "USER", []string{"go:empregabilidade:editor_com_curadoria"}, http.StatusOK},
		{"editor de cursos e de empregabilidade", "USER", []string{"go:cursos:editor", "go:empregabilidade:editor"}, http.StatusOK},
	}

	for _, tt := range tests {
		for _, method := range []string{http.MethodGet, http.MethodPut} {
			t.Run(tt.name+" "+method, func(t *testing.T) {
				r := gin.New()
				r.Use(func(c *gin.Context) {
					if tt.role != "" {
						c.Set(middlewares.UserRoleKey, tt.role)
					}
					if tt.roles != nil {
						c.Set(middlewares.UserRolesKey, tt.roles)
					}
					c.Next()
				})
				reached := false
				r.Handle(method, "/test", middlewares.CandidaturaGestaoAuthorization(), func(c *gin.Context) {
					reached = true
					c.Status(http.StatusOK)
				})

				w := httptest.NewRecorder()
				r.ServeHTTP(w, httptest.NewRequest(method, "/test", nil))

				assert.Equal(t, tt.wantStatus, w.Code)
				assert.Equal(t, tt.wantStatus == http.StatusOK, reached)
			})
		}
	}
}
