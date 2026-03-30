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

// MockCursoService is a mock implementation of CursoServiceInterface
type MockCursoService struct {
	mock.Mock
}

func (m *MockCursoService) Create(ctx context.Context, curso *models.Curso) (int, error) {
	args := m.Called(ctx, curso)
	return args.Int(0), args.Error(1)
}

func (m *MockCursoService) GetByID(ctx context.Context, id int) (*models.Curso, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Curso), args.Error(1)
}

func (m *MockCursoService) Update(ctx context.Context, curso *models.Curso) error {
	args := m.Called(ctx, curso)
	return args.Error(0)
}

func (m *MockCursoService) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCursoService) List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*models.Curso, int, error) {
	args := m.Called(ctx, filter, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.Curso), args.Int(1), args.Error(2)
}

// MockInscricaoServiceForCourse is a minimal mock for InscricaoServiceInterface used in CourseHandler
type MockInscricaoServiceForCourse struct {
	mock.Mock
}

func (m *MockInscricaoServiceForCourse) Create(ctx context.Context, inscricao *models.Inscricao) error {
	args := m.Called(ctx, inscricao)
	return args.Error(0)
}

func (m *MockInscricaoServiceForCourse) CreateManual(ctx context.Context, inscricao *models.Inscricao) error {
	args := m.Called(ctx, inscricao)
	return args.Error(0)
}

func (m *MockInscricaoServiceForCourse) GetByID(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Inscricao), args.Error(1)
}

func (m *MockInscricaoServiceForCourse) GetByCursoID(ctx context.Context, cursoID int, filter map[string]interface{}, page, pageSize int) ([]*models.Inscricao, int, error) {
	args := m.Called(ctx, cursoID, filter, page, pageSize)
	return args.Get(0).([]*models.Inscricao), args.Int(1), args.Error(2)
}

func (m *MockInscricaoServiceForCourse) UpdateStatus(ctx context.Context, inscricaoID uuid.UUID, status models.StatusInscricao, reason, adminNotes string) error {
	args := m.Called(ctx, inscricaoID, status, reason, adminNotes)
	return args.Error(0)
}

func (m *MockInscricaoServiceForCourse) UpdateMultipleStatus(ctx context.Context, inscricaoIDs []uuid.UUID, status models.StatusInscricao, reason, adminNotes string) (int, error) {
	args := m.Called(ctx, inscricaoIDs, status, reason, adminNotes)
	return args.Int(0), args.Error(1)
}

func (m *MockInscricaoServiceForCourse) GetSummaryByCursoID(ctx context.Context, cursoID int) (*models.EnrollmentSummary, error) {
	args := m.Called(ctx, cursoID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.EnrollmentSummary), args.Error(1)
}

func (m *MockInscricaoServiceForCourse) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockInscricaoServiceForCourse) ListByCPF(ctx context.Context, cpf string, filter map[string]interface{}, offset, limit int) ([]*models.Inscricao, int, error) {
	args := m.Called(ctx, cpf, filter, offset, limit)
	return args.Get(0).([]*models.Inscricao), args.Int(1), args.Error(2)
}

func (m *MockInscricaoServiceForCourse) UpdateCertificate(ctx context.Context, cursoID int, inscricaoID uuid.UUID, certificateURL string) error {
	args := m.Called(ctx, cursoID, inscricaoID, certificateURL)
	return args.Error(0)
}

func (m *MockInscricaoServiceForCourse) UpdateInscricao(ctx context.Context, id uuid.UUID, cursoID int, updateData *models.InscricaoUpdateRequest) error {
	args := m.Called(ctx, id, cursoID, updateData)
	return args.Error(0)
}

func (m *MockInscricaoServiceForCourse) EnrichWithPersonalInfo(ctx context.Context, inscricao *models.Inscricao) {
	m.Called(ctx, inscricao)
}

func (m *MockInscricaoServiceForCourse) EnrichMultipleWithPersonalInfo(ctx context.Context, inscricoes []*models.Inscricao) {
	m.Called(ctx, inscricoes)
}

func (m *MockInscricaoServiceForCourse) ChangeSchedule(ctx context.Context, inscricaoID uuid.UUID, userCPF string, request *models.ScheduleChangeRequest) (*models.Inscricao, error) {
	args := m.Called(ctx, inscricaoID, userCPF, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Inscricao), args.Error(1)
}

