package v1_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	v1 "github.com/prefeitura-rio/app-go-api/internal/handlers/v1"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	"github.com/prefeitura-rio/app-go-api/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// citizenRouter wires Create behind a context that looks like a plain citizen
// (CPF present, no admin role), which is the path the portal exercises.
func citizenRouter(h *v1.InscricaoHandler, role string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/v1/courses/:courseId/enrollments", func(c *gin.Context) {
		c.Set(middlewares.UserCPFKey, "12345678901")
		c.Set(middlewares.UserRoleKey, role)
		h.Create(c)
	})
	return r
}

func postEnrollment(r *gin.Engine) *httptest.ResponseRecorder {
	body := []byte(`{"cpf":"12345678901","name":"Cidadão"}`)
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/courses/2865/enrollments",
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// A turma outside its enrollment window is the citizen's problem, not a server
// fault: the portal needs a 4xx carrying the reason so it can show it, instead
// of a 500 that collapses into a generic "Erro ao inscrever-se no curso".
func TestInscricaoHandler_Create_EnrollmentRuleReturns400(t *testing.T) {
	mockSvc := new(MockInscricaoService)
	mockSvc.On("Create", mock.Anything, mock.Anything).Return(
		services.NewEnrollmentRuleError("as inscrições desta turma já encerraram"),
	)

	w := postEnrollment(citizenRouter(v1.NewInscricaoHandler(mockSvc, nil, nil), "USER"))

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "as inscrições desta turma já encerraram")
	// The message must survive verbatim — no "Erro ao criar inscrição:" prefix
	assert.NotContains(t, w.Body.String(), "Erro ao criar inscrição")
}

// Genuine faults keep answering 500 so they stay visible as incidents.
func TestInscricaoHandler_Create_UnexpectedErrorStays500(t *testing.T) {
	mockSvc := new(MockInscricaoService)
	mockSvc.On("Create", mock.Anything, mock.Anything).Return(errors.New("connection refused"))

	w := postEnrollment(citizenRouter(v1.NewInscricaoHandler(mockSvc, nil, nil), "USER"))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// Already-enrolled stays a 409, unchanged by the 400 mapping.
func TestInscricaoHandler_Create_DuplicateStays409(t *testing.T) {
	mockSvc := new(MockInscricaoService)
	mockSvc.On("Create", mock.Anything, mock.Anything).Return(
		errors.New("CPF já inscrito neste curso"),
	)

	w := postEnrollment(citizenRouter(v1.NewInscricaoHandler(mockSvc, nil, nil), "USER"))

	assert.Equal(t, http.StatusConflict, w.Code)
}

// Documents a sharp edge found while reproducing the bug: a token carrying
// go:admin is routed to CreateByAdmin on this same citizen endpoint, and
// CreateByAdmin does NOT enforce the turma enrollment window. A QA account with
// that role therefore cannot reproduce the citizen behaviour here.
func TestInscricaoHandler_Create_AdminRoleBypassesCitizenPath(t *testing.T) {
	mockSvc := new(MockInscricaoService)
	mockSvc.On("CreateByAdmin", mock.Anything, mock.Anything).Return(nil)

	w := postEnrollment(citizenRouter(v1.NewInscricaoHandler(mockSvc, nil, nil), "ADMIN"))

	assert.Equal(t, http.StatusCreated, w.Code)
	mockSvc.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	mockSvc.AssertCalled(t, "CreateByAdmin", mock.Anything, mock.Anything)
}
