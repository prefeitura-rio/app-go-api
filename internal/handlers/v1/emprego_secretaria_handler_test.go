package v1_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	v1 "github.com/prefeitura-rio/app-go-api/internal/handlers/v1"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
	"github.com/stretchr/testify/assert"
)

// mockEmpregoRepoWithCapture extends the basic mock to capture the filter passed to List.
type mockEmpregoRepoWithCapture struct {
	entity         *models.Emprego
	createID       int
	createErr      error
	getErr         error
	updateErr      error
	deleteErr      error
	listItems      []*models.Emprego
	listTotal      int
	listErr        error
	capturedFilter map[string]interface{}
}

func (m *mockEmpregoRepoWithCapture) Create(_ context.Context, e *models.Emprego) (int, error) {
	if m.createErr != nil {
		return 0, m.createErr
	}
	return m.createID, nil
}

func (m *mockEmpregoRepoWithCapture) GetByID(_ context.Context, _ int) (*models.Emprego, error) {
	return m.entity, m.getErr
}

func (m *mockEmpregoRepoWithCapture) Update(_ context.Context, _ *models.Emprego) error {
	return m.updateErr
}

func (m *mockEmpregoRepoWithCapture) Delete(_ context.Context, _ int) error {
	return m.deleteErr
}

func (m *mockEmpregoRepoWithCapture) List(_ context.Context, filter map[string]interface{}, _, _ int) ([]*models.Emprego, int, error) {
	m.capturedFilter = filter
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	if m.listItems == nil {
		return []*models.Emprego{}, 0, nil
	}
	return m.listItems, m.listTotal, nil
}

// setupEmpregoRouterWithSecretaria builds a gin router with user context middleware injected.
// role: "ADMIN" or "USER"; secretariaIDs: nil means middleware never ran, []string{} means ran but empty.
func setupEmpregoRouterWithSecretaria(repo services.EmpregoRepositoryInterface, role string, secretariaIDs []string) *gin.Engine {
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
	svc := services.NewEmpregoServiceWithInterface(repo)
	h := v1.NewEmpregoHandler(svc)
	r.POST("/api/v1/empregos", h.Create)
	r.GET("/api/v1/empregos", h.List)
	r.GET("/api/v1/empregos/:id", h.GetByID)
	r.PUT("/api/v1/empregos/:id", h.Update)
	r.DELETE("/api/v1/empregos/:id", h.Delete)
	return r
}

// --- List ---

func TestEmpregoHandler_Secretaria_List_FilterApplied(t *testing.T) {
	repo := &mockEmpregoRepoWithCapture{listItems: []*models.Emprego{}}
	router := setupEmpregoRouterWithSecretaria(repo, "USER", []string{"orgao-1", "orgao-2"})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/empregos", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	ids, ok := repo.capturedFilter["orgao_id IN"].([]string)
	assert.True(t, ok, "expected orgao_id IN filter")
	assert.Equal(t, []string{"orgao-1", "orgao-2"}, ids)
}

