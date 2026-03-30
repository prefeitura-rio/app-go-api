package v1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	v1 "github.com/prefeitura-rio/app-go-api/internal/handlers/v1"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockInscricaoService is a mock implementation of InscricaoServiceInterface
type MockInscricaoService struct {
	mock.Mock
}

func (m *MockInscricaoService) Create(ctx context.Context, inscricao *models.Inscricao) error {
	args := m.Called(ctx, inscricao)
	return args.Error(0)
}

func (m *MockInscricaoService) CreateManual(ctx context.Context, inscricao *models.Inscricao) error {
	args := m.Called(ctx, inscricao)
	return args.Error(0)
}

func (m *MockInscricaoService) GetByID(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Inscricao), args.Error(1)
}

func (m *MockInscricaoService) GetByCursoID(ctx context.Context, cursoID int, filter map[string]interface{}, page, pageSize int) ([]*models.Inscricao, int, error) {
	args := m.Called(ctx, cursoID, filter, page, pageSize)
	return args.Get(0).([]*models.Inscricao), args.Int(1), args.Error(2)
}

func (m *MockInscricaoService) UpdateStatus(ctx context.Context, inscricaoID uuid.UUID, status models.StatusInscricao, reason, adminNotes string) error {
	args := m.Called(ctx, inscricaoID, status, reason, adminNotes)
	return args.Error(0)
}

func (m *MockInscricaoService) UpdateMultipleStatus(ctx context.Context, inscricaoIDs []uuid.UUID, status models.StatusInscricao, reason, adminNotes string) (int, error) {
	args := m.Called(ctx, inscricaoIDs, status, reason, adminNotes)
	return args.Int(0), args.Error(1)
}

func (m *MockInscricaoService) GetSummaryByCursoID(ctx context.Context, cursoID int) (*models.EnrollmentSummary, error) {
	args := m.Called(ctx, cursoID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.EnrollmentSummary), args.Error(1)
}

func (m *MockInscricaoService) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockInscricaoService) ListByCPF(ctx context.Context, cpf string, filter map[string]interface{}, offset, limit int) ([]*models.Inscricao, int, error) {
	args := m.Called(ctx, cpf, filter, offset, limit)
	return args.Get(0).([]*models.Inscricao), args.Int(1), args.Error(2)
}

func (m *MockInscricaoService) UpdateCertificate(ctx context.Context, cursoID int, inscricaoID uuid.UUID, certificateURL string) error {
	args := m.Called(ctx, cursoID, inscricaoID, certificateURL)
	return args.Error(0)
}

func (m *MockInscricaoService) UpdateInscricao(ctx context.Context, id uuid.UUID, cursoID int, updateData *models.InscricaoUpdateRequest) error {
	args := m.Called(ctx, id, cursoID, updateData)
	return args.Error(0)
}

func (m *MockInscricaoService) EnrichWithPersonalInfo(ctx context.Context, inscricao *models.Inscricao) {
	m.Called(ctx, inscricao)
}

func (m *MockInscricaoService) EnrichMultipleWithPersonalInfo(ctx context.Context, inscricoes []*models.Inscricao) {
	m.Called(ctx, inscricoes)
}

func (m *MockInscricaoService) ChangeSchedule(ctx context.Context, inscricaoID uuid.UUID, userCPF string, request *models.ScheduleChangeRequest) (*models.Inscricao, error) {
	args := m.Called(ctx, inscricaoID, userCPF, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Inscricao), args.Error(1)
}

// MockJobService is a mock implementation of JobServiceInterface
type MockJobService struct {
	mock.Mock
}

func (m *MockJobService) Create(ctx context.Context, job *models.Job) error {
	args := m.Called(ctx, job)
	if args.Get(0) != nil {
		// Set ID if provided in mock setup
		if id, ok := args.Get(0).(uuid.UUID); ok {
			job.ID = id
		}
	}
	return args.Error(0)
}

func (m *MockJobService) GetByID(ctx context.Context, id uuid.UUID) (*models.Job, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Job), args.Error(1)
}

func (m *MockJobService) Update(ctx context.Context, job *models.Job) error {
	args := m.Called(ctx, job)
	return args.Error(0)
}

