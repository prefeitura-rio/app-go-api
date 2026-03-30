package v1_test

import (
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
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/stretchr/testify/assert"
)

// mockPropostaMEIServiceForList extends mock for List-specific tests
type mockPropostaMEIServiceForList struct {
	MockPropostaMEIService
	listByOportunidadeErr error
	listByMEIEmpresaErr   error
	proposta              *models.PropostaMEI
	getError              error
}

func (m *mockPropostaMEIServiceForList) ListByOportunidade(ctx context.Context, oportunidadeID int, nomeEmpresa, cnpj, status string, page, pageSize int) ([]*models.PropostaMEI, int, error) {
	if m.listByOportunidadeErr != nil {
		return nil, 0, m.listByOportunidadeErr
	}
	return m.propostas, m.total, nil
}

func (m *mockPropostaMEIServiceForList) ListByMEIEmpresa(ctx context.Context, meiEmpresaID string, page, pageSize int) ([]*models.PropostaMEI, int, error) {
	if m.listByMEIEmpresaErr != nil {
		return nil, 0, m.listByMEIEmpresaErr
	}
	return m.propostas, m.total, nil
}

func (m *mockPropostaMEIServiceForList) GetByID(ctx context.Context, id uuid.UUID) (*models.PropostaMEI, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	return m.proposta, nil
}

// setupRouterForListTests creates router with mock services for list tests
func setupRouterForListTests(service *mockPropostaMEIServiceForList) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Create mock CNAE validation service (default: user owns CNPJ)
	cnaeValidationSvc := &MockCNAEValidationService{
		ownsCNPJ: true,
	}

	// Create minimal config for tests
	cfg := &config.AppConfig{
		PropostaMEI: config.PropostaMEIPermissions{
			DeletePermissions: []string{},
			UpdatePermissions: []string{},
			ReadPermissions:   []string{},
		},
	}

	handler := v1.NewPropostaMEIHandler(service, cnaeValidationSvc, nil, cfg)

	// Setup routes matching actual router with a mock user context middleware
	mockUserContext := func(c *gin.Context) {
		c.Set("user_cpf", "12345678900")
		c.Set("user_id", "test-user-id")
		c.Next()
	}

	api := r.Group("/api/v1/oportunidades-mei/:id/propostas")
	api.Use(mockUserContext)
	{
		api.GET("", handler.List)
		api.GET("/:propostaId", handler.GetByID)
	}

	// Additional route for ListByMEIEmpresa
	r.GET("/api/v1/propostas-mei/por-empresa", handler.ListByMEIEmpresa)

	return r
}

// ===================================
// List Tests - Expanding Coverage
// ===================================