func TestEmpregoHandler_Secretaria_List_EmptySecretaria_NoResults(t *testing.T) {
	repo := &mockEmpregoRepoWithCapture{listItems: []*models.Emprego{}}
	// middleware ran but returned empty list → filter must still be set
	router := setupEmpregoRouterWithSecretaria(repo, "USER", []string{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/empregos", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	_, hasFilter := repo.capturedFilter["orgao_id IN"]
	assert.True(t, hasFilter, "expected orgao_id IN filter even for empty list")
}

func TestEmpregoHandler_Secretaria_List_Admin_NoSecretariaFilter(t *testing.T) {
	repo := &mockEmpregoRepoWithCapture{listItems: []*models.Emprego{}}
	router := setupEmpregoRouterWithSecretaria(repo, "ADMIN", []string{"orgao-1"})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/empregos", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	_, hasFilter := repo.capturedFilter["orgao_id IN"]
	assert.False(t, hasFilter, "admin should not have orgao_id IN filter")
}

func TestEmpregoHandler_Secretaria_List_NoMiddleware_NoFilter(t *testing.T) {
	repo := &mockEmpregoRepoWithCapture{listItems: []*models.Emprego{}}
	// nil secretariaIDs → middleware never ran → no filter applied
	router := setupEmpregoRouterWithSecretaria(repo, "USER", nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/empregos", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	_, hasFilter := repo.capturedFilter["orgao_id IN"]
	assert.False(t, hasFilter, "no filter expected when middleware did not run")
}

// --- Create ---

func TestEmpregoHandler_Secretaria_Create_OrgaoForced(t *testing.T) {
	repo := &mockEmpregoRepoWithCapture{createID: 1}
	router := setupEmpregoRouterWithSecretaria(repo, "USER", []string{"orgao-secretaria"})

	body := []byte(`{"titulo": "Vaga Teste", "status": "ABERTO", "tipo_contratacao": "CLT"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/empregos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "orgao-secretaria")
}

func TestEmpregoHandler_Secretaria_Create_EmptyList_Forbidden(t *testing.T) {
	repo := &mockEmpregoRepoWithCapture{}
	router := setupEmpregoRouterWithSecretaria(repo, "USER", []string{})

	body := []byte(`{"titulo": "Vaga Teste"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/empregos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "nenhuma secretaria associada")
}

func TestEmpregoHandler_Secretaria_Create_Admin_NotForced(t *testing.T) {
	repo := &mockEmpregoRepoWithCapture{createID: 1}
	router := setupEmpregoRouterWithSecretaria(repo, "ADMIN", []string{"orgao-secretaria"})

	body := []byte(`{"titulo": "Vaga Admin", "orgao_id": "orgao-personalizado", "tipo_contratacao": "CLT", "status": "ABERTO"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/empregos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	// admin orgao_id should not be overridden by secretaria
	assert.Contains(t, w.Body.String(), "orgao-personalizado")
}

// --- Update ---

func TestEmpregoHandler_Secretaria_Update_AllowedOrgao(t *testing.T) {
	repo := &mockEmpregoRepoWithCapture{
		entity: &models.Emprego{ID: 1, OrgaoID: "orgao-1"},
	}
	router := setupEmpregoRouterWithSecretaria(repo, "USER", []string{"orgao-1"})

	body := []byte(`{"titulo": "Atualizado", "tipo_contratacao": "CLT", "status": "ABERTO"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/empregos/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEmpregoHandler_Secretaria_Update_DeniedOrgao(t *testing.T) {
	repo := &mockEmpregoRepoWithCapture{
		entity: &models.Emprego{ID: 1, OrgaoID: "orgao-outra-secretaria"},
	}
	router := setupEmpregoRouterWithSecretaria(repo, "USER", []string{"orgao-minha-secretaria"})

	body := []byte(`{"titulo": "Tentativa", "tipo_contratacao": "CLT", "status": "ABERTO"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/empregos/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "não pertence à sua secretaria")
}

func TestEmpregoHandler_Secretaria_Update_EmptySecretaria_Denied(t *testing.T) {
	repo := &mockEmpregoRepoWithCapture{
		entity: &models.Emprego{ID: 1, OrgaoID: "orgao-1"},
	}
	router := setupEmpregoRouterWithSecretaria(repo, "USER", []string{})

	body := []byte(`{"titulo": "Tentativa"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/empregos/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestEmpregoHandler_Secretaria_Update_MultipleOrgaos_Allowed(t *testing.T) {
	repo := &mockEmpregoRepoWithCapture{
		entity: &models.Emprego{ID: 1, OrgaoID: "orgao-2"},
	}
	router := setupEmpregoRouterWithSecretaria(repo, "USER", []string{"orgao-1", "orgao-2", "orgao-3"})

	body := []byte(`{"titulo": "Atualizado", "tipo_contratacao": "CLT", "status": "ABERTO"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/empregos/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEmpregoHandler_Secretaria_Update_Admin_Bypasses(t *testing.T) {
	repo := &mockEmpregoRepoWithCapture{}
	router := setupEmpregoRouterWithSecretaria(repo, "ADMIN", []string{"orgao-diferente"})

	body := []byte(`{"titulo": "Admin update"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/empregos/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// admin bypasses ownership check (GetByID returns nil → 404, but no 403)
	assert.NotEqual(t, http.StatusForbidden, w.Code)
}

// --- Delete ---

func TestEmpregoHandler_Secretaria_Delete_AllowedOrgao(t *testing.T) {
	repo := &mockEmpregoRepoWithCapture{
		entity: &models.Emprego{ID: 1, OrgaoID: "orgao-1"},
	}
	router := setupEmpregoRouterWithSecretaria(repo, "USER", []string{"orgao-1"})

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/empregos/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "excluído com sucesso")
}

func TestEmpregoHandler_Secretaria_Delete_DeniedOrgao(t *testing.T) {
	repo := &mockEmpregoRepoWithCapture{
		entity: &models.Emprego{ID: 1, OrgaoID: "orgao-outra"},
	}
	router := setupEmpregoRouterWithSecretaria(repo, "USER", []string{"orgao-minha"})

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/empregos/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "não pertence à sua secretaria")
}

func TestEmpregoHandler_Secretaria_Delete_EmptySecretaria_Denied(t *testing.T) {
	repo := &mockEmpregoRepoWithCapture{
		entity: &models.Emprego{ID: 1, OrgaoID: "orgao-1"},
	}
	router := setupEmpregoRouterWithSecretaria(repo, "USER", []string{})

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/empregos/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestEmpregoHandler_Secretaria_Delete_Admin_Bypasses(t *testing.T) {
	repo := &mockEmpregoRepoWithCapture{
		entity: &models.Emprego{ID: 1, OrgaoID: "orgao-qualquer"},
	}
	router := setupEmpregoRouterWithSecretaria(repo, "ADMIN", []string{"orgao-diferente"})

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/empregos/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
