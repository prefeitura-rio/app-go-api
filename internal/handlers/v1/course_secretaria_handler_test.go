package v1_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	v1 "github.com/prefeitura-rio/app-go-api/internal/handlers/v1"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// setupCourseRouterWithSecretaria builds a gin router with user context middleware injected.
// role: "ADMIN" or "USER"; secretariaIDs: nil means middleware never ran, []string{} means ran but empty.
func setupCourseRouterWithSecretaria(svc *MockCursoService, role string, secretariaIDs []string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if role != "" {
			c.Set(middlewares.UserRoleKey, role)
		}
		if secretariaIDs != nil {
			c.Set(middlewares.UserSecretariaOrgaoIDsKey, secretariaIDs)
		}
		c.Next()
	})
	h := v1.NewCourseHandler(svc, nil, nil)
	r.POST("/api/v1/courses", h.Create)
	r.POST("/api/v1/courses/draft", h.CreateDraft)
	r.GET("/api/v1/courses", h.List)
	r.GET("/api/v1/courses/:courseId", h.GetByID)
	r.PUT("/api/v1/courses/:courseId", h.Update)
	r.DELETE("/api/v1/courses/:courseId", h.Delete)
	return r
}

// --- List ---

func TestCourseHandler_Secretaria_List_FilterApplied(t *testing.T) {
	svc := &MockCursoService{}
	svc.On("List", mock.Anything, mock.MatchedBy(func(f map[string]interface{}) bool {
		ids, ok := f["orgao_id IN"].([]string)
		return ok && len(ids) == 2 && ids[0] == "orgao-1"
	}), mock.Anything, mock.Anything).Return([]*models.Curso{}, 0, nil)

	router := setupCourseRouterWithSecretaria(svc, "USER", []string{"orgao-1", "orgao-2"})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestCourseHandler_Secretaria_List_EmptySecretaria_FilterStillSet(t *testing.T) {
	svc := &MockCursoService{}
	svc.On("List", mock.Anything, mock.MatchedBy(func(f map[string]interface{}) bool {
		_, hasKey := f["orgao_id IN"]
		return hasKey
	}), mock.Anything, mock.Anything).Return([]*models.Curso{}, 0, nil)

	router := setupCourseRouterWithSecretaria(svc, "USER", []string{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestCourseHandler_Secretaria_List_Admin_NoSecretariaFilter(t *testing.T) {
	svc := &MockCursoService{}
	svc.On("List", mock.Anything, mock.MatchedBy(func(f map[string]interface{}) bool {
		_, hasKey := f["orgao_id IN"]
		return !hasKey
	}), mock.Anything, mock.Anything).Return([]*models.Curso{}, 0, nil)

	router := setupCourseRouterWithSecretaria(svc, "ADMIN", []string{"orgao-1"})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestCourseHandler_Secretaria_List_NoMiddleware_NoFilter(t *testing.T) {
	svc := &MockCursoService{}
	svc.On("List", mock.Anything, mock.MatchedBy(func(f map[string]interface{}) bool {
		_, hasKey := f["orgao_id IN"]
		return !hasKey
	}), mock.Anything, mock.Anything).Return([]*models.Curso{}, 0, nil)

	// nil secretariaIDs → middleware never ran
	router := setupCourseRouterWithSecretaria(svc, "USER", nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

// --- Create ---

func TestCourseHandler_Secretaria_Create_OrgaoForced(t *testing.T) {
	svc := &MockCursoService{}
	svc.On("Create", mock.Anything, mock.MatchedBy(func(c *models.Curso) bool {
		return c.OrgaoID == "orgao-secretaria"
	})).Return(1, nil)

	router := setupCourseRouterWithSecretaria(svc, "USER", []string{"orgao-secretaria"})

	body := []byte(`{"titulo": "Curso Teste", "descricao": "Desc"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	svc.AssertExpectations(t)
}

func TestCourseHandler_Secretaria_Create_EmptyList_Forbidden(t *testing.T) {
	svc := &MockCursoService{}

	router := setupCourseRouterWithSecretaria(svc, "USER", []string{})

	body := []byte(`{"titulo": "Curso Teste"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "nenhuma secretaria associada")
	svc.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestCourseHandler_Secretaria_CreateDraft_OrgaoForced(t *testing.T) {
	svc := &MockCursoService{}
	svc.On("Create", mock.Anything, mock.MatchedBy(func(c *models.Curso) bool {
		return c.OrgaoID == "orgao-draft"
	})).Return(2, nil)

	router := setupCourseRouterWithSecretaria(svc, "USER", []string{"orgao-draft"})

	body := []byte(`{"titulo": "Rascunho"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses/draft", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	svc.AssertExpectations(t)
}

func TestCourseHandler_Secretaria_CreateDraft_EmptyList_Forbidden(t *testing.T) {
	svc := &MockCursoService{}

	router := setupCourseRouterWithSecretaria(svc, "USER", []string{})

	body := []byte(`{"titulo": "Rascunho"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses/draft", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCourseHandler_Secretaria_Create_Admin_NotForced(t *testing.T) {
	svc := &MockCursoService{}
	svc.On("Create", mock.Anything, mock.MatchedBy(func(c *models.Curso) bool {
		return c.OrgaoID == "orgao-admin"
	})).Return(1, nil)

	router := setupCourseRouterWithSecretaria(svc, "ADMIN", []string{"orgao-secretaria"})

	body := []byte(`{"titulo": "Curso Admin", "orgao_id": "orgao-admin"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	svc.AssertExpectations(t)
}

// --- Update ---

func TestCourseHandler_Secretaria_Update_AllowedOrgao(t *testing.T) {
	svc := &MockCursoService{}
	existing := &models.Curso{ID: 1, OrgaoID: "orgao-1"}
	svc.On("GetByID", mock.Anything, 1).Return(existing, nil)
	svc.On("Update", mock.Anything, mock.Anything).Return(nil)

	router := setupCourseRouterWithSecretaria(svc, "USER", []string{"orgao-1"})

	body := []byte(`{"titulo": "Atualizado"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/courses/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestCourseHandler_Secretaria_Update_DeniedOrgao(t *testing.T) {
	svc := &MockCursoService{}
	existing := &models.Curso{ID: 1, OrgaoID: "orgao-outra-secretaria"}
	svc.On("GetByID", mock.Anything, 1).Return(existing, nil)

	router := setupCourseRouterWithSecretaria(svc, "USER", []string{"orgao-minha"})

	body := []byte(`{"titulo": "Tentativa"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/courses/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "não pertence à sua secretaria")
	svc.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestCourseHandler_Secretaria_Update_EmptySecretaria_Denied(t *testing.T) {
	svc := &MockCursoService{}
	existing := &models.Curso{ID: 1, OrgaoID: "orgao-1"}
	svc.On("GetByID", mock.Anything, 1).Return(existing, nil)

	router := setupCourseRouterWithSecretaria(svc, "USER", []string{})

	body := []byte(`{"titulo": "Tentativa"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/courses/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCourseHandler_Secretaria_Update_MultipleOrgaos_Allowed(t *testing.T) {
	svc := &MockCursoService{}
	existing := &models.Curso{ID: 1, OrgaoID: "orgao-2"}
	svc.On("GetByID", mock.Anything, 1).Return(existing, nil)
	svc.On("Update", mock.Anything, mock.Anything).Return(nil)

	router := setupCourseRouterWithSecretaria(svc, "USER", []string{"orgao-1", "orgao-2", "orgao-3"})

	body := []byte(`{"titulo": "Atualizado"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/courses/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestCourseHandler_Secretaria_Update_Admin_Bypasses(t *testing.T) {
	svc := &MockCursoService{}
	existing := &models.Curso{ID: 1, OrgaoID: "orgao-qualquer"}
	svc.On("GetByID", mock.Anything, 1).Return(existing, nil)
	svc.On("Update", mock.Anything, mock.Anything).Return(nil)

	router := setupCourseRouterWithSecretaria(svc, "ADMIN", []string{"orgao-diferente"})

	body := []byte(`{"titulo": "Admin update"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/courses/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusForbidden, w.Code)
}

// --- Delete ---

func TestCourseHandler_Secretaria_Delete_AllowedOrgao(t *testing.T) {
	svc := &MockCursoService{}
	existing := &models.Curso{ID: 1, OrgaoID: "orgao-1"}
	svc.On("GetByID", mock.Anything, 1).Return(existing, nil)
	svc.On("Delete", mock.Anything, 1).Return(nil)

	router := setupCourseRouterWithSecretaria(svc, "USER", []string{"orgao-1"})

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/courses/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestCourseHandler_Secretaria_Delete_DeniedOrgao(t *testing.T) {
	svc := &MockCursoService{}
	existing := &models.Curso{ID: 1, OrgaoID: "orgao-outra"}
	svc.On("GetByID", mock.Anything, 1).Return(existing, nil)

	router := setupCourseRouterWithSecretaria(svc, "USER", []string{"orgao-minha"})

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/courses/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "não pertence à sua secretaria")
	svc.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

func TestCourseHandler_Secretaria_Delete_EmptySecretaria_Denied(t *testing.T) {
	svc := &MockCursoService{}
	existing := &models.Curso{ID: 1, OrgaoID: "orgao-1"}
	svc.On("GetByID", mock.Anything, 1).Return(existing, nil)

	router := setupCourseRouterWithSecretaria(svc, "USER", []string{})

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/courses/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCourseHandler_Secretaria_Delete_Admin_Bypasses(t *testing.T) {
	svc := &MockCursoService{}
	existing := &models.Curso{ID: 1, OrgaoID: "orgao-qualquer"}
	svc.On("GetByID", mock.Anything, 1).Return(existing, nil)
	svc.On("Delete", mock.Anything, 1).Return(nil)

	router := setupCourseRouterWithSecretaria(svc, "ADMIN", []string{"orgao-diferente"})

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/courses/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusForbidden, w.Code)
}
