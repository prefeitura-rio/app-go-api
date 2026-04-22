package v1_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	v1 "github.com/prefeitura-rio/app-go-api/internal/handlers/v1"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/stretchr/testify/assert"
)

// Test basic handler initialization
func TestNewInscricaoHandler(t *testing.T) {
	handler := v1.NewInscricaoHandler(nil, nil, nil)
	assert.NotNil(t, handler)
}

// Test Create endpoint with invalid courseId
func TestInscricaoHandler_Create_InvalidCourseID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewInscricaoHandler(nil, nil, nil)
	r.POST("/api/v1/courses/:courseId/enrollments", h.Create)

	body := []byte(`{"cpf": "12345678901"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses/invalid/enrollments", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "ID do curso inválido")
}

// Test Create endpoint with invalid JSON
func TestInscricaoHandler_Create_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewInscricaoHandler(nil, nil, nil)
	r.POST("/api/v1/courses/:courseId/enrollments", h.Create)

	body := []byte(`{invalid json}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses/1/enrollments", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Dados inválidos")
}

// Test List endpoint with invalid courseId
func TestInscricaoHandler_List_InvalidCourseID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewInscricaoHandler(nil, nil, nil)
	r.GET("/api/v1/courses/:courseId/enrollments", h.List)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses/invalid/enrollments", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test Update endpoint with invalid courseId
func TestInscricaoHandler_Update_InvalidCourseID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewInscricaoHandler(nil, nil, nil)
	r.PUT("/api/v1/courses/:courseId/enrollments/:enrollmentId", h.Update)

	body := []byte(`{"status": "approved"}`)
	enrollmentID := uuid.New().String()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/courses/invalid/enrollments/"+enrollmentID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test Update endpoint with invalid enrollmentId
