package v1_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/prefeitura-rio/app-go-api/internal/apperrors"
	v1 "github.com/prefeitura-rio/app-go-api/internal/handlers/v1"
	"github.com/prefeitura-rio/app-go-api/internal/models"
)

// DEV-49: os handlers de cursos respondem com o status certo e sem o texto do
// banco no corpo.

// courseDBLeak imita uma falha de banco sem constraint conhecida.
var courseDBLeak = fmt.Errorf("erro ao buscar curso: %w", &pgconn.PgError{
	Code:    "57014",
	Message: `canceling statement due to statement timeout on relation "cursos"`,
})

var courseLeakMarkers = []string{"SQLSTATE", "statement", "relation", "erro ao buscar curso", "pgconn"}

func assertNoCourseLeak(t *testing.T, body string) {
	t.Helper()
	for _, marker := range courseLeakMarkers {
		assert.NotContains(t, body, marker, "o corpo vaza detalhe interno")
	}
}

func newCourseErrRouter(svc *MockCursoService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := v1.NewCourseHandler(svc, new(MockInscricaoServiceForCourse), new(MockCursoRepositoryForCourseHandler))
	r := gin.New()
	r.POST("/courses", h.Create)
	r.GET("/courses/:courseId", h.GetByID)
	r.PUT("/courses/:courseId", h.Update)
	r.DELETE("/courses/:courseId", h.Delete)
	r.GET("/courses", h.List)
	return r
}

func doCourseReq(r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestCourseHandler_Delete_NotFound_Returns404(t *testing.T) {
	svc := new(MockCursoService)
	svc.On("Delete", mock.Anything, 999).Return(apperrors.NotFound("Curso não encontrado"))

	w := doCourseReq(newCourseErrRouter(svc), http.MethodDelete, "/courses/999", "")

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.JSONEq(t, `{"error":"Curso não encontrado"}`, w.Body.String())
	svc.AssertExpectations(t)
}

func TestCourseHandler_GetByID_NotFound_Returns404(t *testing.T) {
	svc := new(MockCursoService)
	svc.On("GetByID", mock.Anything, 999).Return(nil, nil)

	w := doCourseReq(newCourseErrRouter(svc), http.MethodGet, "/courses/999", "")

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCourseHandler_Update_DuplicateKey_Returns409(t *testing.T) {
	svc := new(MockCursoService)
	svc.On("GetByID", mock.Anything, 1).Return(&models.Curso{ID: 1, Titulo: "Curso"}, nil)
	svc.On("Update", mock.Anything, mock.Anything).Return(fmt.Errorf("erro ao atualizar curso: %w", &pgconn.PgError{
		Code:           "23505",
		ConstraintName: "uni_cursos_nome",
		Message:        `duplicate key value violates unique constraint "uni_cursos_nome"`,
	}))

	w := doCourseReq(newCourseErrRouter(svc), http.MethodPut, "/courses/1", `{"title":"Curso"}`)

	assert.Equal(t, http.StatusConflict, w.Code, w.Body.String())
	assert.NotContains(t, w.Body.String(), "uni_cursos_nome")
	assertNoCourseLeak(t, w.Body.String())
}

func TestCourseHandler_InvalidPayload_Returns400(t *testing.T) {
	svc := new(MockCursoService)
	svc.On("GetByID", mock.Anything, 1).Return(&models.Curso{ID: 1}, nil)
	r := newCourseErrRouter(svc)

	for _, tc := range []struct{ method, path string }{
		{http.MethodPost, "/courses"},
		{http.MethodPut, "/courses/1"},
	} {
		t.Run(tc.method, func(t *testing.T) {
			w := doCourseReq(r, tc.method, tc.path, `{"title": `)
			assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
		})
	}
}

func TestCourseHandler_DBError_Returns500WithoutLeak(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*MockCursoService)
		method      string
		path        string
		body        string
		wantMessage string
	}{
		{
			name:        "GetByID",
			setup:       func(s *MockCursoService) { s.On("GetByID", mock.Anything, 1).Return(nil, courseDBLeak) },
			method:      http.MethodGet,
			path:        "/courses/1",
			wantMessage: "Erro ao buscar curso",
		},
		{
			name:        "Update busca o curso",
			setup:       func(s *MockCursoService) { s.On("GetByID", mock.Anything, 1).Return(nil, courseDBLeak) },
			method:      http.MethodPut,
			path:        "/courses/1",
			body:        `{"title":"x"}`,
			wantMessage: "Erro ao buscar curso",
		},
		{
			name:        "Delete",
			setup:       func(s *MockCursoService) { s.On("Delete", mock.Anything, 1).Return(courseDBLeak) },
			method:      http.MethodDelete,
			path:        "/courses/1",
			wantMessage: "Erro ao excluir curso",
		},
		{
			name: "List",
			setup: func(s *MockCursoService) {
				s.On("List", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, 0, courseDBLeak)
			},
			method:      http.MethodGet,
			path:        "/courses",
			wantMessage: "Erro ao listar cursos",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(MockCursoService)
			tt.setup(svc)

			w := doCourseReq(newCourseErrRouter(svc), tt.method, tt.path, tt.body)

			assert.Equal(t, http.StatusInternalServerError, w.Code, w.Body.String())
			assert.JSONEq(t, `{"error":"`+tt.wantMessage+`"}`, w.Body.String())
			assertNoCourseLeak(t, w.Body.String())
		})
	}
}

// DEV-119: o PUT que removeria turma com inscrição responde 409 com a mensagem
// do domínio, não 500.
func TestCourseHandler_Update_ScheduleWithEnrollments_Returns409(t *testing.T) {
	const msg = "Não é possível remover turmas que têm inscrições: a turma 3d901bcb-a112-44f1-b71d-b75e4329ca00 tem 2 inscrição(ões)."
	svc := new(MockCursoService)
	svc.On("GetByID", mock.Anything, 1).Return(&models.Curso{ID: 1, Titulo: "Curso"}, nil)
	svc.On("Update", mock.Anything, mock.Anything).
		Return(fmt.Errorf("erro ao atualizar remote class: %w", apperrors.Conflict(msg)))

	w := doCourseReq(newCourseErrRouter(svc), http.MethodPut, "/courses/1", `{"title":"Curso"}`)

	assert.Equal(t, http.StatusConflict, w.Code, w.Body.String())
	assert.JSONEq(t, `{"error":"`+msg+`"}`, w.Body.String())
	svc.AssertExpectations(t)
}
