package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	"github.com/stretchr/testify/assert"
)

func TestBancoCurriculosAuthorization(t *testing.T) {
	cases := []struct {
		name  string
		setup func(c *gin.Context)
		want  int
	}{
		{
			name:  "admin total",
			setup: func(c *gin.Context) { c.Set(middlewares.UserRoleKey, "ADMIN") },
			want:  http.StatusOK,
		},
		{
			name:  "admin do módulo",
			setup: func(c *gin.Context) { c.Set(middlewares.UserRolesKey, []string{middlewares.RoleCurriculosAdmin}) },
			want:  http.StatusOK,
		},
		{
			name:  "editor do módulo",
			setup: func(c *gin.Context) { c.Set(middlewares.UserRolesKey, []string{middlewares.RoleCurriculosEditor}) },
			want:  http.StatusOK,
		},
		{
			name:  "admin de empregabilidade sem role do módulo",
			setup: func(c *gin.Context) { c.Set(middlewares.UserRolesKey, []string{"go:empregabilidade:admin"}) },
			want:  http.StatusForbidden,
		},
		{
			name:  "cidadão autenticado",
			setup: func(c *gin.Context) { c.Set(middlewares.UserCPFKey, "11111111111") },
			want:  http.StatusForbidden,
		},
		{
			name:  "sem contexto de usuário",
			setup: func(c *gin.Context) {},
			want:  http.StatusForbidden,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.Use(func(c *gin.Context) { tc.setup(c); c.Next() })
			r.Use(middlewares.BancoCurriculosAuthorization())
			r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))

			assert.Equal(t, tc.want, w.Code)
		})
	}
}
