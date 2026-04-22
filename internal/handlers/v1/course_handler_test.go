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

func (m *MockCursoService) SendToReview(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCursoService) Approve(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCursoService) RequestChanges(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCursoService) RequestDeletion(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
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

func (m *MockCursoRepositoryForCourseHandler) UpdateStatus(ctx context.Context, id int, status models.StatusCurso) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
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

	testCases := []struct {
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

// Test calculateRemainingVacancies with location schedules
func TestCourseHandler_CalculateRemainingVacancies_LocationSchedules(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/v1/courses/:courseId", handler.GetByID)

	schedID1 := uuid.New()
	schedID2 := uuid.New()

	curso := &models.Curso{
		ID:     1,
		Titulo: "Course with schedules",
		LocationClasses: []models.LocationClass{
			{
				Schedules: []models.CourseSchedule{
					{ID: schedID1, Vacancies: 10},
					{ID: schedID2, Vacancies: 5},
				},
			},
		},
	}

	enrollmentCounts := map[uuid.UUID]int64{
		schedID1: 7,
		schedID2: 3,
	}

	mockService.On("GetByID", mock.Anything, 1).Return(curso, nil)
	mockRepo.On("CountEnrollmentsByScheduleIDs", mock.Anything, mock.MatchedBy(func(ids []uuid.UUID) bool {
		return len(ids) == 2
	})).Return(enrollmentCounts, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	mockService.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

// Test calculateRemainingVacancies with remote schedules
func TestCourseHandler_CalculateRemainingVacancies_RemoteSchedules(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/v1/courses/:courseId", handler.GetByID)

	schedID := uuid.New()

	curso := &models.Curso{
		ID:     1,
		Titulo: "Remote Course",
		RemoteClass: &models.RemoteClass{
			Schedules: []models.RemoteSchedule{
				{ID: schedID, Vacancies: 20},
			},
		},
	}

	enrollmentCounts := map[uuid.UUID]int64{
		schedID: 15,
	}

	mockService.On("GetByID", mock.Anything, 1).Return(curso, nil)
	mockRepo.On("CountEnrollmentsByScheduleIDs", mock.Anything, mock.MatchedBy(func(ids []uuid.UUID) bool {
		return len(ids) == 1
	})).Return(enrollmentCounts, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

// Test calculateRemainingVacancies with overfilled schedules
func TestCourseHandler_CalculateRemainingVacancies_Overfilled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/v1/courses/:courseId", handler.GetByID)

	schedID := uuid.New()

	curso := &models.Curso{
		ID:     1,
		Titulo: "Overfilled Course",
		LocationClasses: []models.LocationClass{
			{
				Schedules: []models.CourseSchedule{
					{ID: schedID, Vacancies: 5},
				},
			},
		},
	}

	enrollmentCounts := map[uuid.UUID]int64{
		schedID: 10,
	}

	mockService.On("GetByID", mock.Anything, 1).Return(curso, nil)
	mockRepo.On("CountEnrollmentsByScheduleIDs", mock.Anything, mock.Anything).Return(enrollmentCounts, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

// Test calculateRemainingVacancies with no schedules
func TestCourseHandler_CalculateRemainingVacancies_NoSchedules(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/v1/courses/:courseId", handler.GetByID)

	curso := &models.Curso{
		ID:              1,
		Titulo:          "No Schedules",
		LocationClasses: []models.LocationClass{},
		RemoteClass:     nil,
	}

	mockService.On("GetByID", mock.Anything, 1).Return(curso, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

// Test calculateRemainingVacancies repository error
func TestCourseHandler_CalculateRemainingVacancies_RepositoryError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/v1/courses/:courseId", handler.GetByID)

	schedID := uuid.New()

	curso := &models.Curso{
		ID:     1,
		Titulo: "Course",
		LocationClasses: []models.LocationClass{
			{
				Schedules: []models.CourseSchedule{
					{ID: schedID, Vacancies: 10},
				},
			},
		},
	}

	mockService.On("GetByID", mock.Anything, 1).Return(curso, nil)
	mockRepo.On("CountEnrollmentsByScheduleIDs", mock.Anything, mock.Anything).Return(nil, errors.New("repository error"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Erro ao calcular vagas restantes")

	mockService.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

// Test ListDrafts with service error
func TestCourseHandler_ListDrafts_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/v1/courses/drafts", handler.ListDrafts)

	mockService.On("List", mock.Anything, mock.Anything, 1, 10).Return(nil, 0, errors.New("database error"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses/drafts", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Erro ao listar rascunhos")

	mockService.AssertExpectations(t)
}

// Test ListDrafts with calculateRemainingVacancies error
func TestCourseHandler_ListDrafts_CalculateVacanciesError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/v1/courses/drafts", handler.ListDrafts)

	schedID := uuid.New()
	drafts := []*models.Curso{
		{
			ID:     1,
			Titulo: "Draft",
			Status: models.StatusCursoDraft,
			LocationClasses: []models.LocationClass{
				{
					Schedules: []models.CourseSchedule{
						{ID: schedID, Vacancies: 10},
					},
				},
			},
		},
	}

	mockService.On("List", mock.Anything, mock.Anything, 1, 10).Return(drafts, 1, nil)
	mockRepo.On("CountEnrollmentsByScheduleIDs", mock.Anything, mock.Anything).Return(nil, errors.New("repo error"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses/drafts", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Erro ao calcular vagas restantes")

	mockService.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

// Test ListDrafts with filters
func TestCourseHandler_ListDrafts_WithFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/v1/courses/drafts", handler.ListDrafts)

	drafts := []*models.Curso{
		{ID: 1, Titulo: "Draft 1", Status: models.StatusCursoDraft, LocationClasses: []models.LocationClass{}},
	}

	mockService.On("List", mock.Anything, mock.MatchedBy(func(filter map[string]interface{}) bool {
		status, hasStatus := filter["status"]
		org, hasOrg := filter["organization"]
		search, hasSearch := filter["title ILIKE"]
		return hasStatus && status == models.StatusCursoDraft &&
			hasOrg && org == "test-org" &&
			hasSearch && search == "%test%"
	}), 1, 10).Return(drafts, 1, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses/drafts?organization=test-org&search=test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

// Test ListDrafts with pagination
func TestCourseHandler_ListDrafts_WithPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/v1/courses/drafts", handler.ListDrafts)

	drafts := []*models.Curso{
		{ID: 21, Titulo: "Draft 21", Status: models.StatusCursoDraft, LocationClasses: []models.LocationClass{}},
	}

	mockService.On("List", mock.Anything, mock.Anything, 3, 20).Return(drafts, 50, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses/drafts?page=3&limit=20", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	data := response["data"].(map[string]interface{})
	pagination := data["pagination"].(map[string]interface{})
	assert.Equal(t, float64(3), pagination["page"])
	assert.Equal(t, float64(20), pagination["limit"])
	assert.Equal(t, float64(50), pagination["total"])

	mockService.AssertExpectations(t)
}

// Test List with calculateRemainingVacancies error
func TestCourseHandler_List_CalculateVacanciesError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/v1/courses", handler.List)

	schedID := uuid.New()
	cursos := []*models.Curso{
		{
			ID:     1,
			Titulo: "Course",
			LocationClasses: []models.LocationClass{
				{
					Schedules: []models.CourseSchedule{
						{ID: schedID, Vacancies: 10},
					},
				},
			},
		},
	}

	mockService.On("List", mock.Anything, mock.Anything, 1, 10).Return(cursos, 1, nil)
	mockRepo.On("CountEnrollmentsByScheduleIDs", mock.Anything, mock.Anything).Return(nil, errors.New("repo error"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Erro ao calcular vagas restantes")

	mockService.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

// Test ListByUser with service error
func TestCourseHandler_ListByUser_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/v1/users/:userId/courses", handler.ListByUser)

	mockService.On("List", mock.Anything, mock.Anything, 1, 10).Return(nil, 0, errors.New("database error"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/user123/courses", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Erro ao listar cursos do usuário")

	mockService.AssertExpectations(t)
}

// Test ListByUser with calculateRemainingVacancies error
func TestCourseHandler_ListByUser_CalculateVacanciesError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/v1/users/:userId/courses", handler.ListByUser)

	schedID := uuid.New()
	cursos := []*models.Curso{
		{
			ID:      1,
			Titulo:  "User Course",
			OrgaoID: "user123",
			LocationClasses: []models.LocationClass{
				{
					Schedules: []models.CourseSchedule{
						{ID: schedID, Vacancies: 10},
					},
				},
			},
		},
	}

	mockService.On("List", mock.Anything, mock.Anything, 1, 10).Return(cursos, 1, nil)
	mockRepo.On("CountEnrollmentsByScheduleIDs", mock.Anything, mock.Anything).Return(nil, errors.New("repo error"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/user123/courses", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Erro ao calcular vagas restantes")

	mockService.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

// Test Update with GetByID error
func TestCourseHandler_Update_GetByIDError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.PUT("/api/v1/courses/:courseId", handler.Update)

	mockService.On("GetByID", mock.Anything, 1).Return(nil, errors.New("database error"))

	body := []byte(`{"titulo":"Updated"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/courses/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Erro ao buscar curso")

	mockService.AssertExpectations(t)
}

// Test Update with validation error
func TestCourseHandler_Update_ValidationError(t *testing.T) {
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

	mockService.On("GetByID", mock.Anything, 1).Return(existingCurso, nil)
	mockService.On("Update", mock.Anything, mock.Anything).Return(errors.New("erro de validação: campo inválido"))

	body := []byte(`{"titulo":"New"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/courses/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "campo inválido")

	mockService.AssertExpectations(t)
}

// Test Delete with GetByID error
func TestCourseHandler_Delete_GetByIDError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.DELETE("/api/v1/courses/:courseId", handler.Delete)

	mockService.On("GetByID", mock.Anything, 1).Return(nil, errors.New("database error"))

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/courses/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Erro ao buscar curso")

	mockService.AssertExpectations(t)
}

// ========== Additional Coverage Tests ==========

// Test calculateRemainingVacanciesForCourses with empty course list
func TestCourseHandler_CalculateRemainingVacanciesForCourses_EmptyList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/v1/courses", handler.List)

	// Empty list of courses
	mockService.On("List", mock.Anything, mock.Anything, 1, 10).Return([]*models.Curso{}, 0, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	data := response["data"].(map[string]interface{})
	courses := data["courses"].([]interface{})
	assert.Equal(t, 0, len(courses))

	mockService.AssertExpectations(t)
}

// Test calculateRemainingVacanciesForCourses with courses missing schedules
func TestCourseHandler_CalculateRemainingVacanciesForCourses_MissingSchedules(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/v1/courses", handler.List)

	// Courses with no schedules
	cursos := []*models.Curso{
		{ID: 1, Titulo: "Course 1", LocationClasses: []models.LocationClass{}},
		{ID: 2, Titulo: "Course 2", LocationClasses: []models.LocationClass{}, RemoteClass: nil},
	}

	mockService.On("List", mock.Anything, mock.Anything, 1, 10).Return(cursos, 2, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

// Test calculateRemainingVacanciesForCourses with mixed schedule types
func TestCourseHandler_CalculateRemainingVacanciesForCourses_MixedScheduleTypes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/v1/courses", handler.List)

	schedID1 := uuid.New()
	schedID2 := uuid.New()
	schedID3 := uuid.New()

	cursos := []*models.Curso{
		{
			ID:     1,
			Titulo: "Mixed Course 1",
			LocationClasses: []models.LocationClass{
				{
					Schedules: []models.CourseSchedule{
						{ID: schedID1, Vacancies: 20},
					},
				},
			},
			RemoteClass: &models.RemoteClass{
				Schedules: []models.RemoteSchedule{
					{ID: schedID2, Vacancies: 50},
				},
			},
		},
		{
			ID:     2,
			Titulo: "Location Only",
			LocationClasses: []models.LocationClass{
				{
					Schedules: []models.CourseSchedule{
						{ID: schedID3, Vacancies: 15},
					},
				},
			},
		},
	}

	enrollmentCounts := map[uuid.UUID]int64{
		schedID1: 5,
		schedID2: 30,
		schedID3: 10,
	}

	mockService.On("List", mock.Anything, mock.Anything, 1, 10).Return(cursos, 2, nil)
	mockRepo.On("CountEnrollmentsByScheduleIDs", mock.Anything, mock.MatchedBy(func(ids []uuid.UUID) bool {
		return len(ids) == 3
	})).Return(enrollmentCounts, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

// Test Update with empty body
func TestCourseHandler_Update_EmptyBody(t *testing.T) {
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
	}

	mockService.On("GetByID", mock.Anything, 1).Return(existingCurso, nil)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/courses/1", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	mockService.AssertExpectations(t)
}

// Test Update with database error from service
func TestCourseHandler_Update_DatabaseError(t *testing.T) {
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

	mockService.On("GetByID", mock.Anything, 1).Return(existingCurso, nil)
	mockService.On("Update", mock.Anything, mock.Anything).Return(errors.New("database connection lost"))

	body := []byte(`{"titulo":"Updated Title"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/courses/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockService.AssertExpectations(t)
}

// Test List with pagination boundary page=0
func TestCourseHandler_List_PaginationPageZero(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/v1/courses", handler.List)

	cursos := []*models.Curso{
		{ID: 1, Titulo: "Course 1"},
	}

	// Page 0 should default to page 1
	mockService.On("List", mock.Anything, mock.Anything, 1, 10).Return(cursos, 1, nil)
	mockRepo.On("CountEnrollmentsByScheduleIDs", mock.Anything, mock.Anything).Return(make(map[uuid.UUID]int64), nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses?page=0", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	data := response["data"].(map[string]interface{})
	pagination := data["pagination"].(map[string]interface{})
	assert.Equal(t, float64(1), pagination["page"])

	mockService.AssertExpectations(t)
}

// Test List with pagination boundary limit=0
func TestCourseHandler_List_PaginationLimitZero(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/v1/courses", handler.List)

	cursos := []*models.Curso{
		{ID: 1, Titulo: "Course 1"},
	}

	// Limit 0 should default to limit 10
	mockService.On("List", mock.Anything, mock.Anything, 1, 10).Return(cursos, 1, nil)
	mockRepo.On("CountEnrollmentsByScheduleIDs", mock.Anything, mock.Anything).Return(make(map[uuid.UUID]int64), nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses?limit=0", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	data := response["data"].(map[string]interface{})
	pagination := data["pagination"].(map[string]interface{})
	assert.Equal(t, float64(10), pagination["limit"])

	mockService.AssertExpectations(t)
}

// Test List with pagination limit > max
func TestCourseHandler_List_PaginationLimitExceedsMax(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/v1/courses", handler.List)

	cursos := []*models.Curso{}

	// Limit > 1000 should default to limit 10
	mockService.On("List", mock.Anything, mock.Anything, 1, 10).Return(cursos, 0, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses?limit=2000", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

// Test List with multiple filters combination
func TestCourseHandler_List_MultipleFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/v1/courses", handler.List)

	cursos := []*models.Curso{
		{ID: 1, Titulo: "Filtered Course"},
	}

	mockService.On("List", mock.Anything, mock.MatchedBy(func(filter map[string]interface{}) bool {
		return filter["modalidade"] == "presencial" &&
			filter["organization"] == "test-org" &&
			filter["status"] == "opened" &&
			filter["categoria_id"] == 5 &&
			filter["acessibilidade_id"] == 3 &&
			filter["neighborhood_zone"] == "norte"
	}), 1, 10).Return(cursos, 1, nil)
	mockRepo.On("CountEnrollmentsByScheduleIDs", mock.Anything, mock.Anything).Return(make(map[uuid.UUID]int64), nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses?modalidade=presencial&organization=test-org&status=opened&categoria_id=5&acessibilidade_id=3&neighborhood_zone=norte", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

// Test ListDrafts with pagination boundary page=0
func TestCourseHandler_ListDrafts_PaginationPageZero(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/v1/courses/drafts", handler.ListDrafts)

	drafts := []*models.Curso{
		{ID: 1, Titulo: "Draft 1", Status: models.StatusCursoDraft, LocationClasses: []models.LocationClass{}},
	}

	mockService.On("List", mock.Anything, mock.Anything, 1, 10).Return(drafts, 1, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses/drafts?page=0", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	data := response["data"].(map[string]interface{})
	pagination := data["pagination"].(map[string]interface{})
	assert.Equal(t, float64(1), pagination["page"])

	mockService.AssertExpectations(t)
}

// Test ListDrafts with pagination limit=0
func TestCourseHandler_ListDrafts_PaginationLimitZero(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/v1/courses/drafts", handler.ListDrafts)

	drafts := []*models.Curso{}

	mockService.On("List", mock.Anything, mock.Anything, 1, 10).Return(drafts, 0, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses/drafts?limit=0", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

// Test ListDrafts with empty drafts
func TestCourseHandler_ListDrafts_EmptyDrafts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/v1/courses/drafts", handler.ListDrafts)

	mockService.On("List", mock.Anything, mock.Anything, 1, 10).Return([]*models.Curso{}, 0, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses/drafts", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	data := response["data"].(map[string]interface{})
	drafts := data["drafts"].([]interface{})
	assert.Equal(t, 0, len(drafts))

	mockService.AssertExpectations(t)
}

// Test ListByUser with empty user ID
func TestCourseHandler_ListByUser_EmptyUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/v1/users/:userId/courses", handler.ListByUser)

	mockService.On("List", mock.Anything, mock.MatchedBy(func(filter map[string]interface{}) bool {
		return filter["orgao_id"] == ""
	}), 1, 10).Return([]*models.Curso{}, 0, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users//courses", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

// Test ListByUser with no courses
func TestCourseHandler_ListByUser_NoCourses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/v1/users/:userId/courses", handler.ListByUser)

	mockService.On("List", mock.Anything, mock.MatchedBy(func(filter map[string]interface{}) bool {
		return filter["orgao_id"] == "user999"
	}), 1, 10).Return([]*models.Curso{}, 0, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/user999/courses", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	data := response["data"].(map[string]interface{})
	courses := data["courses"].([]interface{})
	assert.Equal(t, 0, len(courses))

	mockService.AssertExpectations(t)
}

// Test ListByUser with multiple courses
func TestCourseHandler_ListByUser_MultipleCourses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/v1/users/:userId/courses", handler.ListByUser)

	cursos := []*models.Curso{
		{ID: 1, Titulo: "User Course 1", OrgaoID: "user123"},
		{ID: 2, Titulo: "User Course 2", OrgaoID: "user123"},
		{ID: 3, Titulo: "User Course 3", OrgaoID: "user123"},
	}

	mockService.On("List", mock.Anything, mock.Anything, 1, 10).Return(cursos, 3, nil)
	mockRepo.On("CountEnrollmentsByScheduleIDs", mock.Anything, mock.Anything).Return(make(map[uuid.UUID]int64), nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/user123/courses", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	data := response["data"].(map[string]interface{})
	courses := data["courses"].([]interface{})
	assert.Equal(t, 3, len(courses))

	mockService.AssertExpectations(t)
}

// Test ListByUser with pagination boundaries
func TestCourseHandler_ListByUser_PaginationBoundaries(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/v1/users/:userId/courses", handler.ListByUser)

	cursos := []*models.Curso{}

	// Page=-1 should default to 1, limit=2000 should default to 10
	mockService.On("List", mock.Anything, mock.Anything, 1, 10).Return(cursos, 0, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/user123/courses?page=-1&limit=2000", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

// Test ListByUser with filters
func TestCourseHandler_ListByUser_WithFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/v1/users/:userId/courses", handler.ListByUser)

	cursos := []*models.Curso{
		{ID: 1, Titulo: "Filtered User Course", OrgaoID: "user123", Status: models.StatusCursoOpened},
	}

	mockService.On("List", mock.Anything, mock.MatchedBy(func(filter map[string]interface{}) bool {
		return filter["orgao_id"] == "user123" &&
			filter["status"] == "opened" &&
			filter["modalidade"] == "remoto" &&
			filter["categoria_id"] == 2
	}), 1, 10).Return(cursos, 1, nil)
	mockRepo.On("CountEnrollmentsByScheduleIDs", mock.Anything, mock.Anything).Return(make(map[uuid.UUID]int64), nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/user123/courses?status=opened&modalidade=remoto&categoria_id=2", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

// Test calculateRemainingVacancies with multiple location classes
func TestCourseHandler_CalculateRemainingVacancies_MultipleLocationClasses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/v1/courses/:courseId", handler.GetByID)

	schedID1 := uuid.New()
	schedID2 := uuid.New()
	schedID3 := uuid.New()

	curso := &models.Curso{
		ID:     1,
		Titulo: "Multi-location Course",
		LocationClasses: []models.LocationClass{
			{
				Schedules: []models.CourseSchedule{
					{ID: schedID1, Vacancies: 10},
					{ID: schedID2, Vacancies: 15},
				},
			},
			{
				Schedules: []models.CourseSchedule{
					{ID: schedID3, Vacancies: 20},
				},
			},
		},
	}

	enrollmentCounts := map[uuid.UUID]int64{
		schedID1: 5,
		schedID2: 12,
		schedID3: 18,
	}

	mockService.On("GetByID", mock.Anything, 1).Return(curso, nil)
	mockRepo.On("CountEnrollmentsByScheduleIDs", mock.Anything, mock.MatchedBy(func(ids []uuid.UUID) bool {
		return len(ids) == 3
	})).Return(enrollmentCounts, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

// Test Update with partial update preserving IsVisible
func TestCourseHandler_Update_PreserveIsVisible(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.PUT("/api/v1/courses/:courseId", handler.Update)

	isVisibleTrue := true
	existingCurso := &models.Curso{
		ID:        1,
		Titulo:    "Old Title",
		Status:    models.StatusCursoOpened,
		IsVisible: &isVisibleTrue,
	}

	// Request doesn't include is_visible, should preserve existing value
	updateCurso := models.Curso{
		Titulo: "New Title",
	}

	mockService.On("GetByID", mock.Anything, 1).Return(existingCurso, nil)
	mockService.On("Update", mock.Anything, mock.MatchedBy(func(c *models.Curso) bool {
		return c.IsVisible != nil && *c.IsVisible == true
	})).Return(nil)

	body, _ := json.Marshal(updateCurso)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/courses/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

// Test Update with partial update preserving AutoApproveEnrollments
func TestCourseHandler_Update_PreserveAutoApprove(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.PUT("/api/v1/courses/:courseId", handler.Update)

	autoApproveTrue := true
	existingCurso := &models.Curso{
		ID:                     1,
		Titulo:                 "Old Title",
		Status:                 models.StatusCursoOpened,
		AutoApproveEnrollments: &autoApproveTrue,
	}

	updateCurso := models.Curso{
		Titulo: "New Title",
	}

	mockService.On("GetByID", mock.Anything, 1).Return(existingCurso, nil)
	mockService.On("Update", mock.Anything, mock.MatchedBy(func(c *models.Curso) bool {
		return c.AutoApproveEnrollments != nil && *c.AutoApproveEnrollments == true
	})).Return(nil)

	body, _ := json.Marshal(updateCurso)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/courses/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

// Test List with search filter
func TestCourseHandler_List_SearchFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/v1/courses", handler.List)

	cursos := []*models.Curso{
		{ID: 1, Titulo: "Test Course"},
	}

	mockService.On("List", mock.Anything, mock.MatchedBy(func(filter map[string]interface{}) bool {
		search, ok := filter["title ILIKE"]
		return ok && search == "%test%"
	}), 1, 10).Return(cursos, 1, nil)
	mockRepo.On("CountEnrollmentsByScheduleIDs", mock.Anything, mock.Anything).Return(make(map[uuid.UUID]int64), nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses?search=test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

// Test ListDrafts with multiple filters
func TestCourseHandler_ListDrafts_MultipleFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/v1/courses/drafts", handler.ListDrafts)

	drafts := []*models.Curso{
		{ID: 1, Titulo: "Draft", Status: models.StatusCursoDraft, LocationClasses: []models.LocationClass{}},
	}

	mockService.On("List", mock.Anything, mock.MatchedBy(func(filter map[string]interface{}) bool {
		return filter["status"] == models.StatusCursoDraft &&
			filter["modalidade"] == "hibrido" &&
			filter["categoria_id"] == 7 &&
			filter["acessibilidade_id"] == 2 &&
			filter["neighborhood_zone"] == "sul"
	}), 1, 10).Return(drafts, 1, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses/drafts?modalidade=hibrido&categoria_id=7&acessibilidade_id=2&neighborhood_zone=sul", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

// Test List with empty result set
func TestCourseHandler_List_EmptyResultSet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/v1/courses", handler.List)

	mockService.On("List", mock.Anything, mock.Anything, 1, 10).Return([]*models.Curso{}, 0, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses?status=nonexistent", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	data := response["data"].(map[string]interface{})
	courses := data["courses"].([]interface{})
	assert.Equal(t, 0, len(courses))

	mockService.AssertExpectations(t)
}

// Test Update preserving schedule AcceptingEnrollments
func TestCourseHandler_Update_PreserveScheduleAcceptingEnrollments(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.PUT("/api/v1/courses/:courseId", handler.Update)

	schedID1 := uuid.MustParse("12345678-1234-1234-1234-123456789012")
	schedID2 := uuid.MustParse("87654321-4321-4321-4321-210987654321")
	acceptingTrue := true
	acceptingFalse := false

	existingCurso := &models.Curso{
		ID:     1,
		Titulo: "Existing Course",
		Status: models.StatusCursoOpened,
		LocationClasses: []models.LocationClass{
			{
				Schedules: []models.CourseSchedule{
					{ID: schedID1, AcceptingEnrollments: &acceptingTrue},
				},
			},
		},
		RemoteClass: &models.RemoteClass{
			Schedules: []models.RemoteSchedule{
				{ID: schedID2, AcceptingEnrollments: &acceptingFalse},
			},
		},
	}

	// Update request doesn't include AcceptingEnrollments, should preserve
	updateData := map[string]interface{}{
		"titulo": "Updated Course",
		"locations": []map[string]interface{}{
			{
				"schedules": []map[string]interface{}{
					{
						"id":        schedID1.String(),
						"vacancies": 10,
					},
				},
			},
		},
		"remote_class": map[string]interface{}{
			"schedules": []map[string]interface{}{
				{
					"id":        schedID2.String(),
					"vacancies": 20,
				},
			},
		},
	}

	mockService.On("GetByID", mock.Anything, 1).Return(existingCurso, nil)
	mockService.On("Update", mock.Anything, mock.MatchedBy(func(c *models.Curso) bool {
		// Verify AcceptingEnrollments is preserved
		if len(c.LocationClasses) > 0 && len(c.LocationClasses[0].Schedules) > 0 {
			if c.LocationClasses[0].Schedules[0].AcceptingEnrollments == nil {
				return false
			}
			if !*c.LocationClasses[0].Schedules[0].AcceptingEnrollments {
				return false
			}
		}
		if c.RemoteClass != nil && len(c.RemoteClass.Schedules) > 0 {
			if c.RemoteClass.Schedules[0].AcceptingEnrollments == nil {
				return false
			}
			if *c.RemoteClass.Schedules[0].AcceptingEnrollments {
				return false
			}
		}
		return true
	})).Return(nil)

	body, _ := json.Marshal(updateData)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/courses/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

// Test Update with zero UUID should not preserve AcceptingEnrollments
func TestCourseHandler_Update_ZeroUUIDNoPreserve(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.PUT("/api/v1/courses/:courseId", handler.Update)

	existingCurso := &models.Curso{
		ID:     1,
		Titulo: "Existing Course",
		Status: models.StatusCursoOpened,
	}

	zeroUUID := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	updateData := map[string]interface{}{
		"titulo": "Updated Course",
		"locations": []map[string]interface{}{
			{
				"schedules": []map[string]interface{}{
					{
						"id":        zeroUUID.String(),
						"vacancies": 10,
					},
				},
			},
		},
	}

	mockService.On("GetByID", mock.Anything, 1).Return(existingCurso, nil)
	mockService.On("Update", mock.Anything, mock.Anything).Return(nil)

	body, _ := json.Marshal(updateData)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/courses/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

// Test Create with database error non-validation
func TestCourseHandler_Create_NonValidationDatabaseError(t *testing.T) {
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

	// Database error without "erro de validação" prefix
	mockService.On("Create", mock.Anything, mock.Anything).Return(0, errors.New("unique constraint violation"))

	body, _ := json.Marshal(curso)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Should be handled as database error
	assert.NotEqual(t, http.StatusCreated, w.Code)

	mockService.AssertExpectations(t)
}

// Test CreateDraft with non-validation database error
func TestCourseHandler_CreateDraft_NonValidationDatabaseError(t *testing.T) {
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

	mockService.On("Create", mock.Anything, mock.Anything).Return(0, errors.New("unique constraint violation"))

	body, _ := json.Marshal(curso)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses/draft", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusCreated, w.Code)

	mockService.AssertExpectations(t)
}

// Test Update with non-validation database error
func TestCourseHandler_Update_NonValidationDatabaseError(t *testing.T) {
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

	mockService.On("GetByID", mock.Anything, 1).Return(existingCurso, nil)
	mockService.On("Update", mock.Anything, mock.Anything).Return(errors.New("unique constraint violation"))

	body := []byte(`{"titulo":"New Title"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/courses/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

// Test CreateDraft with validation error
func TestCourseHandler_CreateDraft_ValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.POST("/api/v1/courses/draft", handler.CreateDraft)

	curso := models.Curso{
		Titulo: "D",
	}

	mockService.On("Create", mock.Anything, mock.Anything).Return(0, errors.New("erro de validação: título muito curto"))

	body, _ := json.Marshal(curso)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses/draft", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "título muito curto")

	mockService.AssertExpectations(t)
}

func setupTransitionHandler(t *testing.T) (*MockCursoService, *MockCursoRepositoryForCourseHandler, *v1.CourseHandler) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockRepo := new(MockCursoRepositoryForCourseHandler)
	h := v1.NewCourseHandler(mockService, nil, mockRepo)
	return mockService, mockRepo, h
}

func TestCourseHandler_SendToReview(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockService, _, h := setupTransitionHandler(t)
		mockService.On("SendToReview", mock.Anything, 1).Return(nil)
		r := gin.New()
		r.PUT("/courses/:courseId/send-to-review", h.SendToReview)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("PUT", "/courses/1/send-to-review", nil))
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "revisão")
		mockService.AssertExpectations(t)
	})

	t.Run("invalid_id", func(t *testing.T) {
		_, _, h := setupTransitionHandler(t)
		r := gin.New()
		r.PUT("/courses/:courseId/send-to-review", h.SendToReview)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("PUT", "/courses/abc/send-to-review", nil))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("conflict_invalid_state", func(t *testing.T) {
		mockService, _, h := setupTransitionHandler(t)
		mockService.On("SendToReview", mock.Anything, 1).Return(errors.New("curso não está em estado válido"))
		r := gin.New()
		r.PUT("/courses/:courseId/send-to-review", h.SendToReview)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("PUT", "/courses/1/send-to-review", nil))
		assert.Equal(t, http.StatusConflict, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("not_found", func(t *testing.T) {
		mockService, _, h := setupTransitionHandler(t)
		mockService.On("SendToReview", mock.Anything, 1).Return(errors.New("curso não encontrado"))
		r := gin.New()
		r.PUT("/courses/:courseId/send-to-review", h.SendToReview)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("PUT", "/courses/1/send-to-review", nil))
		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("bad_request_incomplete_course", func(t *testing.T) {
		mockService, _, h := setupTransitionHandler(t)
		mockService.On("SendToReview", mock.Anything, 1).Return(errors.New("curso incompleto para envio de revisão: modalidade inválida"))
		r := gin.New()
		r.PUT("/courses/:courseId/send-to-review", h.SendToReview)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("PUT", "/courses/1/send-to-review", nil))
		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestCourseHandler_Approve(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockService, _, h := setupTransitionHandler(t)
		mockService.On("Approve", mock.Anything, 2).Return(nil)
		r := gin.New()
		r.PUT("/courses/:courseId/approve", h.Approve)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("PUT", "/courses/2/approve", nil))
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("conflict", func(t *testing.T) {
		mockService, _, h := setupTransitionHandler(t)
		mockService.On("Approve", mock.Anything, 2).Return(errors.New("curso não está em estado válido para aprovação"))
		r := gin.New()
		r.PUT("/courses/:courseId/approve", h.Approve)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("PUT", "/courses/2/approve", nil))
		assert.Equal(t, http.StatusConflict, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestCourseHandler_RequestChanges(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockService, _, h := setupTransitionHandler(t)
		mockService.On("RequestChanges", mock.Anything, 4).Return(nil)
		r := gin.New()
		r.PUT("/courses/:courseId/request-changes", h.RequestChanges)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("PUT", "/courses/4/request-changes", nil))
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("conflict", func(t *testing.T) {
		mockService, _, h := setupTransitionHandler(t)
		mockService.On("RequestChanges", mock.Anything, 4).Return(errors.New("curso não está em estado válido para solicitação de edição"))
		r := gin.New()
		r.PUT("/courses/:courseId/request-changes", h.RequestChanges)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("PUT", "/courses/4/request-changes", nil))
		assert.Equal(t, http.StatusConflict, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestCourseHandler_RequestDeletion(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockService, _, h := setupTransitionHandler(t)
		mockService.On("RequestDeletion", mock.Anything, 5).Return(nil)
		r := gin.New()
		r.PUT("/courses/:courseId/request-deletion", h.RequestDeletion)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("PUT", "/courses/5/request-deletion", nil))
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("not_found", func(t *testing.T) {
		mockService, _, h := setupTransitionHandler(t)
		mockService.On("RequestDeletion", mock.Anything, 5).Return(errors.New("curso não encontrado"))
		r := gin.New()
		r.PUT("/courses/:courseId/request-deletion", h.RequestDeletion)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("PUT", "/courses/5/request-deletion", nil))
		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})
}