func (m *MockJobService) UpdateStatus(ctx context.Context, id uuid.UUID, status models.JobStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockJobService) UpdateProgress(ctx context.Context, id uuid.UUID, progress, successCount, errorCount int) error {
	args := m.Called(ctx, id, progress, successCount, errorCount)
	return args.Error(0)
}

func (m *MockJobService) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Job, int, error) {
	args := m.Called(ctx, filter, limit, offset)
	return args.Get(0).([]*models.Job), args.Int(1), args.Error(2)
}

func (m *MockJobService) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockCursoRepository is a mock implementation of CursoRepositoryInterface
type MockCursoRepository struct {
	mock.Mock
}

func (m *MockCursoRepository) Create(ctx context.Context, curso *models.Curso) (int, error) {
	args := m.Called(ctx, curso)
	return args.Int(0), args.Error(1)
}

func (m *MockCursoRepository) GetByID(ctx context.Context, id int) (*models.Curso, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Curso), args.Error(1)
}

func (m *MockCursoRepository) Update(ctx context.Context, curso *models.Curso) error {
	args := m.Called(ctx, curso)
	return args.Error(0)
}

func (m *MockCursoRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCursoRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Curso, int, error) {
	args := m.Called(ctx, filter, limit, offset)
	return args.Get(0).([]*models.Curso), args.Int(1), args.Error(2)
}

func (m *MockCursoRepository) CreateCustomFields(ctx context.Context, fields []models.CustomField) error {
	args := m.Called(ctx, fields)
	return args.Error(0)
}

func (m *MockCursoRepository) CreateRemoteClass(ctx context.Context, remoteClass *models.RemoteClass) error {
	args := m.Called(ctx, remoteClass)
	return args.Error(0)
}

func (m *MockCursoRepository) CreateLocationClasses(ctx context.Context, locations []models.LocationClass) error {
	args := m.Called(ctx, locations)
	return args.Error(0)
}

func (m *MockCursoRepository) ValidateForEnrollment(ctx context.Context, cursoID int) (status string, enrollmentStart *time.Time, enrollmentEnd *time.Time, autoApprove bool, err error) {
	args := m.Called(ctx, cursoID)
	return args.String(0), args.Get(1).(*time.Time), args.Get(2).(*time.Time), args.Bool(3), args.Error(4)
}

func (m *MockCursoRepository) CountEnrollmentsByScheduleID(ctx context.Context, scheduleID uuid.UUID) (int64, error) {
	args := m.Called(ctx, scheduleID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCursoRepository) CountEnrollmentsByScheduleIDs(ctx context.Context, scheduleIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	args := m.Called(ctx, scheduleIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[uuid.UUID]int64), args.Error(1)
}

func (m *MockCursoRepository) GetCourseScheduleByID(ctx context.Context, scheduleID uuid.UUID) (*models.CourseSchedule, error) {
	args := m.Called(ctx, scheduleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CourseSchedule), args.Error(1)
}

func (m *MockCursoRepository) GetRemoteScheduleByID(ctx context.Context, scheduleID uuid.UUID) (*models.RemoteSchedule, error) {
	args := m.Called(ctx, scheduleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RemoteSchedule), args.Error(1)
}

// Test Create with service success
func TestInscricaoHandler_Create_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockInscricaoService)
	mockJobService := new(MockJobService)
	mockCursoRepo := new(MockCursoRepository)

	handler := v1.NewInscricaoHandler(mockService, mockJobService, mockCursoRepo)

	r := gin.New()
	r.POST("/api/v1/courses/:courseId/enrollments", handler.Create)

	inscricaoID := uuid.New()
	mockService.On("Create", mock.Anything, mock.AnythingOfType("*models.Inscricao")).
		Run(func(args mock.Arguments) {
			inscricao := args.Get(1).(*models.Inscricao)
			inscricao.ID = inscricaoID
			inscricao.Status = models.StatusInscricaoPending
			inscricao.EnrolledAt = time.Now()
		}).
		Return(nil)

	reqBody := map[string]interface{}{
		"cpf":   "12345678901",
		"name":  "Test User",
		"email": "test@example.com",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses/1/enrollments", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "success")
	mockService.AssertExpectations(t)
}

