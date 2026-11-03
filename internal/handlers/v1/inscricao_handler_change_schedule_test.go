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
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

// DEV-136: a troca de turma usa só o schedule_id; o enrolled_unit do corpo não
// é mais obrigatório (o servidor o monta), e as recusas de regra da turma
// (outro curso, turma fechada, sem schedule_id) respondem 400, não 500.

func changeScheduleRouter(h *v1.InscricaoHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.PUT("/enrollments/:enrollmentId/schedule", func(c *gin.Context) {
		c.Set(middlewares.UserCPFKey, "12345678901")
		h.ChangeSchedule(c)
	})
	return r
}

func TestInscricaoHandler_ChangeSchedule_ScheduleIDOnly_ReachesService(t *testing.T) {
	id := uuid.New()
	scheduleID := uuid.New()
	svc := new(MockInscricaoService)
	svc.On("ChangeSchedule", mock.Anything, id, "12345678901", mock.MatchedBy(func(req *models.ScheduleChangeRequest) bool {
		return req.ScheduleID != nil && *req.ScheduleID == scheduleID && req.EnrolledUnit == nil
	})).Return(&models.Inscricao{ID: id, ScheduleID: &scheduleID}, nil)

	w := doInscricaoReq(changeScheduleRouter(v1.NewInscricaoHandler(svc, nil, nil)), http.MethodPut,
		"/enrollments/"+id.String()+"/schedule", `{"schedule_id":"`+scheduleID.String()+`"}`)

	assert.Equal(t, http.StatusOK, w.Code, w.Body.String())
	assert.Contains(t, w.Body.String(), "Turma alterada com sucesso")
	svc.AssertExpectations(t)
}

func TestInscricaoHandler_ChangeSchedule_RuleError_Returns400(t *testing.T) {
	tests := []struct {
		name string
		msg  string
	}{
		{"turma_de_outro_curso", "a turma selecionada não pertence a este curso"},
		{"turma_fechada", "esta turma não está aceitando inscrições no momento"},
		{"sem_schedule_id", "não é possível trocar de turma: selecione a nova turma"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := uuid.New()
			svc := new(MockInscricaoService)
			svc.On("ChangeSchedule", mock.Anything, id, "12345678901", mock.Anything).
				Return(nil, services.NewEnrollmentRuleError(tt.msg))

			w := doInscricaoReq(changeScheduleRouter(v1.NewInscricaoHandler(svc, nil, nil)), http.MethodPut,
				"/enrollments/"+id.String()+"/schedule",
				`{"schedule_id":"`+uuid.New().String()+`","enrolled_unit":{"id":"`+uuid.New().String()+`"}}`)

			assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
			assert.JSONEq(t, `{"error":"`+tt.msg+`"}`, w.Body.String())
			svc.AssertExpectations(t)
		})
	}
}
