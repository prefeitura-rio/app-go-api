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
			// O acesso ao banco é liberado pessoa a pessoa, separado de empregabilidade.
			name:  "admin de empregabilidade sem role do módulo",
			setup: func(c *gin.Context) { c.Set(middlewares.UserRolesKey, []string{"go:empregabilidade:admin"}) },
			want:  http.StatusForbidden,
		},
		{
			name:  "editor de empregabilidade sem role do módulo",
			setup: func(c *gin.Context) { c.Set(middlewares.UserRolesKey, []string{"go:empregabilidade:editor_com_curadoria"}) },
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

func TestBancoCurriculosAssociacaoAuthorization(t *testing.T) {
	cases := []struct {
		name       string
		setup      func(c *gin.Context)
		want       int
		wantOrgaos []string
	}{
		{
			name:  "admin total associa a qualquer órgão",
			setup: func(c *gin.Context) { c.Set(middlewares.UserRoleKey, "ADMIN") },
			want:  http.StatusOK,
		},
		{
			name:  "admin de empregabilidade associa a qualquer órgão",
			setup: func(c *gin.Context) { c.Set(middlewares.UserRolesKey, []string{"go:empregabilidade:admin"}) },
			want:  http.StatusOK,
		},
		{
			name: "editor com secretaria fica restrito a ela",
			setup: func(c *gin.Context) {
				c.Set(middlewares.UserRolesKey, []string{"go:empregabilidade:editor_sem_curadoria"})
				c.Set(middlewares.UserSecretariaOrgaoIDsKey, []string{"2600", "2601"})
			},
			want:       http.StatusOK,
			wantOrgaos: []string{"2600", "2601"},
		},
		{
			name: "editor com órgão fica restrito a ele",
			setup: func(c *gin.Context) {
				c.Set(middlewares.UserRolesKey, []string{"go:empregabilidade:editor_com_curadoria"})
				c.Set(middlewares.UserGroupsKey, []string{"go:orgao:2700"})
			},
			want:       http.StatusOK,
			wantOrgaos: []string{"2700"},
		},
		{
			name:  "editor sem secretaria nem órgão",
			setup: func(c *gin.Context) { c.Set(middlewares.UserRolesKey, []string{"go:empregabilidade:editor_sem_curadoria"}) },
			want:  http.StatusForbidden,
		},
		{
			name:  "perfil só do banco de currículos não associa",
			setup: func(c *gin.Context) { c.Set(middlewares.UserRolesKey, []string{middlewares.RoleCurriculosAdmin}) },
			want:  http.StatusForbidden,
		},
		{
			name:  "cidadão autenticado",
			setup: func(c *gin.Context) { c.Set(middlewares.UserCPFKey, "11111111111") },
			want:  http.StatusForbidden,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.Use(func(c *gin.Context) { tc.setup(c); c.Next() })
			r.Use(middlewares.BancoCurriculosAssociacaoAuthorization())
			var orgaos []string
			r.POST("/test", func(c *gin.Context) {
				orgaos = middlewares.GetBancoCurriculosOrgaosPermitidos(c)
				c.Status(http.StatusOK)
			})

			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/test", nil))

			assert.Equal(t, tc.want, w.Code)
			assert.Equal(t, tc.wantOrgaos, orgaos)
		})
	}
}