// MockCursoRepositoryForCourseHandler is a minimal mock for CursoRepositoryInterface
type MockCursoRepositoryForCourseHandler struct {
	mock.Mock
}

func (m *MockCursoRepositoryForCourseHandler) Create(ctx context.Context, curso *models.Curso) (int, error) {
	args := m.Called(ctx, curso)
	return args.Int(0), args.Error(1)
}

func (m *MockCursoRepositoryForCourseHandler) GetByID(ctx context.Context, id int) (*models.Curso, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Curso), args.Error(1)
}

func (m *MockCursoRepositoryForCourseHandler) Update(ctx context.Context, curso *models.Curso) error {
	args := m.Called(ctx, curso)
	return args.Error(0)
}

func (m *MockCursoRepositoryForCourseHandler) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCursoRepositoryForCourseHandler) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Curso, int, error) {
	args := m.Called(ctx, filter, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.Curso), args.Int(1), args.Error(2)
}

func (m *MockCursoRepositoryForCourseHandler) CreateCustomFields(ctx context.Context, fields []models.CustomField) error {
	args := m.Called(ctx, fields)
	return args.Error(0)
}

func (m *MockCursoRepositoryForCourseHandler) CreateRemoteClass(ctx context.Context, remoteClass *models.RemoteClass) error {
	args := m.Called(ctx, remoteClass)
	return args.Error(0)
}

func (m *MockCursoRepositoryForCourseHandler) CreateLocationClasses(ctx context.Context, locations []models.LocationClass) error {
	args := m.Called(ctx, locations)
	return args.Error(0)
}

func (m *MockCursoRepositoryForCourseHandler) ValidateForEnrollment(ctx context.Context, cursoID int) (status string, enrollmentStart *time.Time, enrollmentEnd *time.Time, autoApprove bool, err error) {
	args := m.Called(ctx, cursoID)
	var es *time.Time
	var ee *time.Time
	if args.Get(1) != nil {
		es = args.Get(1).(*time.Time)
	}
	if args.Get(2) != nil {
		ee = args.Get(2).(*time.Time)
	}
	return args.String(0), es, ee, args.Bool(3), args.Error(4)
}

func (m *MockCursoRepositoryForCourseHandler) CountEnrollmentsByScheduleID(ctx context.Context, scheduleID uuid.UUID) (int64, error) {
	args := m.Called(ctx, scheduleID)
	return int64(args.Int(0)), args.Error(1)
}

func (m *MockCursoRepositoryForCourseHandler) CountEnrollmentsByScheduleIDs(ctx context.Context, scheduleIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	args := m.Called(ctx, scheduleIDs)
	if args.Get(0) == nil {
		return make(map[uuid.UUID]int64), args.Error(1)
	}
	return args.Get(0).(map[uuid.UUID]int64), args.Error(1)
}

func (m *MockCursoRepositoryForCourseHandler) GetCourseScheduleByID(ctx context.Context, scheduleID uuid.UUID) (*models.CourseSchedule, error) {
	args := m.Called(ctx, scheduleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CourseSchedule), args.Error(1)
}

func (m *MockCursoRepositoryForCourseHandler) GetRemoteScheduleByID(ctx context.Context, scheduleID uuid.UUID) (*models.RemoteSchedule, error) {
	args := m.Called(ctx, scheduleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RemoteSchedule), args.Error(1)
}

// Test basic handler initialization
func TestNewCourseHandler(t *testing.T) {
	handler := v1.NewCourseHandler(nil, nil, nil)
	assert.NotNil(t, handler)
}

// Test WithCache method
func TestCourseHandler_WithCache(t *testing.T) {
	handler := v1.NewCourseHandler(nil, nil, nil)
	result := handler.WithCache(nil)
	assert.NotNil(t, result)
}

