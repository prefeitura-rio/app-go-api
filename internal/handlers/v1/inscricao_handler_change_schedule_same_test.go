package v1_test

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	v1 "github.com/prefeitura-rio/app-go-api/internal/handlers/v1"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

// DEV-141: the refusal of a change to the enrollment's current turma reaches
// the citizen as 400 with its own message, not as "sem vagas" nor as 500.
func TestInscricaoHandler_ChangeSchedule_SameSchedule_Returns400AlreadyEnrolled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	id := uuid.New()
	scheduleID := uuid.New()
	const msg = "não é possível trocar de turma: você já está nesta turma"

	svc := new(MockInscricaoService)
	svc.On("ChangeSchedule", mock.Anything, id, "12345678901", mock.Anything).
		Return(nil, services.NewEnrollmentRuleError(msg))

	h := v1.NewInscricaoHandler(svc, nil, nil)
	r := gin.New()
	r.PUT("/enrollments/:enrollmentId/schedule", func(c *gin.Context) {
		c.Set(middlewares.UserCPFKey, "12345678901")
		h.ChangeSchedule(c)
	})

	w := doInscricaoReq(r, http.MethodPut, "/enrollments/"+id.String()+"/schedule",
		`{"schedule_id":"`+scheduleID.String()+`","enrolled_unit":{"id":"`+uuid.New().String()+`"}}`)

	assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
	assert.JSONEq(t, `{"error":"`+msg+`"}`, w.Body.String())
	assert.NotContains(t, w.Body.String(), "não há vagas")
	svc.AssertExpectations(t)
}
