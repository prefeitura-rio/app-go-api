package middlewares_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
)

const jobAuthTestJobID = "5f0c7d8e-1a2b-4c3d-9e8f-0a1b2c3d4e5f"

type jobAuthLoaders struct {
	jobCursoID  int
	jobFound    bool
	jobErr      error
	jobCalls    int
	orgaoID     string
	courseFound bool
	courseErr   error
	courseCalls int
	courseID    int
}

func (l *jobAuthLoaders) jobLoader(_ context.Context, _ uuid.UUID) (int, bool, error) {
	l.jobCalls++
	return l.jobCursoID, l.jobFound, l.jobErr
}

func (l *jobAuthLoaders) courseLoader(_ context.Context, id int) (string, bool, error) {
	l.courseCalls++
	l.courseID = id
	return l.orgaoID, l.courseFound, l.courseErr
}

func serveJobAccess(t *testing.T, inject gin.HandlerFunc, l *jobAuthLoaders, jobID string) *httptest.ResponseRecorder {
	t.Helper()
	r := setupCourseTestRouter(inject)
	r.GET("/api/v1/jobs/:jobId/status", middlewares.JobAccess(l.jobLoader, l.courseLoader), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/jobs/"+jobID+"/status", nil))
	return w
}

func injectAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(middlewares.UserRoleKey, "ADMIN")
		c.Next()
	}
}

func ownedJob() *jobAuthLoaders {
	return &jobAuthLoaders{jobCursoID: 7, jobFound: true, orgaoID: "orgao-a", courseFound: true}
}

func TestJobAccess_NoRole_Forbidden_DoesNotLoadJob(t *testing.T) {
	l := ownedJob()
	w := serveJobAccess(t, injectCourseRoles("", "", nil), l, jobAuthTestJobID)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, 0, l.jobCalls, "cidadão sem papel não deve nem descobrir se o job existe")
}

func TestJobAccess_SecretariaOnly_Forbidden(t *testing.T) {
	l := ownedJob()
	w := serveJobAccess(t, injectCourseRoles("", "", []string{"orgao-a"}), l, jobAuthTestJobID)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, 0, l.jobCalls)
}

func TestJobAccess_EditorWithoutScope_Forbidden(t *testing.T) {
	l := ownedJob()
	w := serveJobAccess(t, injectCourseRoles("go:cursos:editor", "", nil), l, jobAuthTestJobID)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, 0, l.jobCalls)
}

func TestJobAccess_Admin_Passes(t *testing.T) {
	l := ownedJob()
	w := serveJobAccess(t, injectAdmin(), l, jobAuthTestJobID)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 7, l.courseID, "o curso checado é o do metadata do job")
}

func TestJobAccess_CasaCivil_Passes(t *testing.T) {
	l := ownedJob()
	w := serveJobAccess(t, injectCourseRoles("go:cursos:casa_civil", "", nil), l, jobAuthTestJobID)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestJobAccess_EditorSameOrgao_Passes(t *testing.T) {
	l := ownedJob()
	w := serveJobAccess(t, injectCourseRoles("go:cursos:editor", "orgao-a", nil), l, jobAuthTestJobID)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestJobAccess_EditorSecretariaCoversOrgao_Passes(t *testing.T) {
	l := ownedJob()
	w := serveJobAccess(t, injectCourseRoles("go:cursos:editor", "", []string{"orgao-x", "orgao-a"}), l, jobAuthTestJobID)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestJobAccess_EditorOtherOrgao_Forbidden(t *testing.T) {
	l := ownedJob()
	w := serveJobAccess(t, injectCourseRoles("go:cursos:editor", "orgao-b", nil), l, jobAuthTestJobID)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, 7, l.courseID)
}

func TestJobAccess_InvalidJobID_BadRequest(t *testing.T) {
	l := ownedJob()
	w := serveJobAccess(t, injectAdmin(), l, "nao-e-uuid")

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, 0, l.jobCalls)
}

func TestJobAccess_JobNotFound_NotFound(t *testing.T) {
	l := &jobAuthLoaders{jobFound: false}
	w := serveJobAccess(t, injectAdmin(), l, jobAuthTestJobID)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, 0, l.courseCalls)
}

func TestJobAccess_JobLoaderError_InternalServerError(t *testing.T) {
	l := &jobAuthLoaders{jobErr: errors.New("db down: detalhe interno")}
	w := serveJobAccess(t, injectAdmin(), l, jobAuthTestJobID)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NotContains(t, w.Body.String(), "detalhe interno")
}

func TestJobAccess_CourseNotFound_NotFound(t *testing.T) {
	l := &jobAuthLoaders{jobCursoID: 7, jobFound: true, courseFound: false}
	w := serveJobAccess(t, injectCourseRoles("go:cursos:editor", "orgao-a", nil), l, jobAuthTestJobID)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestJobAccess_CourseLoaderError_InternalServerError(t *testing.T) {
	l := &jobAuthLoaders{jobCursoID: 7, jobFound: true, courseErr: errors.New("db down")}
	w := serveJobAccess(t, injectCourseRoles("go:cursos:editor", "orgao-a", nil), l, jobAuthTestJobID)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestJobAccess_JobWithoutCourse_AdminPasses(t *testing.T) {
	l := &jobAuthLoaders{jobCursoID: 0, jobFound: true}
	w := serveJobAccess(t, injectAdmin(), l, jobAuthTestJobID)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, l.courseCalls)
}

func TestJobAccess_JobWithoutCourse_EditorForbidden(t *testing.T) {
	l := &jobAuthLoaders{jobCursoID: 0, jobFound: true}
	w := serveJobAccess(t, injectCourseRoles("go:cursos:editor", "orgao-a", nil), l, jobAuthTestJobID)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, 0, l.courseCalls)
}
