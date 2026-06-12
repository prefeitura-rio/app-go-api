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
	"github.com/stretchr/testify/assert"
)

// mockInscricaoServiceForList extends the basic mock for List-specific tests
type mockInscricaoServiceForList struct {
	getByCursoErr error
	summaryErr    error
	listByCPFErr  error
	enrollments   []*models.Inscricao
	total         int
	summary       *models.EnrollmentSummary
}

func (m *mockInscricaoServiceForList) Create(ctx context.Context, inscricao *models.Inscricao) error {
	return nil
}

func (m *mockInscricaoServiceForList) CreateByAdmin(ctx context.Context, inscricao *models.Inscricao) error {
	return nil
}

func (m *mockInscricaoServiceForList) CreateManual(ctx context.Context, inscricao *models.Inscricao) error {
	return nil
}

func (m *mockInscricaoServiceForList) GetByID(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
	return nil, nil
}

func (m *mockInscricaoServiceForList) GetByCursoID(ctx context.Context, cursoID int, filter map[string]interface{}, page, limit int) ([]*models.Inscricao, int, error) {
	if m.getByCursoErr != nil {
		return nil, 0, m.getByCursoErr
	}
	return m.enrollments, m.total, nil
}

func (m *mockInscricaoServiceForList) GetSummaryByCursoID(ctx context.Context, cursoID int) (*models.EnrollmentSummary, error) {
	if m.summaryErr != nil {
		return nil, m.summaryErr
	}
	return m.summary, nil
}

func (m *mockInscricaoServiceForList) ListByCPF(ctx context.Context, cpf string, filter map[string]interface{}, offset, limit int) ([]*models.Inscricao, int, error) {
	if m.listByCPFErr != nil {
		return nil, 0, m.listByCPFErr
	}
	return m.enrollments, m.total, nil
}

func (m *mockInscricaoServiceForList) EnrichWithPersonalInfo(ctx context.Context, inscricao *models.Inscricao) {
}

func (m *mockInscricaoServiceForList) EnrichMultipleWithPersonalInfo(ctx context.Context, inscricoes []*models.Inscricao) {
}

func (m *mockInscricaoServiceForList) UpdateStatus(ctx context.Context, id uuid.UUID, status models.StatusInscricao, reason, adminNotes string) error {
	return nil
}

func (m *mockInscricaoServiceForList) UpdateMultipleStatus(ctx context.Context, ids []uuid.UUID, status models.StatusInscricao, reason, adminNotes string) (int, error) {
	return 0, nil
}

func (m *mockInscricaoServiceForList) UpdateCertificate(ctx context.Context, cursoID int, id uuid.UUID, certURL string) error {
	return nil
}

func (m *mockInscricaoServiceForList) UpdateInscricao(ctx context.Context, id uuid.UUID, cursoID int, updateData *models.InscricaoUpdateRequest) error {
	return nil
}

func (m *mockInscricaoServiceForList) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockInscricaoServiceForList) ChangeSchedule(ctx context.Context, id uuid.UUID, userCPF string, request *models.ScheduleChangeRequest) (*models.Inscricao, error) {
	return nil, nil
}

// ===================================
// List Tests - Expanding Coverage
// ===================================

// Test List with negative page number (should default to 1)
func TestInscricaoHandler_List_NegativePage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	mockSvc := &mockInscricaoServiceForList{
		enrollments: []*models.Inscricao{},
		total:       0,
		summary:     &models.EnrollmentSummary{},
	}
	h := v1.NewInscricaoHandler(mockSvc, nil, nil)
	r.GET("/api/v1/courses/:courseId/enrollments", h.List)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses/1/enrollments?page=-5", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should succeed with page defaulted to 1
	assert.Equal(t, http.StatusOK, w.Code)
}

// Test List with zero limit (should default to 20)
func TestInscricaoHandler_List_ZeroLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	mockSvc := &mockInscricaoServiceForList{
		enrollments: []*models.Inscricao{},
		total:       0,
		summary:     &models.EnrollmentSummary{},
	}
	h := v1.NewInscricaoHandler(mockSvc, nil, nil)
	r.GET("/api/v1/courses/:courseId/enrollments", h.List)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses/1/enrollments?limit=0", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should succeed with limit defaulted to 20
	assert.Equal(t, http.StatusOK, w.Code)
}

// Test List with excessive limit (should cap at 1000 -> 20)
func TestInscricaoHandler_List_ExcessiveLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	mockSvc := &mockInscricaoServiceForList{
		enrollments: []*models.Inscricao{},
		total:       0,
		summary:     &models.EnrollmentSummary{},
	}
	h := v1.NewInscricaoHandler(mockSvc, nil, nil)
	r.GET("/api/v1/courses/:courseId/enrollments", h.List)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses/1/enrollments?limit=5000", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should succeed with limit capped to 20
	assert.Equal(t, http.StatusOK, w.Code)
}

