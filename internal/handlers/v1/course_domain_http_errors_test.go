package v1_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/prefeitura-rio/app-go-api/internal/apperrors"
	v1 "github.com/prefeitura-rio/app-go-api/internal/handlers/v1"
	"github.com/prefeitura-rio/app-go-api/internal/models"
)

func setupCourseDomainRouter(svc *MockCursoService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := v1.NewCourseHandler(svc, nil, nil)
	r.POST("/api/v1/courses", h.Create)
	r.GET("/api/v1/courses/:courseId", h.GetByID)
	r.PUT("/api/v1/courses/:courseId/send-to-review", h.SendToReview)
	return r
}

func TestCourseHandler_DomainErrors_HTTPStatusMapping(t *testing.T) {
	t.Run("not found returns 404", func(t *testing.T) {
		svc := new(MockCursoService)
		svc.On("SendToReview", mock.Anything, 99).Return(apperrors.NotFound("curso não encontrado"))
		r := setupCourseDomainRouter(svc)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/courses/99/send-to-review", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code, "body: %s", w.Body.String())
		assert.Equal(t, "curso não encontrado", decodeCourseError(t, w.Body.Bytes()))
		svc.AssertExpectations(t)
	})

	t.Run("wrong state returns 409", func(t *testing.T) {
		svc := new(MockCursoService)
		svc.On("SendToReview", mock.Anything, 1).Return(
			apperrors.Conflict("curso não está em estado válido para ser enviado para revisão"),
		)
		r := setupCourseDomainRouter(svc)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/courses/1/send-to-review", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code, "body: %s", w.Body.String())
		assert.Contains(t, decodeCourseError(t, w.Body.Bytes()), "não está em estado")
		svc.AssertExpectations(t)
	})

	t.Run("validation returns 400", func(t *testing.T) {
		svc := new(MockCursoService)
		svc.On("Create", mock.Anything, mock.AnythingOfType("*models.Curso")).
			Return(0, apperrors.Validation("título é obrigatório"))
		r := setupCourseDomainRouter(svc)

		body := `{"titulo":"","modalidade":"Presencial"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/courses", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "body: %s", w.Body.String())
		assert.Equal(t, "título é obrigatório", decodeCourseError(t, w.Body.Bytes()))
		svc.AssertExpectations(t)
	})

	t.Run("invalid payload returns 400", func(t *testing.T) {
		svc := new(MockCursoService)
		r := setupCourseDomainRouter(svc)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/courses", strings.NewReader(`{bad`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unique violation returns 409", func(t *testing.T) {
		svc := new(MockCursoService)
		svc.On("Create", mock.Anything, mock.AnythingOfType("*models.Curso")).
			Return(0, &pq.Error{Code: pq.ErrorCode("23505"), Constraint: "cursos_titulo_key"})
		r := setupCourseDomainRouter(svc)

		body := `{"titulo":"Curso X","modalidade":"Presencial"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/courses", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code, "body: %s", w.Body.String())
		assert.NotContains(t, w.Body.String(), "cursos_titulo_key")
		svc.AssertExpectations(t)
	})

	t.Run("unknown error returns 500 without leak", func(t *testing.T) {
		svc := new(MockCursoService)
		svc.On("GetByID", mock.Anything, 1).
			Return((*models.Curso)(nil), errors.New("pq: connection refused SEGREDO-CURSO"))
		r := setupCourseDomainRouter(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/courses/1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "Erro ao buscar curso", decodeCourseError(t, w.Body.Bytes()))
		assert.NotContains(t, w.Body.String(), "SEGREDO-CURSO")
		svc.AssertExpectations(t)
	})
}

func decodeCourseError(t *testing.T, body []byte) string {
	t.Helper()
	var resp map[string]any
	require.NoError(t, json.Unmarshal(body, &resp), "body: %s", string(body))
	msg, _ := resp["error"].(string)
	return msg
}