// Test Create endpoint with invalid JSON
func TestCourseHandler_Create_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewCourseHandler(nil, nil, nil)
	r.POST("/api/v1/courses", h.Create)

	body := []byte(`{invalid json}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Dados inválidos")
}

// Test Create with empty body
func TestCourseHandler_Create_EmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewCourseHandler(nil, nil, nil)
	r.POST("/api/v1/courses", h.Create)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test CreateDraft endpoint with invalid JSON
func TestCourseHandler_CreateDraft_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewCourseHandler(nil, nil, nil)
	r.POST("/api/v1/courses/draft", h.CreateDraft)

	body := []byte(`{invalid}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses/draft", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test CreateDraft with empty body
func TestCourseHandler_CreateDraft_EmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewCourseHandler(nil, nil, nil)
	r.POST("/api/v1/courses/draft", h.CreateDraft)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses/draft", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test Update endpoint with invalid ID
func TestCourseHandler_Update_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewCourseHandler(nil, nil, nil)
	r.PUT("/api/v1/courses/:courseId", h.Update)

	body := []byte(`{"titulo": "Test"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/courses/invalid", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "ID do curso inválido")
}

// Test GetByID endpoint with invalid ID
func TestCourseHandler_GetByID_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewCourseHandler(nil, nil, nil)
	r.GET("/api/v1/courses/:courseId", h.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses/invalid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "ID do curso inválido")
}

// Test GetByID with special characters in ID
func TestCourseHandler_GetByID_SpecialCharsInID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewCourseHandler(nil, nil, nil)
	r.GET("/api/v1/courses/:courseId", h.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses/@#$", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test Delete endpoint with invalid ID
func TestCourseHandler_Delete_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewCourseHandler(nil, nil, nil)
	r.DELETE("/api/v1/courses/:courseId", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/courses/invalid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test Delete with float ID
func TestCourseHandler_Delete_FloatID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewCourseHandler(nil, nil, nil)
	r.DELETE("/api/v1/courses/:courseId", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/courses/123.456", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test JSON malformations for Create
func TestCourseHandler_Create_MalformedJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewCourseHandler(nil, nil, nil)
	r.POST("/api/v1/courses", h.Create)

	testCases := []struct{
		body string
		desc string
	}{
		{`{"titulo": "Test", "descricao": }`, "missing value"},
		{`{"titulo": ]`, "wrong bracket"},
		{`{titulo: "Test"}`, "unquoted key"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			body := []byte(tc.body)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/courses", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

// Test CreateDraft JSON malformations
func TestCourseHandler_CreateDraft_MalformedJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewCourseHandler(nil, nil, nil)
	r.POST("/api/v1/courses/draft", h.CreateDraft)

	testCases := []string{
		`{"titulo": ]`,
		`{invalid`,
	}

	for _, tc := range testCases {
		body := []byte(tc)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/courses/draft", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	}
}

// Test various invalid IDs for GetByID
func TestCourseHandler_GetByID_VariousIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewCourseHandler(nil, nil, nil)
	r.GET("/api/v1/courses/:courseId", h.GetByID)

	invalidIDs := []string{
		"abc123",
		"123.456",
	}

	for _, id := range invalidIDs {
		path := "/api/v1/courses/" + id
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	}
}

// Test Delete with various invalid IDs
func TestCourseHandler_Delete_VariousInvalidIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewCourseHandler(nil, nil, nil)
	r.DELETE("/api/v1/courses/:courseId", h.Delete)

	invalidIDs := []string{
		"abc",
		"12.34",
		"xyz",
	}

	for _, id := range invalidIDs {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/courses/"+id, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	}
}

// Test Update with various invalid IDs
func TestCourseHandler_Update_VariousInvalidIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewCourseHandler(nil, nil, nil)
	r.PUT("/api/v1/courses/:courseId", h.Update)

	invalidIDs := []string{
		"notanumber",
		"1.5",
		"@test",
	}

	for _, id := range invalidIDs {
		body := []byte(`{"titulo": "Test"}`)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/courses/"+id, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	}
}

// ========== Comprehensive Tests with Mocks ==========

// Test Create - Success
func TestCourseHandler_Create_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.POST("/api/v1/courses", handler.Create)

	curso := models.Curso{
		Titulo:    "Test Course",
		Descricao: "Test Description",
	}

	mockService.On("Create", mock.Anything, mock.MatchedBy(func(c *models.Curso) bool {
		return c.Status == models.StatusCursoOpened
	})).Return(1, nil)

	body, _ := json.Marshal(curso)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))
	assert.Equal(t, "Curso criado com sucesso", response["message"])

	mockService.AssertExpectations(t)
}

// Test Create - Validation Error
func TestCourseHandler_Create_ValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.POST("/api/v1/courses", handler.Create)

	curso := models.Curso{
		Titulo: "Test",
	}

	mockService.On("Create", mock.Anything, mock.Anything).Return(0, errors.New("erro de validação: título muito curto"))

	body, _ := json.Marshal(curso)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "título muito curto")

	mockService.AssertExpectations(t)
}

// Test Create - Database Error
func TestCourseHandler_Create_DatabaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.POST("/api/v1/courses", handler.Create)

	curso := models.Curso{
		Titulo:    "Test Course",
		Descricao: "Description",
	}

	mockService.On("Create", mock.Anything, mock.Anything).Return(0, errors.New("database connection error"))

	body, _ := json.Marshal(curso)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockService.AssertExpectations(t)
}

// Test CreateDraft - Success
func TestCourseHandler_CreateDraft_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.POST("/api/v1/courses/draft", handler.CreateDraft)

	curso := models.Curso{
		Titulo:    "Draft Course",
		Descricao: "Draft Description",
	}

	mockService.On("Create", mock.Anything, mock.MatchedBy(func(c *models.Curso) bool {
		return c.Status == models.StatusCursoDraft
	})).Return(2, nil)

	body, _ := json.Marshal(curso)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses/draft", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))
	assert.Equal(t, "Rascunho salvo com sucesso", response["message"])

	mockService.AssertExpectations(t)
}

// Test CreateDraft - Service Error
func TestCourseHandler_CreateDraft_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.POST("/api/v1/courses/draft", handler.CreateDraft)

	curso := models.Curso{
		Titulo: "Draft",
	}

	mockService.On("Create", mock.Anything, mock.Anything).Return(0, errors.New("service error"))

	body, _ := json.Marshal(curso)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses/draft", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockService.AssertExpectations(t)
}

// Test GetByID - Success
func TestCourseHandler_GetByID_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/v1/courses/:courseId", handler.GetByID)

	curso := &models.Curso{
		ID:        1,
		Titulo:    "Test Course",
		Descricao: "Test Description",
	}

	mockService.On("GetByID", mock.Anything, 1).Return(curso, nil)
	mockRepo.On("CountEnrollmentsByScheduleIDs", mock.Anything, mock.Anything).Return(make(map[uuid.UUID]int64), nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))

	mockService.AssertExpectations(t)
}

// Test GetByID - Not Found
func TestCourseHandler_GetByID_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/v1/courses/:courseId", handler.GetByID)

	mockService.On("GetByID", mock.Anything, 999).Return(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses/999", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Curso não encontrado")

	mockService.AssertExpectations(t)
}

// Test GetByID - Service Error
func TestCourseHandler_GetByID_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/v1/courses/:courseId", handler.GetByID)

	mockService.On("GetByID", mock.Anything, 1).Return(nil, errors.New("database error"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockService.AssertExpectations(t)
}

// Test Update - Success
func TestCourseHandler_Update_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.PUT("/api/v1/courses/:courseId", handler.Update)

	existingCurso := &models.Curso{
		ID:     1,
		Titulo: "Old Title",
		Status: models.StatusCursoOpened,
	}

	updateCurso := models.Curso{
		Titulo:    "New Title",
		Descricao: "New Description",
	}

	mockService.On("GetByID", mock.Anything, 1).Return(existingCurso, nil)
	mockService.On("Update", mock.Anything, mock.Anything).Return(nil)

	body, _ := json.Marshal(updateCurso)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/courses/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))
	assert.Equal(t, "Curso atualizado com sucesso", response["message"])

	mockService.AssertExpectations(t)
}

// Test Update - Not Found
func TestCourseHandler_Update_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.PUT("/api/v1/courses/:courseId", handler.Update)

	updateCurso := models.Curso{
		Titulo: "New Title",
	}

	mockService.On("GetByID", mock.Anything, 999).Return(nil, nil)

	body, _ := json.Marshal(updateCurso)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/courses/999", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Curso não encontrado")

	mockService.AssertExpectations(t)
}

// Test Update - Publish Draft
func TestCourseHandler_Update_PublishDraft(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.PUT("/api/v1/courses/:courseId", handler.Update)

	existingCurso := &models.Curso{
		ID:     1,
		Titulo: "Draft Course",
		Status: models.StatusCursoDraft,
	}

	updateCurso := models.Curso{
		Titulo: "Published Course",
		Status: models.StatusCursoOpened,
	}

	mockService.On("GetByID", mock.Anything, 1).Return(existingCurso, nil)
	mockService.On("Update", mock.Anything, mock.Anything).Return(nil)

	body, _ := json.Marshal(updateCurso)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/courses/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Curso publicado com sucesso", response["message"])

	mockService.AssertExpectations(t)
}

// Test Update - Invalid JSON
func TestCourseHandler_Update_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.PUT("/api/v1/courses/:courseId", handler.Update)

	existingCurso := &models.Curso{
		ID:     1,
		Titulo: "Test",
	}

	mockService.On("GetByID", mock.Anything, 1).Return(existingCurso, nil)

	body := []byte(`{invalid json}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/courses/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Dados inválidos")

	mockService.AssertExpectations(t)
}

// Test Delete - Success
func TestCourseHandler_Delete_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.DELETE("/api/v1/courses/:courseId", handler.Delete)

	curso := &models.Curso{
		ID:     1,
		Titulo: "Test Course",
	}

	mockService.On("GetByID", mock.Anything, 1).Return(curso, nil)
	mockService.On("Delete", mock.Anything, 1).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/courses/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))
	assert.Equal(t, "Curso excluído com sucesso", response["message"])

	mockService.AssertExpectations(t)
}

