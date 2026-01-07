package v1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/config"
	v1 "github.com/prefeitura-rio/app-go-api/internal/handlers/v1"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	"github.com/prefeitura-rio/app-go-api/internal/models"
)

// Mock PropostaMEI Service
type MockPropostaMEIService struct {
	createID    uuid.UUID
	createError error
	proposta    *models.PropostaMEI
	getError    error
	updateError error
	deleteError error
	propostas   []*models.PropostaMEI
	total       int
	listError   error
}

func (m *MockPropostaMEIService) Create(ctx context.Context, proposta *models.PropostaMEI, authToken string) (uuid.UUID, error) {
	if m.createError != nil {
		return uuid.Nil, m.createError
	}
	return m.createID, nil
}

func (m *MockPropostaMEIService) GetByID(ctx context.Context, id uuid.UUID) (*models.PropostaMEI, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	return m.proposta, nil
}

func (m *MockPropostaMEIService) Update(ctx context.Context, proposta *models.PropostaMEI) error {
	return m.updateError
}

func (m *MockPropostaMEIService) UpdateProposta(ctx context.Context, id uuid.UUID, oportunidadeID int, valorProposta *float64) error {
	return m.updateError
}

func (m *MockPropostaMEIService) UpdateStatusCidadao(ctx context.Context, id uuid.UUID, status models.StatusPropostaCidadao) error {
	return m.updateError
}

func (m *MockPropostaMEIService) Approve(ctx context.Context, id uuid.UUID) error {
	return m.updateError
}

func (m *MockPropostaMEIService) Reject(ctx context.Context, id uuid.UUID) error {
	return m.updateError
}

func (m *MockPropostaMEIService) Delete(ctx context.Context, id uuid.UUID) error {
	return m.deleteError
}

func (m *MockPropostaMEIService) ListByOportunidade(ctx context.Context, oportunidadeID int, nomeEmpresa, cnpj, status string, page, pageSize int) ([]*models.PropostaMEI, int, error) {
	if m.listError != nil {
		return nil, 0, m.listError
	}
	return m.propostas, m.total, nil
}

func (m *MockPropostaMEIService) ListByMEIEmpresa(ctx context.Context, meiEmpresaID string, page, pageSize int) ([]*models.PropostaMEI, int, error) {
	if m.listError != nil {
		return nil, 0, m.listError
	}
	return m.propostas, m.total, nil
}

func (m *MockPropostaMEIService) ListByStatus(ctx context.Context, status models.StatusPropostaCidadao, page, pageSize int) ([]*models.PropostaMEI, int, error) {
	if m.listError != nil {
		return nil, 0, m.listError
	}
	return m.propostas, m.total, nil
}

func (m *MockPropostaMEIService) UpdateMultipleStatus(ctx context.Context, propostaIDs []uuid.UUID, status models.StatusPropostaCidadao) (int, error) {
	if m.updateError != nil {
		return 0, m.updateError
	}
	return len(propostaIDs), nil
}

// Mock CNAE Validation Service
type MockCNAEValidationService struct {
	ownsCNPJ       bool
	ownershipError error
}

func (m *MockCNAEValidationService) ValidatePropostaForCNAE(ctx context.Context, authToken string, cnpj string, opportunityCNAEIDs []string) error {
	return nil
}

func (m *MockCNAEValidationService) CheckCNPJOwnership(ctx context.Context, authToken string, cpf string, cnpj string) (bool, error) {
	// For tests, default to true if not set
	if m.ownershipError != nil {
		return false, m.ownershipError
	}
	if m.ownsCNPJ {
		return true, nil
	}
	// Default to true for tests
	return true, nil
}

func setupRouter(service *MockPropostaMEIService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Create mock CNAE validation service
	cnaeValidationSvc := &MockCNAEValidationService{}

	// Create minimal config for tests
	cfg := &config.AppConfig{
		PropostaMEI: config.PropostaMEIPermissions{
			DeletePermissions: []string{},
			UpdatePermissions: []string{},
		},
	}

	handler := v1.NewPropostaMEIHandler(service, cnaeValidationSvc, nil, cfg)

	// Setup routes matching actual router with user context middleware
	api := r.Group("/api/v1/oportunidades-mei/:id/propostas")
	api.Use(middlewares.ExtractUserContext())
	{
		api.POST("", handler.Create)
		api.GET("/:propostaId", handler.GetByID)
		api.PUT("/:propostaId", handler.Update)
		api.PUT("/:propostaId/status", handler.UpdateStatus)
		api.PUT("/status", handler.UpdateStatusBulk)
		api.DELETE("/:propostaId", handler.Delete)
		api.GET("", handler.List)
	}

	// Additional route for ListByMEIEmpresa
	r.GET("/api/v1/propostas-mei/por-empresa", handler.ListByMEIEmpresa)

	return r
}