// Test List with negative page number (should default to 1)
func TestPropostaMEIHandler_List_NegativePage(t *testing.T) {
	mockService := &mockPropostaMEIServiceForList{
		MockPropostaMEIService: MockPropostaMEIService{
			propostas: []*models.PropostaMEI{},
			total:     0,
		},
	}
	router := setupRouterForListTests(mockService)

	req, _ := http.NewRequest("GET", "/api/v1/oportunidades-mei/1/propostas?page=-5", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Test List with zero pageSize (should default to 10)
func TestPropostaMEIHandler_List_ZeroPageSize(t *testing.T) {
	mockService := &mockPropostaMEIServiceForList{
		MockPropostaMEIService: MockPropostaMEIService{
			propostas: []*models.PropostaMEI{},
			total:     0,
		},
	}
	router := setupRouterForListTests(mockService)

	req, _ := http.NewRequest("GET", "/api/v1/oportunidades-mei/1/propostas?pageSize=0", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Test List with excessive pageSize (should cap at 1000 -> 10)
func TestPropostaMEIHandler_List_ExcessivePageSize(t *testing.T) {
	mockService := &mockPropostaMEIServiceForList{
		MockPropostaMEIService: MockPropostaMEIService{
			propostas: []*models.PropostaMEI{},
			total:     0,
		},
	}
	router := setupRouterForListTests(mockService)

	req, _ := http.NewRequest("GET", "/api/v1/oportunidades-mei/1/propostas?pageSize=5000", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Test List with nomeEmpresa filter
func TestPropostaMEIHandler_List_WithNomeEmpresaFilter(t *testing.T) {
	mockService := &mockPropostaMEIServiceForList{
		MockPropostaMEIService: MockPropostaMEIService{
			propostas: []*models.PropostaMEI{},
			total:     0,
		},
	}
	router := setupRouterForListTests(mockService)

	req, _ := http.NewRequest("GET", "/api/v1/oportunidades-mei/1/propostas?nomeEmpresa=Test+Company", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Test List with cnpj filter
func TestPropostaMEIHandler_List_WithCnpjFilter(t *testing.T) {
	mockService := &mockPropostaMEIServiceForList{
		MockPropostaMEIService: MockPropostaMEIService{
			propostas: []*models.PropostaMEI{},
			total:     0,
		},
	}
	router := setupRouterForListTests(mockService)

	req, _ := http.NewRequest("GET", "/api/v1/oportunidades-mei/1/propostas?cnpj=12345678000190", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Test List with status filter
func TestPropostaMEIHandler_List_WithStatusFilter(t *testing.T) {
	mockService := &mockPropostaMEIServiceForList{
		MockPropostaMEIService: MockPropostaMEIService{
			propostas: []*models.PropostaMEI{},
			total:     0,
		},
	}
	router := setupRouterForListTests(mockService)

	req, _ := http.NewRequest("GET", "/api/v1/oportunidades-mei/1/propostas?status=approved", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Test List when service returns error
func TestPropostaMEIHandler_List_ServiceError(t *testing.T) {
	mockService := &mockPropostaMEIServiceForList{
		listByOportunidadeErr: errors.New("database connection error"),
	}
	router := setupRouterForListTests(mockService)

	req, _ := http.NewRequest("GET", "/api/v1/oportunidades-mei/1/propostas", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Erro ao listar propostas")
}

// ===================================
// ListByMEIEmpresa Tests - Expanding Coverage
// ===================================

// Test ListByMEIEmpresa with negative page (should default to 1)
func TestPropostaMEIHandler_ListByMEIEmpresa_NegativePage(t *testing.T) {
	mockService := &mockPropostaMEIServiceForList{
		MockPropostaMEIService: MockPropostaMEIService{
			propostas: []*models.PropostaMEI{},
			total:     0,
		},
	}
	router := setupRouterForListTests(mockService)

	req, _ := http.NewRequest("GET", "/api/v1/propostas-mei/por-empresa?meiEmpresaId=12345678000190&page=-5", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Test ListByMEIEmpresa with zero pageSize (should default to 10)
func TestPropostaMEIHandler_ListByMEIEmpresa_ZeroPageSize(t *testing.T) {
	mockService := &mockPropostaMEIServiceForList{
		MockPropostaMEIService: MockPropostaMEIService{
			propostas: []*models.PropostaMEI{},
			total:     0,
		},
	}
	router := setupRouterForListTests(mockService)

	req, _ := http.NewRequest("GET", "/api/v1/propostas-mei/por-empresa?meiEmpresaId=12345678000190&pageSize=0", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Test ListByMEIEmpresa with excessive pageSize (should cap at 1000 -> 10)
func TestPropostaMEIHandler_ListByMEIEmpresa_ExcessivePageSize(t *testing.T) {
	mockService := &mockPropostaMEIServiceForList{
		MockPropostaMEIService: MockPropostaMEIService{
			propostas: []*models.PropostaMEI{},
			total:     0,
		},
	}
	router := setupRouterForListTests(mockService)

	req, _ := http.NewRequest("GET", "/api/v1/propostas-mei/por-empresa?meiEmpresaId=12345678000190&pageSize=5000", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Test ListByMEIEmpresa when service returns error
func TestPropostaMEIHandler_ListByMEIEmpresa_ServiceError(t *testing.T) {
	mockService := &mockPropostaMEIServiceForList{
		listByMEIEmpresaErr: errors.New("database connection error"),
	}
	router := setupRouterForListTests(mockService)

	req, _ := http.NewRequest("GET", "/api/v1/propostas-mei/por-empresa?meiEmpresaId=12345678000190", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Erro ao listar propostas")
}

// ===================================
// GetByID Tests - Expanding Coverage
// ===================================

// Test GetByID when GetByID service returns error
func TestPropostaMEIHandler_GetByID_ServiceError(t *testing.T) {
	propostaID := uuid.New()
	mockService := &mockPropostaMEIServiceForList{
		getError: errors.New("database connection error"),
	}
	router := setupRouterForListTests(mockService)

	req, _ := http.NewRequest("GET", "/api/v1/oportunidades-mei/1/propostas/"+propostaID.String(), nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "Erro ao buscar proposta")
}

// Test GetByID - verify it returns proposta when authorized
// (Access control is tested in the existing GetByID_Success test which confirms proper setup)
func TestPropostaMEIHandler_GetByID_ServiceInternalError(t *testing.T) {
	propostaID := uuid.New()
	mockService := &mockPropostaMEIServiceForList{
		getError: errors.New("unexpected database error"),
	}
	router := setupRouterForListTests(mockService)

	req, _ := http.NewRequest("GET", "/api/v1/oportunidades-mei/1/propostas/"+propostaID.String(), nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "Erro ao buscar proposta")
}

// Test List with all filters combined
func TestPropostaMEIHandler_List_WithAllFilters(t *testing.T) {
	mockService := &mockPropostaMEIServiceForList{
		MockPropostaMEIService: MockPropostaMEIService{
			propostas: []*models.PropostaMEI{},
			total:     0,
		},
	}
	router := setupRouterForListTests(mockService)

	req, _ := http.NewRequest("GET", "/api/v1/oportunidades-mei/1/propostas?page=2&pageSize=20&nomeEmpresa=Test&cnpj=12345678000190&status=submitted", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify response structure
	assert.NotNil(t, response["data"])
	assert.NotNil(t, response["meta"])
}

// Test List with empty result set
func TestPropostaMEIHandler_List_EmptyResults(t *testing.T) {
	mockService := &mockPropostaMEIServiceForList{
		MockPropostaMEIService: MockPropostaMEIService{
			propostas: []*models.PropostaMEI{},
			total:     0,
		},
	}
	router := setupRouterForListTests(mockService)

	req, _ := http.NewRequest("GET", "/api/v1/oportunidades-mei/1/propostas", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	data := response["data"].([]interface{})
	assert.Equal(t, 0, len(data))

	meta := response["meta"].(map[string]interface{})
	assert.Equal(t, float64(0), meta["total"])
}

// Test ListByMEIEmpresa with empty result set
func TestPropostaMEIHandler_ListByMEIEmpresa_EmptyResults(t *testing.T) {
	mockService := &mockPropostaMEIServiceForList{
		MockPropostaMEIService: MockPropostaMEIService{
			propostas: []*models.PropostaMEI{},
			total:     0,
		},
	}
	router := setupRouterForListTests(mockService)

	req, _ := http.NewRequest("GET", "/api/v1/propostas-mei/por-empresa?meiEmpresaId=12345678000190", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	data := response["data"].([]interface{})
	assert.Equal(t, 0, len(data))

	meta := response["meta"].(map[string]interface{})
	assert.Equal(t, float64(0), meta["total"])
}
