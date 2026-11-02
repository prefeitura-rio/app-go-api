package router

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"

	v1 "github.com/prefeitura-rio/app-go-api/internal/handlers/v1"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

// jobRepoGetByIDStub satisfies services.JobRepositoryInterface for GET
// /jobs/:jobId/status; any other method panics (nil embedded interface).
type jobRepoGetByIDStub struct {
	services.JobRepositoryInterface
	job   *models.Job
	calls int
}

func (s *jobRepoGetByIDStub) GetByID(_ context.Context, _ uuid.UUID) (*models.Job, error) {
	s.calls++
	return s.job, nil
}

// newJobStatusRouter wires the real core routes with a job of course 1
// (orgao-a) whose errors carry a row identifier, and the caller's context set
// by inject.
func newJobStatusRouter(t *testing.T, inject gin.HandlerFunc, metadata string) (*gin.Engine, *jobRepoGetByIDStub, uuid.UUID) {
	t.Helper()
	jobID := uuid.New()
	errorsJSON, err := json.Marshal([]models.JobError{{Line: 2, Message: "CPF já inscrito neste curso", Data: "M. S. - 123******09"}})
	require.NoError(t, err)

	repo := &jobRepoGetByIDStub{job: &models.Job{
		ID:       jobID,
		Type:     models.JobTypeEnrollmentImport,
		Status:   models.JobStatusCompleted,
		Errors:   datatypes.JSON(errorsJSON),
		Metadata: datatypes.JSON(metadata),
	}}

	router := gin.New()
	apiV1 := router.Group("/api/v1")
	apiV1.Use(inject)
	apiPublic := router.Group("/api/public")

	app := createMockApplicationContainer()
	app.CursoService = services.NewCursoServiceWithInterface(&cursoRepoGetByIDStub{
		curso: &models.Curso{ID: 1, OrgaoID: "orgao-a"},
	})
	app.JobService = services.NewJobServiceWithInterface(repo)
	app.JobHandler = v1.NewJobHandler(app.JobService)
	registerCoreRoutes(apiV1, apiPublic, app, newTestCoreConfig())

	return router, repo, jobID
}

func getJobStatus(router *gin.Engine, jobID uuid.UUID) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/jobs/"+jobID.String()+"/status", nil))
	return w
}

// TestRegisterCoreRoutes_JobStatus_RequiresCourseAccess proves the GET
// /jobs/:jobId/status chain applies the same access control as the import that
// creates the job (DEV-149): only admin, casa civil or an editor scoped to the
// course's orgao reads it.
func TestRegisterCoreRoutes_JobStatus_RequiresCourseAccess(t *testing.T) {
	const metadata = `{"curso_id":1,"file_name":"inscricoes.csv"}`

	tests := []struct {
		name       string
		inject     gin.HandlerFunc
		wantStatus int
		wantLoaded bool
	}{
		{
			name: "cidadão autenticado sem papel",
			inject: func(c *gin.Context) {
				c.Set(middlewares.UserCPFKey, "98765432100")
				c.Next()
			},
			wantStatus: http.StatusForbidden,
			wantLoaded: false,
		},
		{
			name: "editor de outro órgão",
			inject: func(c *gin.Context) {
				c.Set(middlewares.UserRolesKey, []string{"go:cursos:editor"})
				c.Set(middlewares.UserSecretariaOrgaoIDsKey, []string{"orgao-b"})
				c.Next()
			},
			wantStatus: http.StatusForbidden,
			wantLoaded: true,
		},
		{
			name: "editor do órgão do curso",
			inject: func(c *gin.Context) {
				c.Set(middlewares.UserRolesKey, []string{"go:cursos:editor"})
				c.Set(middlewares.UserSecretariaOrgaoIDsKey, []string{"orgao-a"})
				c.Next()
			},
			wantStatus: http.StatusOK,
			wantLoaded: true,
		},
		{
			name: "admin",
			inject: func(c *gin.Context) {
				c.Set(middlewares.UserRoleKey, "ADMIN")
				c.Next()
			},
			wantStatus: http.StatusOK,
			wantLoaded: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, repo, jobID := newJobStatusRouter(t, tt.inject, metadata)

			w := getJobStatus(router, jobID)

			assert.Equal(t, tt.wantStatus, w.Code, w.Body.String())
			if tt.wantStatus != http.StatusOK {
				assert.NotContains(t, w.Body.String(), "errors")
				assert.NotContains(t, w.Body.String(), jobID.String())
			} else {
				assert.Contains(t, w.Body.String(), jobID.String())
			}
			assert.Equal(t, tt.wantLoaded, repo.calls > 0)
		})
	}
}

// TestRegisterCoreRoutes_JobStatus_JobWithoutCourse_OnlyAdmin covers a job whose
// metadata names no course (or is unreadable): no editor scope can be proven,
// so only admin / casa civil read it.
func TestRegisterCoreRoutes_JobStatus_JobWithoutCourse_OnlyAdmin(t *testing.T) {
	editor := func(c *gin.Context) {
		c.Set(middlewares.UserRolesKey, []string{"go:cursos:editor"})
		c.Set(middlewares.UserSecretariaOrgaoIDsKey, []string{"orgao-a"})
		c.Next()
	}
	admin := func(c *gin.Context) {
		c.Set(middlewares.UserRoleKey, "ADMIN")
		c.Next()
	}

	for _, metadata := range []string{`{}`, `{"curso_id":"abc"}`} {
		router, _, jobID := newJobStatusRouter(t, editor, metadata)
		assert.Equal(t, http.StatusForbidden, getJobStatus(router, jobID).Code, metadata)

		router, _, jobID = newJobStatusRouter(t, admin, metadata)
		assert.Equal(t, http.StatusOK, getJobStatus(router, jobID).Code, metadata)
	}
}