func TestPropostaMEIHandler_Create_Success(t *testing.T) {
	mockService := &MockPropostaMEIService{
		createID: uuid.New(),
	}

	router := setupRouter(mockService)

	requestBody := map[string]interface{}{
		"mei_empresa_id": "12345678000190",
		"valor_proposta": 1500.00,
	}
	body, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", "/api/v1/oportunidades-mei/1/propostas", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var response models.PropostaMEI
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.ID == uuid.Nil {
		t.Error("Expected non-nil UUID in response")
	}
}

func TestPropostaMEIHandler_Create_MissingAuth(t *testing.T) {
	mockService := &MockPropostaMEIService{}
	router := setupRouter(mockService)

	requestBody := map[string]interface{}{
		"mei_empresa_id": "12345678000190",
	}
	body, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", "/api/v1/oportunidades-mei/1/propostas", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	// No Authorization header

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["error"] != "Token de autorização não fornecido" {
		t.Errorf("Unexpected error message: %v", response["error"])
	}
}

func TestPropostaMEIHandler_Create_CNPJOwnershipError(t *testing.T) {
	mockService := &MockPropostaMEIService{
		createError: errors.New("o CNPJ 12345678000190 não pertence ao seu CPF"),
	}

	router := setupRouter(mockService)

	requestBody := map[string]interface{}{
		"mei_empresa_id": "12345678000190",
	}
	body, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", "/api/v1/oportunidades-mei/1/propostas", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestPropostaMEIHandler_Create_CNAEMismatchError(t *testing.T) {
	mockService := &MockPropostaMEIService{
		createError: errors.New("o CNPJ 12345678000190 não possui CNAE compatível com esta oportunidade"),
	}

	router := setupRouter(mockService)

	requestBody := map[string]interface{}{
		"mei_empresa_id": "12345678000190",
	}
	body, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", "/api/v1/oportunidades-mei/1/propostas", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestPropostaMEIHandler_Create_OpportunityNotFound(t *testing.T) {
	mockService := &MockPropostaMEIService{
		createError: errors.New("oportunidade não encontrada"),
	}

	router := setupRouter(mockService)

	requestBody := map[string]interface{}{
		"mei_empresa_id": "12345678000190",
	}
	body, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", "/api/v1/oportunidades-mei/999/propostas", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestPropostaMEIHandler_Create_RMIServiceUnavailable(t *testing.T) {
	mockService := &MockPropostaMEIService{
		createError: errors.New("não foi possível validar seus CNPJs no momento"),
	}

	router := setupRouter(mockService)

	requestBody := map[string]interface{}{
		"mei_empresa_id": "12345678000190",
	}
	body, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", "/api/v1/oportunidades-mei/1/propostas", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", w.Code)
	}
}

func TestPropostaMEIHandler_Create_InvalidOportunidadeID(t *testing.T) {
	mockService := &MockPropostaMEIService{}
	router := setupRouter(mockService)

	requestBody := map[string]interface{}{
		"mei_empresa_id": "12345678000190",
	}
	body, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", "/api/v1/oportunidades-mei/invalid/propostas", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["error"] != "ID da oportunidade inválido" {
		t.Errorf("Unexpected error message: %v", response["error"])
	}
}