// Test Create with duplicate enrollment (409 Conflict)
func TestInscricaoHandler_Create_DuplicateEnrollment(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockInscricaoService)
	mockJobService := new(MockJobService)
	mockCursoRepo := new(MockCursoRepository)

	handler := v1.NewInscricaoHandler(mockService, mockJobService, mockCursoRepo)

	r := gin.New()
	r.POST("/api/v1/courses/:courseId/enrollments", handler.Create)

	mockService.On("Create", mock.Anything, mock.AnythingOfType("*models.Inscricao")).
		Return(errors.New("CPF já inscrito neste curso"))

	reqBody := map[string]interface{}{
		"cpf": "12345678901",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses/1/enrollments", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "CPF já inscrito")
	mockService.AssertExpectations(t)
}

// Test Create with service error
func TestInscricaoHandler_Create_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockInscricaoService)
	mockJobService := new(MockJobService)
	mockCursoRepo := new(MockCursoRepository)

	handler := v1.NewInscricaoHandler(mockService, mockJobService, mockCursoRepo)

	r := gin.New()
	r.POST("/api/v1/courses/:courseId/enrollments", handler.Create)

	mockService.On("Create", mock.Anything, mock.AnythingOfType("*models.Inscricao")).
		Return(errors.New("database error"))

	reqBody := map[string]interface{}{
		"cpf": "12345678901",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses/1/enrollments", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// Test List with success
func TestInscricaoHandler_List_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockInscricaoService)
	mockJobService := new(MockJobService)
	mockCursoRepo := new(MockCursoRepository)

	handler := v1.NewInscricaoHandler(mockService, mockJobService, mockCursoRepo)

	r := gin.New()
	r.GET("/api/v1/courses/:courseId/enrollments", handler.List)

	inscricoes := []*models.Inscricao{
		{
			ID:      uuid.New(),
			CursoID: 1,
			CPF:     "12345678901",
			Status:  models.StatusInscricaoPending,
		},
	}

	summary := &models.EnrollmentSummary{
		Total:    10,
		Pending:  5,
		Approved: 3,
		Rejected: 2,
	}

	mockService.On("GetByCursoID", mock.Anything, 1, mock.Anything, 1, 20).
		Return(inscricoes, 10, nil)
	mockCursoRepo.On("CountEnrollmentsByScheduleIDs", mock.Anything, mock.Anything).
		Return(make(map[uuid.UUID]int64), nil).Maybe()
	mockService.On("EnrichMultipleWithPersonalInfo", mock.Anything, mock.Anything).
		Return()
	mockService.On("GetSummaryByCursoID", mock.Anything, 1).
		Return(summary, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses/1/enrollments?page=1&limit=20", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
	mockService.AssertExpectations(t)
}

