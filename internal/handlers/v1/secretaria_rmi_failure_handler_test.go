package v1_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	v1 "github.com/prefeitura-rio/app-go-api/internal/handlers/v1"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// failingSecretariaResolver simulates app-rmi being down (DEV-25).
func failingSecretariaResolver(_ context.Context, _ string) ([]string, error) {
	return nil, errors.New("RMI API returned status 503: unavailable")
}

// setupEmpregoRouterRMIDown wires the real ExtractSecretariaOrgaoIDs middleware with a
// resolver that always fails, in front of the emprego handler.
func setupEmpregoRouterRMIDown(repo services.EmpregoRepositoryInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(middlewares.UserRoleKey, "USER")
		c.Set(middlewares.UserCPFKey, "12345678900")
		c.Next()
	})
	r.Use(middlewares.ExtractSecretariaOrgaoIDs(failingSecretariaResolver))
	h := v1.NewEmpregoHandler(services.NewEmpregoServiceWithInterface(repo))
	r.POST("/api/v1/empregos", h.Create)
	r.GET("/api/v1/empregos", h.List)
	r.PUT("/api/v1/empregos/:id", h.Update)
	r.DELETE("/api/v1/empregos/:id", h.Delete)
	return r
}

func TestEmpregoHandler_RMIDown_List_FailOpen_FullList(t *testing.T) {
	repo := &mockEmpregoRepoWithCapture{
		listItems: []*models.Emprego{{ID: 1, Titulo: "Vaga A"}, {ID: 2, Titulo: "Vaga B"}},
		listTotal: 2,
	}
	router := setupEmpregoRouterRMIDown(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/empregos", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	_, hasFilter := repo.capturedFilter["orgao_id IN"]
	assert.False(t, hasFilter, "RMI down must not restrict the read to an empty secretaria scope")
	assert.Contains(t, w.Body.String(), "Vaga A")
	assert.Contains(t, w.Body.String(), "Vaga B")
}

func TestEmpregoHandler_RMIDown_Writes_FailClosed_Forbidden(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"create", http.MethodPost, "/api/v1/empregos", `{"titulo": "Vaga"}`},
		{"update", http.MethodPut, "/api/v1/empregos/1", `{"titulo": "Vaga"}`},
		{"delete", http.MethodDelete, "/api/v1/empregos/1", ``},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockEmpregoRepoWithCapture{createID: 1, entity: &models.Emprego{ID: 1, OrgaoID: "orgao-1"}}
			router := setupEmpregoRouterRMIDown(repo)

			req := httptest.NewRequest(tt.method, tt.path, bytes.NewReader([]byte(tt.body)))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusForbidden, w.Code)
		})
	}
}

// setupCourseRouterRMIDown mirrors the course chain with the real secretaria middleware
// and a failing resolver. heimdallRoles are the caller's Heimdall roles.
func setupCourseRouterRMIDown(svc *MockCursoService, heimdallRoles ...string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(middlewares.UserRoleKey, "USER")
		c.Set(middlewares.UserCPFKey, "12345678900")
		if len(heimdallRoles) > 0 {
			c.Set(middlewares.UserRolesKey, heimdallRoles)
		}
		c.Next()
	})
	r.Use(middlewares.ExtractSecretariaOrgaoIDs(failingSecretariaResolver))
	h := v1.NewCourseHandler(svc, nil, nil)
	r.POST("/api/v1/courses", middlewares.CourseAuthorization(), middlewares.CourseOrgaoInjector(), h.Create)
	r.POST("/api/v1/courses/draft", middlewares.CourseAuthorization(), middlewares.CourseDraftAuthorization(), h.CreateDraft)
	return r
}

func TestCourseHandler_RMIDown_List_FailOpen_NoSecretariaFilter(t *testing.T) {
	svc := &MockCursoService{}
	svc.On("List", mock.Anything, mock.MatchedBy(func(f map[string]interface{}) bool {
		_, restricted := f["orgao_id IN"]
		_, empty := f["_no_results"]
		return !restricted && !empty
	}), mock.Anything, mock.Anything).Return([]*models.Curso{{ID: 7, Titulo: "Curso Aberto"}}, 1, nil)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middlewares.UserRoleKey, "USER")
		c.Set(middlewares.UserCPFKey, "12345678900")
		c.Next()
	})
	router.Use(middlewares.ExtractSecretariaOrgaoIDs(failingSecretariaResolver))
	var secretariaKeySet bool
	router.GET("/api/v1/courses", func(c *gin.Context) {
		_, secretariaKeySet = c.Get(middlewares.UserSecretariaOrgaoIDsKey)
		c.Next()
	}, v1.NewCourseHandler(svc, nil, nil).List)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/courses", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.False(t, secretariaKeySet, "RMI down must not set the secretaria key on a course read")
	assert.Contains(t, w.Body.String(), "Curso Aberto")
	svc.AssertExpectations(t)
}

func TestCourseHandler_RMIDown_EditorWrites_Forbidden(t *testing.T) {
	for _, path := range []string{"/api/v1/courses", "/api/v1/courses/draft"} {
		t.Run(path, func(t *testing.T) {
			svc := &MockCursoService{}
			router := setupCourseRouterRMIDown(svc, "go:cursos:editor")

			body := []byte(`{"titulo": "Curso", "orgao_id": "orgao-1"}`)
			req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusForbidden, w.Code)
			svc.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
		})
	}
}