// Test Delete - Not Found
func TestCourseHandler_Delete_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.DELETE("/api/v1/courses/:courseId", handler.Delete)

	mockService.On("GetByID", mock.Anything, 999).Return(nil, nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/courses/999", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Curso não encontrado")

	mockService.AssertExpectations(t)
}

// Test Delete - Service Error
func TestCourseHandler_Delete_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.DELETE("/api/v1/courses/:courseId", handler.Delete)

	curso := &models.Curso{ID: 1}

	mockService.On("GetByID", mock.Anything, 1).Return(curso, nil)
	mockService.On("Delete", mock.Anything, 1).Return(errors.New("delete error"))

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/courses/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockService.AssertExpectations(t)
}

// Test List - Success
func TestCourseHandler_List_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/v1/courses", handler.List)

	cursos := []*models.Curso{
		{ID: 1, Titulo: "Course 1"},
		{ID: 2, Titulo: "Course 2"},
	}

	mockService.On("List", mock.Anything, mock.Anything, 1, 10).Return(cursos, 2, nil)
	mockRepo.On("CountEnrollmentsByScheduleIDs", mock.Anything, mock.Anything).Return(make(map[uuid.UUID]int64), nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses?page=1&limit=10", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))

	mockService.AssertExpectations(t)
}