// Test List with filters
func TestInscricaoHandler_List_WithFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockInscricaoService)
	mockJobService := new(MockJobService)
	mockCursoRepo := new(MockCursoRepository)

	handler := v1.NewInscricaoHandler(mockService, mockJobService, mockCursoRepo)

	r := gin.New()
	r.GET("/api/v1/courses/:courseId/enrollments", handler.List)

	mockService.On("GetByCursoID", mock.Anything, 1, mock.MatchedBy(func(filter map[string]interface{}) bool {
		return filter["status"] == "approved" && filter["search"] == "test"
	}), 1, 20).
		Return([]*models.Inscricao{}, 0, nil)
	mockCursoRepo.On("CountEnrollmentsByScheduleIDs", mock.Anything, mock.Anything).
		Return(make(map[uuid.UUID]int64), nil)
	mockService.On("EnrichMultipleWithPersonalInfo", mock.Anything, mock.Anything).
		Return()
	mockService.On("GetSummaryByCursoID", mock.Anything, 1).
		Return(&models.EnrollmentSummary{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses/1/enrollments?status=approved&search=test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// Test Update with success
func TestInscricaoHandler_Update_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockInscricaoService)
	mockJobService := new(MockJobService)
	mockCursoRepo := new(MockCursoRepository)

	handler := v1.NewInscricaoHandler(mockService, mockJobService, mockCursoRepo)

	r := gin.New()
	r.PUT("/api/v1/courses/:courseId/enrollments/:enrollmentId", handler.Update)

	enrollmentID := uuid.New()
	updatedInscricao := &models.Inscricao{
		ID:      enrollmentID,
		CursoID: 1,
		CPF:     "12345678901",
		Name:    "Updated Name",
	}

	mockService.On("UpdateInscricao", mock.Anything, enrollmentID, 1, mock.AnythingOfType("*models.InscricaoUpdateRequest")).
		Return(nil)
	mockService.On("GetByID", mock.Anything, enrollmentID).
		Return(updatedInscricao, nil)

	reqBody := map[string]interface{}{
		"name": "Updated Name",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/courses/1/enrollments/"+enrollmentID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// Test Update with not found error
func TestInscricaoHandler_Update_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockInscricaoService)
	mockJobService := new(MockJobService)
	mockCursoRepo := new(MockCursoRepository)

	handler := v1.NewInscricaoHandler(mockService, mockJobService, mockCursoRepo)

	r := gin.New()
	r.PUT("/api/v1/courses/:courseId/enrollments/:enrollmentId", handler.Update)

	enrollmentID := uuid.New()

	mockService.On("UpdateInscricao", mock.Anything, enrollmentID, 1, mock.AnythingOfType("*models.InscricaoUpdateRequest")).
		Return(errors.New("inscrição não encontrada"))

	reqBody := map[string]interface{}{
		"name": "Updated Name",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/courses/1/enrollments/"+enrollmentID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

// Test UpdateStatus with success
func TestInscricaoHandler_UpdateStatus_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockInscricaoService)
	mockJobService := new(MockJobService)
	mockCursoRepo := new(MockCursoRepository)

	handler := v1.NewInscricaoHandler(mockService, mockJobService, mockCursoRepo)

	r := gin.New()
	r.PUT("/api/v1/courses/:courseId/enrollments/status", handler.UpdateStatus)

	ids := []uuid.UUID{uuid.New(), uuid.New()}

	mockService.On("UpdateMultipleStatus", mock.Anything, mock.MatchedBy(func(enrollmentIDs []uuid.UUID) bool {
		return len(enrollmentIDs) == 2
	}), models.StatusInscricaoApproved, "Test reason", "Admin notes").
		Return(2, nil)

	reqBody := map[string]interface{}{
		"enrollment_ids": ids,
		"status":         "approved",
		"reason":         "Test reason",
		"admin_notes":    "Admin notes",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/courses/1/enrollments/status", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "2")
	mockService.AssertExpectations(t)
}

// Test UpdateIndividualStatus with success
func TestInscricaoHandler_UpdateIndividualStatus_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockInscricaoService)
	mockJobService := new(MockJobService)
	mockCursoRepo := new(MockCursoRepository)

	handler := v1.NewInscricaoHandler(mockService, mockJobService, mockCursoRepo)

	r := gin.New()
	r.PUT("/api/v1/courses/:courseId/enrollments/:enrollmentId/status", handler.UpdateIndividualStatus)

	enrollmentID := uuid.New()

	mockService.On("UpdateStatus", mock.Anything, enrollmentID, models.StatusInscricaoApproved, "Approved", "Notes").
		Return(nil)

	reqBody := map[string]interface{}{
		"status":      "approved",
		"reason":      "Approved",
		"admin_notes": "Notes",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/courses/1/enrollments/"+enrollmentID.String()+"/status", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
	mockService.AssertExpectations(t)
}

