package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	"github.com/stretchr/testify/assert"
)

func noAuthHandler(c *gin.Context) { c.Status(http.StatusOK) }

func setupSelfEnrollRouter(injector gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	if injector != nil {
		r.Use(injector)
	}
	r.POST("/api/v1/courses/:courseId/enrollments", noAuthHandler)
	return r
}

func TestSelfEnrollment_RegressionMatrix(t *testing.T) {
	tests := []struct {
		name          string
		role          string
		userRole      string
		orgaoID       string
		secretariaIDs []string
		wantStatus    int
	}{
		{"cursos_editor_no_secretaria", "go:cursos:editor", "", "", nil, http.StatusOK},
		{"empregabilidade_editor_com_curadoria_with_secretaria", "go:empregabilidade:editor_com_curadoria", "", "", []string{"1"}, http.StatusOK},
		{"empregabilidade_editor_sem_curadoria", "go:empregabilidade:editor_sem_curadoria", "", "", nil, http.StatusOK},
		{"platform_admin_with_secretarias", "", "ADMIN", "", []string{"3000", "3001", "300ASASAS1"}, http.StatusOK},
		{"go_admin_no_secretaria", "go:admin", "", "", nil, http.StatusOK},
		{"cursos_casa_civil_with_secretaria", "go:cursos:casa_civil", "", "", []string{"3000", "1"}, http.StatusOK},
		{"empregabilidade_admin_no_secretaria", "go:empregabilidade:admin", "", "", nil, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			injector := func(c *gin.Context) {
				if tt.userRole != "" {
					c.Set(middlewares.UserRoleKey, tt.userRole)
				}
				if tt.role != "" {
					c.Set(middlewares.UserRolesKey, []string{tt.role})
				}
				if tt.orgaoID != "" {
					c.Set(middlewares.UserGroupsKey, []string{"go:orgao:" + tt.orgaoID})
				}
				if tt.secretariaIDs != nil {
					c.Set(middlewares.UserSecretariaOrgaoIDsKey, tt.secretariaIDs)
				}
				c.Next()
			}

			r := setupSelfEnrollRouter(injector)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/courses/1/enrollments", nil))

			assert.Equal(t, tt.wantStatus, w.Code,
				"role=%q userRole=%q secretariaIDs=%v", tt.role, tt.userRole, tt.secretariaIDs)
		})
	}
}

func TestSelfEnrollment_NoAuth_NotForbidden(t *testing.T) {
	r := setupSelfEnrollRouter(nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/courses/42/enrollments", nil))
	assert.NotEqual(t, http.StatusForbidden, w.Code)
}
