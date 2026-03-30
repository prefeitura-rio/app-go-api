package v1_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	v1 "github.com/prefeitura-rio/app-go-api/internal/handlers/v1"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
	"github.com/stretchr/testify/assert"
)

type mockJobRepoForHandler struct {
	job    *models.Job
	getErr error
}

func (m *mockJobRepoForHandler) GetByID(_ context.Context, id uuid.UUID) (*models.Job, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.job, nil
}

func (m *mockJobRepoForHandler) Create(_ context.Context, job *models.Job) error {
	return nil
}

func (m *mockJobRepoForHandler) Update(_ context.Context, job *models.Job) error {
	return nil
}

func (m *mockJobRepoForHandler) Delete(_ context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockJobRepoForHandler) List(_ context.Context, _ map[string]interface{}, _, _ int) ([]*models.Job, int, error) {
	return []*models.Job{}, 0, nil
}

func (m *mockJobRepoForHandler) UpdateProgress(_ context.Context, _ uuid.UUID, _, _, _ int) error {
	return nil
}

func (m *mockJobRepoForHandler) UpdateStatus(_ context.Context, _ uuid.UUID, _ models.JobStatus) error {
	return nil
}

func setupJobRouter(repo services.JobRepositoryInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	svc := services.NewJobServiceWithInterface(repo)
	h := v1.NewJobHandler(svc)
	r.GET("/api/v1/jobs/:jobId/status", h.GetStatus)
	return r
}

func TestJobHandler_GetStatus_Success(t *testing.T) {
	jobID := uuid.New()
	repo := &mockJobRepoForHandler{
		job: &models.Job{
			ID:       jobID,
			Type:     "import_enrollments",
			Status:   "completed",
			Progress: 100,
		},
	}
	router := setupJobRouter(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/jobs/"+jobID.String()+"/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
	assert.Contains(t, w.Body.String(), "completed")
}

func TestJobHandler_GetStatus_InvalidUUID(t *testing.T) {
	repo := &mockJobRepoForHandler{}
	router := setupJobRouter(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/jobs/invalid-uuid/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "ID do job inválido")
}

func TestJobHandler_GetStatus_NotFound(t *testing.T) {
	jobID := uuid.New()
	repo := &mockJobRepoForHandler{
		job: nil,
	}
	router := setupJobRouter(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/jobs/"+jobID.String()+"/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Job não encontrado")
}

func TestJobHandler_GetStatus_ServiceError(t *testing.T) {
	jobID := uuid.New()
	repo := &mockJobRepoForHandler{
		getErr: errors.New("database error"),
	}
	router := setupJobRouter(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/jobs/"+jobID.String()+"/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Erro ao buscar job")
}

func TestNewJobHandler(t *testing.T) {
	repo := &mockJobRepoForHandler{}
	svc := services.NewJobServiceWithInterface(repo)
	handler := v1.NewJobHandler(svc)

	assert.NotNil(t, handler)
}