// Test List - Service Error
func TestCourseHandler_List_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/v1/courses", handler.List)

	mockService.On("List", mock.Anything, mock.Anything, 1, 10).Return(nil, 0, errors.New("database error"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockService.AssertExpectations(t)
}

// Test ListDrafts - Success
func TestCourseHandler_ListDrafts_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/v1/courses/drafts", handler.ListDrafts)

	drafts := []*models.Curso{
		{ID: 1, Titulo: "Draft 1", Status: models.StatusCursoDraft},
	}

	mockService.On("List", mock.Anything, mock.Anything, 1, 10).Return(drafts, 1, nil)
	mockRepo.On("CountEnrollmentsByScheduleIDs", mock.Anything, mock.Anything).Return(make(map[uuid.UUID]int64), nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses/drafts", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))

	mockService.AssertExpectations(t)
}

// Test ListByUser - Success
func TestCourseHandler_ListByUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/v1/users/:userId/courses", handler.ListByUser)

	cursos := []*models.Curso{
		{ID: 1, Titulo: "User Course 1", OrgaoID: "user123"},
	}

	mockService.On("List", mock.Anything, mock.Anything, 1, 10).Return(cursos, 1, nil)
	mockRepo.On("CountEnrollmentsByScheduleIDs", mock.Anything, mock.Anything).Return(make(map[uuid.UUID]int64), nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/user123/courses", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))

	mockService.AssertExpectations(t)
}

// Test List with filters
func TestCourseHandler_List_WithFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/v1/courses", handler.List)

	cursos := []*models.Curso{
		{ID: 1, Titulo: "Filtered Course", Status: models.StatusCursoOpened},
	}

	mockService.On("List", mock.Anything, mock.MatchedBy(func(filter map[string]interface{}) bool {
		// Handler automatically adds "status NOT": draft to exclude drafts
		_, hasStatusNot := filter["status NOT"]
		statusVal, hasStatus := filter["status"]
		return hasStatus && hasStatusNot && statusVal == "opened"
	}), 1, 10).Return(cursos, 1, nil)
	mockRepo.On("CountEnrollmentsByScheduleIDs", mock.Anything, mock.Anything).Return(make(map[uuid.UUID]int64), nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses?status=opened&page=1&limit=10", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}