// Test List with status filter
func TestInscricaoHandler_List_WithStatusFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	mockSvc := &mockInscricaoServiceForList{
		enrollments: []*models.Inscricao{},
		total:       0,
		summary:     &models.EnrollmentSummary{},
	}
	h := v1.NewInscricaoHandler(mockSvc, nil, nil)
	r.GET("/api/v1/courses/:courseId/enrollments", h.List)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses/1/enrollments?status=approved", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Test List with search filter
func TestInscricaoHandler_List_WithSearchFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	mockSvc := &mockInscricaoServiceForList{
		enrollments: []*models.Inscricao{},
		total:       0,
		summary:     &models.EnrollmentSummary{},
	}
	h := v1.NewInscricaoHandler(mockSvc, nil, nil)
	r.GET("/api/v1/courses/:courseId/enrollments", h.List)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses/1/enrollments?search=john", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Test List when service returns error
func TestInscricaoHandler_List_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	mockSvc := &mockInscricaoServiceForList{
		getByCursoErr: errors.New("database connection error"),
	}
	h := v1.NewInscricaoHandler(mockSvc, nil, nil)
	r.GET("/api/v1/courses/:courseId/enrollments", h.List)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses/1/enrollments", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Erro ao listar inscrições")
}

// Test List when summary service returns error
func TestInscricaoHandler_List_SummaryError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	mockSvc := &mockInscricaoServiceForList{
		enrollments: []*models.Inscricao{},
		total:       0,
		summaryErr:  errors.New("summary query error"),
	}
	h := v1.NewInscricaoHandler(mockSvc, nil, nil)
	r.GET("/api/v1/courses/:courseId/enrollments", h.List)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses/1/enrollments", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Erro ao obter resumo")
}

// ===================================
// ListByUser Tests - Expanding Coverage
// ===================================

// Test ListByUser with negative page (should default to 1)
func TestInscricaoHandler_ListByUser_NegativePage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	mockSvc := &mockInscricaoServiceForList{
		enrollments: []*models.Inscricao{},
		total:       0,
	}
	h := v1.NewInscricaoHandler(mockSvc, nil, nil)
	r.GET("/api/v1/enrollments/user/:cpf", h.ListByUser)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/enrollments/user/12345678901?page=-5", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should succeed with page defaulted to 1
	assert.Equal(t, http.StatusOK, w.Code)
}

// Test ListByUser with zero limit (should default to 10)
func TestInscricaoHandler_ListByUser_ZeroLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	mockSvc := &mockInscricaoServiceForList{
		enrollments: []*models.Inscricao{},
		total:       0,
	}
	h := v1.NewInscricaoHandler(mockSvc, nil, nil)
	r.GET("/api/v1/enrollments/user/:cpf", h.ListByUser)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/enrollments/user/12345678901?limit=0", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should succeed with limit defaulted to 10
	assert.Equal(t, http.StatusOK, w.Code)
}

// Test ListByUser with excessive limit (should cap at 1000 -> 10)
func TestInscricaoHandler_ListByUser_ExcessiveLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	mockSvc := &mockInscricaoServiceForList{
		enrollments: []*models.Inscricao{},
		total:       0,
	}
	h := v1.NewInscricaoHandler(mockSvc, nil, nil)
	r.GET("/api/v1/enrollments/user/:cpf", h.ListByUser)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/enrollments/user/12345678901?limit=5000", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should succeed with limit capped to 10
	assert.Equal(t, http.StatusOK, w.Code)
}

// Test ListByUser with status filter
func TestInscricaoHandler_ListByUser_WithStatusFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	mockSvc := &mockInscricaoServiceForList{
		enrollments: []*models.Inscricao{},
		total:       0,
	}
	h := v1.NewInscricaoHandler(mockSvc, nil, nil)
	r.GET("/api/v1/enrollments/user/:cpf", h.ListByUser)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/enrollments/user/12345678901?status=approved", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Test ListByUser when service returns error
func TestInscricaoHandler_ListByUser_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	mockSvc := &mockInscricaoServiceForList{
		listByCPFErr: errors.New("database connection error"),
	}
	h := v1.NewInscricaoHandler(mockSvc, nil, nil)
	r.GET("/api/v1/enrollments/user/:cpf", h.ListByUser)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/enrollments/user/12345678901", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Erro ao listar inscrições do usuário")
}