// Test GetByID with success
func TestInscricaoHandler_GetByID_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockInscricaoService)
	mockJobService := new(MockJobService)
	mockCursoRepo := new(MockCursoRepository)

	handler := v1.NewInscricaoHandler(mockService, mockJobService, mockCursoRepo)

	r := gin.New()
	// Add context setter middleware for user_role
	r.Use(func(c *gin.Context) {
		c.Set("user_role", "ADMIN")
		c.Next()
	})
	r.GET("/api/v1/courses/:courseId/enrollments/:enrollmentId", handler.GetByID)

	enrollmentID := uuid.New()
	inscricao := &models.Inscricao{
		ID:      enrollmentID,
		CursoID: 1,
		CPF:     "12345678901",
	}

	mockService.On("GetByID", mock.Anything, enrollmentID).
		Return(inscricao, nil)
	mockCursoRepo.On("CountEnrollmentsByScheduleIDs", mock.Anything, mock.Anything).
		Return(make(map[uuid.UUID]int64), nil)
	mockService.On("EnrichWithPersonalInfo", mock.Anything, inscricao).
		Return()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses/1/enrollments/"+enrollmentID.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
	mockService.AssertExpectations(t)
}

// Test GetByID with not found
func TestInscricaoHandler_GetByID_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockInscricaoService)
	mockJobService := new(MockJobService)
	mockCursoRepo := new(MockCursoRepository)

	handler := v1.NewInscricaoHandler(mockService, mockJobService, mockCursoRepo)

	r := gin.New()
	r.GET("/api/v1/courses/:courseId/enrollments/:enrollmentId", handler.GetByID)

	enrollmentID := uuid.New()

	mockService.On("GetByID", mock.Anything, enrollmentID).
		Return(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses/1/enrollments/"+enrollmentID.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

// Test Delete with success
func TestInscricaoHandler_Delete_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockInscricaoService)
	mockJobService := new(MockJobService)
	mockCursoRepo := new(MockCursoRepository)

	handler := v1.NewInscricaoHandler(mockService, mockJobService, mockCursoRepo)

	r := gin.New()
	r.DELETE("/api/v1/courses/:courseId/enrollments/:enrollmentId", handler.Delete)

	enrollmentID := uuid.New()

	mockService.On("Delete", mock.Anything, enrollmentID).
		Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/courses/1/enrollments/"+enrollmentID.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
	mockService.AssertExpectations(t)
}

// Test Delete with not found
func TestInscricaoHandler_Delete_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockInscricaoService)
	mockJobService := new(MockJobService)
	mockCursoRepo := new(MockCursoRepository)

	handler := v1.NewInscricaoHandler(mockService, mockJobService, mockCursoRepo)

	r := gin.New()
	r.DELETE("/api/v1/courses/:courseId/enrollments/:enrollmentId", handler.Delete)

	enrollmentID := uuid.New()

	mockService.On("Delete", mock.Anything, enrollmentID).
		Return(errors.New("inscrição não encontrada"))

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/courses/1/enrollments/"+enrollmentID.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

// Test ListByUser with authorization
func TestInscricaoHandler_ListByUser_Success_Admin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockInscricaoService)
	mockJobService := new(MockJobService)
	mockCursoRepo := new(MockCursoRepository)

	handler := v1.NewInscricaoHandler(mockService, mockJobService, mockCursoRepo)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_role", "ADMIN")
		c.Set("user_cpf", "98765432100")
		c.Next()
	})
	r.GET("/api/v1/enrollments/user/:cpf", handler.ListByUser)

	inscricoes := []*models.Inscricao{
		{
			ID:      uuid.New(),
			CursoID: 1,
			CPF:     "12345678901",
		},
	}

	mockService.On("ListByCPF", mock.Anything, "12345678901", mock.Anything, 0, 10).
		Return(inscricoes, 1, nil)
	mockCursoRepo.On("CountEnrollmentsByScheduleIDs", mock.Anything, mock.Anything).
		Return(make(map[uuid.UUID]int64), nil)
	mockService.On("EnrichMultipleWithPersonalInfo", mock.Anything, inscricoes).
		Return()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/enrollments/user/12345678901", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// Test ListByUser with forbidden access (non-admin trying to view another user)
func TestInscricaoHandler_ListByUser_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockInscricaoService)
	mockJobService := new(MockJobService)
	mockCursoRepo := new(MockCursoRepository)

	handler := v1.NewInscricaoHandler(mockService, mockJobService, mockCursoRepo)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_role", "USER")
		c.Set("user_cpf", "98765432100")
		c.Next()
	})
	r.GET("/api/v1/enrollments/user/:cpf", handler.ListByUser)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/enrollments/user/12345678901", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
