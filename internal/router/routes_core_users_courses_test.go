package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/prefeitura-rio/app-go-api/internal/config"
	v1 "github.com/prefeitura-rio/app-go-api/internal/handlers/v1"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	"github.com/prefeitura-rio/app-go-api/internal/models"
)

// cursoListSvcStub atende GET /users/:userId/courses e conta as chamadas a
// List, para provar que um chamador barrado nunca chega ao serviço. Qualquer
// outro método entra em pânico (interface embutida nula).
type cursoListSvcStub struct {
	cursoServiceNil
	calls int
}

type cursoServiceNil interface {
	Create(ctx context.Context, curso *models.Curso) (int, error)
	GetByID(ctx context.Context, id int) (*models.Curso, error)
	Update(ctx context.Context, curso *models.Curso) error
	Delete(ctx context.Context, id int) error
	SendToReview(ctx context.Context, id int) error
	Approve(ctx context.Context, id int) error
	RequestChanges(ctx context.Context, id int) error
	RequestDeletion(ctx context.Context, id int) error
}

func (s *cursoListSvcStub) List(_ context.Context, _ map[string]interface{}, _, _ int) ([]*models.Curso, int, error) {
	s.calls++
	return []*models.Curso{}, 0, nil
}

// GET /api/v1/users/:userId/courses lista todos os cursos (inclusive rascunhos
// e em revisão) do órgão do :userId. Exige perfil de Cursos mesmo com
// RBAC_ENABLED=false (DEV-506): ADMIN, go:cursos:casa_civil ou
// go:cursos:editor. Com o RBAC ligado, o editor só lista órgão do seu escopo.
func TestRegisterCoreRoutes_UserCourses_Authorization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const orgao = "orgao-a"

	callers := []struct {
		name        string
		rbacEnabled bool
		role        string
		roles       []string
		orgaoID     string
		secretarias []string
		allowed     bool
	}{
		{"cidadão sem role, RBAC desligado", false, "USER", nil, "", nil, false},
		{"cidadão sem role, RBAC ligado", true, "USER", nil, "", nil, false},
		{"admin de empregabilidade, RBAC desligado", false, "USER", []string{"go:empregabilidade:admin"}, "", nil, false},
		{"admin de empregabilidade, RBAC ligado", true, "USER", []string{"go:empregabilidade:admin"}, "", nil, false},
		{"editor de empregabilidade do órgão, RBAC ligado", true, "USER", []string{"go:empregabilidade:editor"}, orgao, nil, false},
		{"admin MEI, RBAC desligado", false, "USER", []string{"go:mei:admin"}, orgao, nil, false},
		{"só secretaria mapeada, sem role, RBAC ligado", true, "USER", nil, "", []string{orgao}, false},
		{"editor de cursos de outro órgão, RBAC ligado", true, "USER", []string{"go:cursos:editor"}, "orgao-b", nil, false},
		{"editor de cursos de outra secretaria, RBAC ligado", true, "USER", []string{"go:cursos:editor"}, "", []string{"orgao-b"}, false},
		{"editor de cursos sem escopo, RBAC ligado", true, "USER", []string{"go:cursos:editor"}, "", nil, false},
		{"ADMIN, RBAC desligado", false, "ADMIN", []string{"go:admin"}, "", nil, true},
		{"ADMIN, RBAC ligado", true, "ADMIN", []string{"go:admin"}, "", nil, true},
		{"casa civil, RBAC desligado", false, "USER", []string{"go:cursos:casa_civil"}, "", nil, true},
		{"casa civil, RBAC ligado", true, "USER", []string{"go:cursos:casa_civil"}, "", nil, true},
		{"editor de cursos sem escopo, RBAC desligado", false, "USER", []string{"go:cursos:editor"}, "", nil, true},
		{"editor de cursos do órgão (go:orgao), RBAC ligado", true, "USER", []string{"go:cursos:editor"}, orgao, nil, true},
		{"editor de cursos da secretaria, RBAC ligado", true, "USER", []string{"go:cursos:editor"}, "", []string{"orgao-b", orgao}, true},
	}

	for _, tt := range callers {
		t.Run(tt.name, func(t *testing.T) {
			svc := &cursoListSvcStub{}
			app := createMockApplicationContainer()
			app.CourseHandler = v1.NewCourseHandler(svc, nil, nil)

			router := gin.New()
			apiV1 := router.Group("/api/v1", injectCaller(tt.role, tt.roles, tt.orgaoID), func(c *gin.Context) {
				if tt.secretarias != nil {
					c.Set(middlewares.UserSecretariaOrgaoIDsKey, tt.secretarias)
				}
				c.Next()
			})
			cfg := &config.AppConfig{App: config.AppSettings{RBACEnabled: tt.rbacEnabled}}
			registerCoreRoutes(apiV1, router.Group("/api/public"), app, cfg)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/users/"+orgao+"/courses?limit=10", nil))

			if tt.allowed {
				assert.Equal(t, http.StatusOK, w.Code, w.Body.String())
				assert.Equal(t, 1, svc.calls, "chamador autorizado tem de chegar ao serviço")
			} else {
				assert.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
				assert.Equal(t, 0, svc.calls, "chamador barrado não pode chegar ao serviço")
			}
		})
	}
}