func TestInscricaoHandler_Update_InvalidEnrollmentID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewInscricaoHandler(nil, nil, nil)
	r.PUT("/api/v1/courses/:courseId/enrollments/:enrollmentId", h.Update)

	body := []byte(`{"status": "approved"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/courses/1/enrollments/invalid-uuid", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "ID da inscrição inválido")
}

// Test Update endpoint with invalid JSON
func TestInscricaoHandler_Update_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewInscricaoHandler(nil, nil, nil)
	r.PUT("/api/v1/courses/:courseId/enrollments/:enrollmentId", h.Update)

	enrollmentID := uuid.New().String()
	body := []byte(`{invalid}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/courses/1/enrollments/"+enrollmentID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test UpdateStatus endpoint with invalid courseId
func TestInscricaoHandler_UpdateStatus_InvalidCourseID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewInscricaoHandler(nil, nil, nil)
	r.PUT("/api/v1/courses/:courseId/enrollments/status", h.UpdateStatus)

	body := []byte(`{"enrollment_ids": [], "status": "approved"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/courses/invalid/enrollments/status", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test UpdateStatus endpoint with invalid JSON
func TestInscricaoHandler_UpdateStatus_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewInscricaoHandler(nil, nil, nil)
	r.PUT("/api/v1/courses/:courseId/enrollments/status", h.UpdateStatus)

	body := []byte(`{invalid}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/courses/1/enrollments/status", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test UpdateIndividualStatus endpoint with invalid enrollmentId
func TestInscricaoHandler_UpdateIndividualStatus_InvalidEnrollmentID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewInscricaoHandler(nil, nil, nil)
	r.PUT("/api/v1/courses/:courseId/enrollments/:enrollmentId/status", h.UpdateIndividualStatus)

	body := []byte(`{"status": "approved"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/courses/1/enrollments/invalid/status", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test UpdateIndividualStatus endpoint with invalid JSON
func TestInscricaoHandler_UpdateIndividualStatus_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewInscricaoHandler(nil, nil, nil)
	r.PUT("/api/v1/courses/:courseId/enrollments/:enrollmentId/status", h.UpdateIndividualStatus)

	enrollmentID := uuid.New().String()
	body := []byte(`{invalid}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/courses/1/enrollments/"+enrollmentID+"/status", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test GetByID endpoint with invalid enrollmentId
func TestInscricaoHandler_GetByID_InvalidEnrollmentID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewInscricaoHandler(nil, nil, nil)
	r.GET("/api/v1/courses/:courseId/enrollments/:enrollmentId", h.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses/1/enrollments/invalid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test UpdateCertificate endpoint with invalid courseId
func TestInscricaoHandler_UpdateCertificate_InvalidCourseID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewInscricaoHandler(nil, nil, nil)
	r.PUT("/api/v1/courses/:courseId/enrollments/:enrollmentId/certificate", h.UpdateCertificate)

	enrollmentID := uuid.New().String()
	body := []byte(`{"certificate_url": "http://example.com/cert.pdf"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/courses/invalid/enrollments/"+enrollmentID+"/certificate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test UpdateCertificate endpoint with invalid enrollmentId
func TestInscricaoHandler_UpdateCertificate_InvalidEnrollmentID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewInscricaoHandler(nil, nil, nil)
	r.PUT("/api/v1/courses/:courseId/enrollments/:enrollmentId/certificate", h.UpdateCertificate)

	body := []byte(`{"certificate_url": "http://example.com/cert.pdf"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/courses/1/enrollments/invalid/certificate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test UpdateCertificate endpoint with invalid JSON
func TestInscricaoHandler_UpdateCertificate_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewInscricaoHandler(nil, nil, nil)
	r.PUT("/api/v1/courses/:courseId/enrollments/:enrollmentId/certificate", h.UpdateCertificate)

	enrollmentID := uuid.New().String()
	body := []byte(`{invalid}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/courses/1/enrollments/"+enrollmentID+"/certificate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test Delete endpoint with invalid enrollmentId
func TestInscricaoHandler_Delete_InvalidEnrollmentID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewInscricaoHandler(nil, nil, nil)
	r.Use(func(c *gin.Context) {
		c.Set("user_role", "ADMIN")
		c.Next()
	})
	r.DELETE("/api/v1/courses/:courseId/enrollments/:enrollmentId", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/courses/1/enrollments/invalid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test CreateManual endpoint with invalid courseId
func TestInscricaoHandler_CreateManual_InvalidCourseID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewInscricaoHandler(nil, nil, nil)
	r.POST("/api/v1/courses/:courseId/enrollments/manual", h.CreateManual)

	body := []byte(`{"cpf": "12345678901"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses/invalid/enrollments/manual", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test CreateManual endpoint with invalid JSON
func TestInscricaoHandler_CreateManual_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewInscricaoHandler(nil, nil, nil)
	r.POST("/api/v1/courses/:courseId/enrollments/manual", h.CreateManual)

	body := []byte(`{invalid}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses/1/enrollments/manual", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test Import endpoint with invalid courseId
func TestInscricaoHandler_Import_InvalidCourseID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewInscricaoHandler(nil, nil, nil)
	r.POST("/api/v1/courses/:courseId/enrollments/import", h.Import)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses/invalid/enrollments/import", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test ListByUser endpoint with missing CPF
func TestInscricaoHandler_ListByUser_MissingCPF(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewInscricaoHandler(nil, nil, nil)
	r.GET("/api/v1/enrollments/user/:cpf", h.ListByUser)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/enrollments/user/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should return 404 for missing route parameter
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// Test ChangeSchedule endpoint with invalid enrollmentId
func TestInscricaoHandler_ChangeSchedule_InvalidEnrollmentID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewInscricaoHandler(nil, nil, nil)
	r.PUT("/api/v1/enrollments/:enrollmentId/schedule", h.ChangeSchedule)

	body := []byte(`{"schedule_id": "123"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/enrollments/invalid/schedule", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test ChangeSchedule endpoint - checks authorization before JSON, so returns 403
func TestInscricaoHandler_ChangeSchedule_MissingAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewInscricaoHandler(nil, nil, nil)
	r.PUT("/api/v1/enrollments/:enrollmentId/schedule", h.ChangeSchedule)

	enrollmentID := uuid.New().String()
	body := []byte(`{"schedule_id": "123"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/enrollments/"+enrollmentID+"/schedule", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Returns 403 when user_cpf is not set in context
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// Test UpdateIndividualStatus endpoint with invalid courseId
func TestInscricaoHandler_UpdateIndividualStatus_InvalidCourseID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewInscricaoHandler(nil, nil, nil)
	r.PUT("/api/v1/courses/:courseId/enrollments/:enrollmentId/status", h.UpdateIndividualStatus)

	enrollmentID := uuid.New().String()
	body := []byte(`{"status": "approved"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/courses/invalid/enrollments/"+enrollmentID+"/status", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "ID do curso inválido")
}

// mockInscricaoService is a mock implementation of InscricaoServiceInterface for testing
type mockInscricaoService struct {
	createManualErr error
}

func (m *mockInscricaoService) CreateManual(ctx context.Context, inscricao *models.Inscricao) error {
	return m.createManualErr
}

// Dummy implementations for other interface methods (not used in CreateManual tests)
func (m *mockInscricaoService) Create(ctx context.Context, inscricao *models.Inscricao) error {
	return nil
}
func (m *mockInscricaoService) GetByID(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
	return nil, nil
}
func (m *mockInscricaoService) GetByCursoID(ctx context.Context, cursoID int, filter map[string]interface{}, page, pageSize int) ([]*models.Inscricao, int, error) {
	return nil, 0, nil
}
func (m *mockInscricaoService) UpdateStatus(ctx context.Context, inscricaoID uuid.UUID, status models.StatusInscricao, reason, adminNotes string) error {
	return nil
}
func (m *mockInscricaoService) UpdateMultipleStatus(ctx context.Context, inscricaoIDs []uuid.UUID, status models.StatusInscricao, reason, adminNotes string) (int, error) {
	return 0, nil
}
func (m *mockInscricaoService) GetSummaryByCursoID(ctx context.Context, cursoID int) (*models.EnrollmentSummary, error) {
	return nil, nil
}
func (m *mockInscricaoService) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockInscricaoService) ListByCPF(ctx context.Context, cpf string, filter map[string]interface{}, offset, limit int) ([]*models.Inscricao, int, error) {
	return nil, 0, nil
}
func (m *mockInscricaoService) UpdateCertificate(ctx context.Context, cursoID int, inscricaoID uuid.UUID, certificateURL string) error {
	return nil
}
func (m *mockInscricaoService) UpdateInscricao(ctx context.Context, id uuid.UUID, cursoID int, updateData *models.InscricaoUpdateRequest) error {
	return nil
}
func (m *mockInscricaoService) EnrichWithPersonalInfo(ctx context.Context, inscricao *models.Inscricao) {
}
func (m *mockInscricaoService) EnrichMultipleWithPersonalInfo(ctx context.Context, inscricoes []*models.Inscricao) {
}
func (m *mockInscricaoService) ChangeSchedule(ctx context.Context, inscricaoID uuid.UUID, userCPF string, request *models.ScheduleChangeRequest) (*models.Inscricao, error) {
	return nil, nil
}

// Test CreateManual endpoint with conflict error (CPF already enrolled)
func TestInscricaoHandler_CreateManual_ConflictError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &mockInscricaoService{
		createManualErr: assert.AnError,
	}
	// Override error message to simulate conflict
	mockService.createManualErr = &conflictError{msg: "CPF já inscrito neste curso"}

	h := v1.NewInscricaoHandler(mockService, nil, nil)
	r := gin.New()
	r.POST("/api/v1/courses/:courseId/enrollments/manual", h.CreateManual)

	body := []byte(`{"cpf": "12345678901", "name": "Test User"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses/1/enrollments/manual", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "CPF já inscrito neste curso")
}

// Test CreateManual endpoint with service error
func TestInscricaoHandler_CreateManual_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &mockInscricaoService{
		createManualErr: &serviceError{msg: "database connection error"},
	}

	h := v1.NewInscricaoHandler(mockService, nil, nil)
	r := gin.New()
	r.POST("/api/v1/courses/:courseId/enrollments/manual", h.CreateManual)

	body := []byte(`{"cpf": "12345678901", "name": "Test User"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses/1/enrollments/manual", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Erro ao criar inscrição")
}

// Test CreateManual endpoint with success
func TestInscricaoHandler_CreateManual_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &mockInscricaoService{
		createManualErr: nil, // No error = success
	}

	h := v1.NewInscricaoHandler(mockService, nil, nil)
	r := gin.New()
	r.POST("/api/v1/courses/:courseId/enrollments/manual", h.CreateManual)

	body := []byte(`{"cpf": "12345678901", "name": "Test User", "email": "test@example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses/1/enrollments/manual", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "success")
	assert.Contains(t, w.Body.String(), "Inscrição manual criada com sucesso")
}

// Helper error types for testing
type conflictError struct {
	msg string
}

func (e *conflictError) Error() string {
	return e.msg
}

type serviceError struct {
	msg string
}

func (e *serviceError) Error() string {
	return e.msg
}