func TestPropostaMEIHandler_GetByID_Success(t *testing.T) {
	propostaID := uuid.New()
	mockService := &MockPropostaMEIService{
		proposta: &models.PropostaMEI{
			ID:           propostaID,
			MEIEmpresaID: "12345678000190",
		},
	}

	router := setupRouter(mockService)

	req, _ := http.NewRequest("GET", "/api/v1/oportunidades-mei/1/propostas/"+propostaID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response models.PropostaMEI
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.ID != propostaID {
		t.Errorf("Expected proposta ID %s, got %s", propostaID, response.ID)
	}
}

func TestPropostaMEIHandler_GetByID_NotFound(t *testing.T) {
	mockService := &MockPropostaMEIService{
		proposta: nil, // Not found
	}

	router := setupRouter(mockService)

	propostaID := uuid.New()
	req, _ := http.NewRequest("GET", "/api/v1/oportunidades-mei/1/propostas/"+propostaID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestPropostaMEIHandler_GetByID_InvalidUUID(t *testing.T) {
	mockService := &MockPropostaMEIService{}
	router := setupRouter(mockService)

	req, _ := http.NewRequest("GET", "/api/v1/oportunidades-mei/1/propostas/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestPropostaMEIHandler_UpdateStatus_Success(t *testing.T) {
	propostaID := uuid.New()
	mockService := &MockPropostaMEIService{
		proposta: &models.PropostaMEI{
			ID:            propostaID,
			StatusCidadao: models.StatusPropostaCidadaoApproved,
		},
	}

	router := setupRouter(mockService)

	requestBody := map[string]interface{}{
		"status": "approved",
	}
	body, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("PUT", "/api/v1/oportunidades-mei/1/propostas/"+propostaID.String()+"/status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestPropostaMEIHandler_UpdateStatus_InvalidStatus(t *testing.T) {
	propostaID := uuid.New()
	mockService := &MockPropostaMEIService{}
	router := setupRouter(mockService)

	requestBody := map[string]interface{}{
		"status": "invalid-status",
	}
	body, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("PUT", "/api/v1/oportunidades-mei/1/propostas/"+propostaID.String()+"/status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	expectedError := "Status inválido. Use: approved, rejected ou submitted"
	if response["error"] != expectedError {
		t.Errorf("Expected error '%s', got '%v'", expectedError, response["error"])
	}
}

func TestPropostaMEIHandler_Delete_Success(t *testing.T) {
	propostaID := uuid.New()
	mockService := &MockPropostaMEIService{
		proposta: &models.PropostaMEI{
			ID:           propostaID,
			MEIEmpresaID: "12345678000190",
		},
	}
	router := setupRouter(mockService)

	req, _ := http.NewRequest("DELETE", "/api/v1/oportunidades-mei/1/propostas/"+propostaID.String(), nil)
	// Add required headers for authorization
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("X-User-CPF", "12345678900")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Response body: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["message"] != "Proposta excluída com sucesso" {
		t.Errorf("Unexpected message: %v", response["message"])
	}
}

func TestPropostaMEIHandler_List_Success(t *testing.T) {
	mockService := &MockPropostaMEIService{
		propostas: []*models.PropostaMEI{
			{ID: uuid.New(), MEIEmpresaID: "12345678000190"},
			{ID: uuid.New(), MEIEmpresaID: "98765432000199"},
		},
		total: 2,
	}

	router := setupRouter(mockService)

	req, _ := http.NewRequest("GET", "/api/v1/oportunidades-mei/1/propostas?page=1&pageSize=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	data := response["data"].([]interface{})
	if len(data) != 2 {
		t.Errorf("Expected 2 propostas, got %d", len(data))
	}

	meta := response["meta"].(map[string]interface{})
	if meta["total"] != float64(2) {
		t.Errorf("Expected total 2, got %v", meta["total"])
	}
}

func TestPropostaMEIHandler_ListByMEIEmpresa_Success(t *testing.T) {
	mockService := &MockPropostaMEIService{
		propostas: []*models.PropostaMEI{
			{ID: uuid.New(), MEIEmpresaID: "12345678000190"},
		},
		total: 1,
	}

	router := setupRouter(mockService)

	req, _ := http.NewRequest("GET", "/api/v1/propostas-mei/por-empresa?meiEmpresaId=12345678000190&page=1&pageSize=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestPropostaMEIHandler_ListByMEIEmpresa_MissingParam(t *testing.T) {
	mockService := &MockPropostaMEIService{}
	router := setupRouter(mockService)

	req, _ := http.NewRequest("GET", "/api/v1/propostas-mei/por-empresa", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["error"] != "CNPJ da MEI empresa não fornecido" {
		t.Errorf("Unexpected error message: %v", response["error"])
	}
}
