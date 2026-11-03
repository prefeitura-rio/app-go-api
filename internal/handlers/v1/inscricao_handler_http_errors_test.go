package v1_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	v1 "github.com/prefeitura-rio/app-go-api/internal/handlers/v1"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

// DEV-49: falhas de banco nas inscrições não chegam ao cliente e saem com o
// status certo (antes, várias respondiam 400 com o texto do driver).

var inscricaoDBLeak = fmt.Errorf("erro ao verificar inscrição: %w", &pgconn.PgError{
	Code:    "08006",
	Message: `connection failure on relation "inscricoes"`,
})

func assertNoInscricaoLeak(t *testing.T, body string) {
	t.Helper()
	for _, marker := range []string{"SQLSTATE", "relation", "inscricoes", "connection failure", "erro ao verificar"} {
		assert.NotContains(t, body, marker, "o corpo vaza detalhe interno")
	}
}

func adminInscricaoRouter(h *v1.InscricaoHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(middlewares.UserCPFKey, "12345678901")
		c.Set(middlewares.UserRoleKey, "ADMIN")
		c.Next()
	})
	r.PUT("/courses/:courseId/enrollments/:enrollmentId", h.Update)
	r.DELETE("/courses/:courseId/enrollments/:enrollmentId", h.Delete)
	r.POST("/courses/:courseId/enrollments", h.Create)
	return r
}

func doInscricaoReq(r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestInscricaoHandler_Update_GetByIDDBError_Returns500WithoutLeak(t *testing.T) {
	id := uuid.New()
	svc := new(MockInscricaoService)
	svc.On("GetByID", mock.Anything, id).Return(nil, inscricaoDBLeak)

	w := doInscricaoReq(adminInscricaoRouter(v1.NewInscricaoHandler(svc, nil, nil)), http.MethodPut, "/courses/1/enrollments/"+id.String(), `{}`)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.JSONEq(t, `{"error":"Erro ao buscar inscrição"}`, w.Body.String())
	assertNoInscricaoLeak(t, w.Body.String())
}

func TestInscricaoHandler_Update_ServiceErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantBody   string
	}{
		{"falha de banco é 500 genérico", inscricaoDBLeak, http.StatusInternalServerError, `{"error":"Erro ao atualizar inscrição"}`},
		{"inscrição removida no meio é 404", services.ErrInscricaoNaoEncontrada, http.StatusNotFound, `{"error":"inscrição não encontrada"}`},
		{"outro curso é 400", services.ErrInscricaoDeOutroCurso, http.StatusBadRequest, `{"error":"inscrição não pertence ao curso especificado"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := uuid.New()
			svc := new(MockInscricaoService)
			svc.On("GetByID", mock.Anything, id).Return(&models.Inscricao{ID: id, CursoID: 1, CPF: "12345678901"}, nil)
			svc.On("UpdateInscricao", mock.Anything, id, 1, mock.Anything).Return(tt.err)

			w := doInscricaoReq(adminInscricaoRouter(v1.NewInscricaoHandler(svc, nil, nil)), http.MethodPut, "/courses/1/enrollments/"+id.String(), `{}`)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.JSONEq(t, tt.wantBody, w.Body.String())
			assertNoInscricaoLeak(t, w.Body.String())
		})
	}
}

func TestInscricaoHandler_Delete_GetByIDDBError_Returns500WithoutLeak(t *testing.T) {
	id := uuid.New()
	svc := new(MockInscricaoService)
	svc.On("GetByID", mock.Anything, id).Return(nil, inscricaoDBLeak)

	w := doInscricaoReq(adminInscricaoRouter(v1.NewInscricaoHandler(svc, nil, nil)), http.MethodDelete, "/courses/1/enrollments/"+id.String(), "")

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.JSONEq(t, `{"error":"Erro ao verificar inscrição"}`, w.Body.String())
}

func TestInscricaoHandler_Create_ErrorMapping(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantBody   string
	}{
		{"duplicada é 409", services.ErrCPFJaInscrito, http.StatusConflict, `{"error":"CPF já inscrito neste curso"}`},
		{"curso inexistente é 404", services.ErrCursoNaoEncontrado, http.StatusNotFound, `{"error":"curso não encontrado"}`},
		{"regra de negócio é 400", services.NewEnrollmentRuleError("período de inscrições já encerrou"), http.StatusBadRequest, `{"error":"período de inscrições já encerrou"}`},
		{"falha de banco é 500 genérico", inscricaoDBLeak, http.StatusInternalServerError, `{"error":"Erro ao criar inscrição"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(MockInscricaoService)
			svc.On("CreateByAdmin", mock.Anything, mock.Anything).Return(tt.err)

			w := doInscricaoReq(adminInscricaoRouter(v1.NewInscricaoHandler(svc, nil, nil)), http.MethodPost, "/courses/1/enrollments", `{"cpf":"12345678901","name":"Cidadão"}`)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.JSONEq(t, tt.wantBody, w.Body.String())
			assertNoInscricaoLeak(t, w.Body.String())
		})
	}
}

func TestInscricaoHandler_ChangeSchedule_UnexpectedError_DoesNotLeak(t *testing.T) {
	gin.SetMode(gin.TestMode)
	id := uuid.New()
	svc := new(MockInscricaoService)
	svc.On("ChangeSchedule", mock.Anything, id, "12345678901", mock.Anything).Return(nil, inscricaoDBLeak)

	h := v1.NewInscricaoHandler(svc, nil, nil)
	r := gin.New()
	r.PUT("/enrollments/:enrollmentId/schedule", func(c *gin.Context) {
		c.Set(middlewares.UserCPFKey, "12345678901")
		h.ChangeSchedule(c)
	})

	w := doInscricaoReq(r, http.MethodPut, "/enrollments/"+id.String()+"/schedule", `{"enrolled_unit":{"id":"`+uuid.New().String()+`"}}`)

	assert.Equal(t, http.StatusInternalServerError, w.Code, w.Body.String())
	assert.JSONEq(t, `{"error":"Erro ao trocar de turma"}`, w.Body.String())
	assertNoInscricaoLeak(t, w.Body.String())
}
