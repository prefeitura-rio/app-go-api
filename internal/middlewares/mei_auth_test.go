package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	"github.com/stretchr/testify/assert"
)

var meiGestaoCases = []struct {
	name  string
	role  string
	roles []string
	allow bool
}{
	{"sem contexto de usuário", "", nil, false},
	{"cidadão sem role", "USER", nil, false},
	{"editor de cursos", "USER", []string{"go:cursos:editor"}, false},
	{"admin de cursos", "USER", []string{"go:cursos:admin"}, false},
	{"casa civil de cursos", "USER", []string{"go:cursos:casa_civil"}, false},
	{"admin do banco de currículos", "USER", []string{"go:curriculos:admin"}, false},
	{"role desconhecida de empregabilidade", "USER", []string{"go:empregabilidade:leitor"}, false},
	{"role parecida, mas não exata", "USER", []string{"go:empregabilidade:editor:x"}, false},
	{"ADMIN", "ADMIN", []string{"go:admin"}, true},
	{"admin de empregabilidade", "USER", []string{"go:empregabilidade:admin"}, true},
	{"editor de empregabilidade", "USER", []string{"go:empregabilidade:editor"}, true},
	{"editor sem curadoria", "USER", []string{"go:empregabilidade:editor_sem_curadoria"}, true},
	{"editor com curadoria", "USER", []string{"go:empregabilidade:editor_com_curadoria"}, true},
	{"editor de cursos e de empregabilidade", "USER", []string{"go:cursos:editor", "go:empregabilidade:editor"}, true},
}

func setMEIUserContext(role string, roles []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if role != "" {
			c.Set(middlewares.UserRoleKey, role)
		}
		if roles != nil {
			c.Set(middlewares.UserRolesKey, roles)
		}
		c.Next()
	}
}

func TestMEIGestaoAuthorization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	for _, tt := range meiGestaoCases {
		for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete} {
			t.Run(tt.name+" "+method, func(t *testing.T) {
				r := gin.New()
				r.Use(setMEIUserContext(tt.role, tt.roles))
				reached := false
				r.Handle(method, "/test", middlewares.MEIGestaoAuthorization(), func(c *gin.Context) {
					reached = true
					c.Status(http.StatusOK)
				})

				w := httptest.NewRecorder()
				r.ServeHTTP(w, httptest.NewRequest(method, "/test", nil))

				if tt.allow {
					assert.Equal(t, http.StatusOK, w.Code)
				} else {
					assert.Equal(t, http.StatusForbidden, w.Code)
				}
				assert.Equal(t, tt.allow, reached)
			})
		}
	}
}

func TestCanGerirMEI(t *testing.T) {
	gin.SetMode(gin.TestMode)

	for _, tt := range meiGestaoCases {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(setMEIUserContext(tt.role, tt.roles))
			var got bool
			r.GET("/test", func(c *gin.Context) {
				got = middlewares.CanGerirMEI(c)
				c.Status(http.StatusOK)
			})

			r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/test", nil))

			assert.Equal(t, tt.allow, got)
		})
	}
}
