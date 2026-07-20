package services_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/prefeitura-rio/app-go-api/internal/config"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

// TestInscricaoService_Create_TurmaEnrollmentWindow covers the per-turma
// enrollment window enforced on the citizen Create flow (enforceWindow=true).
// The course-level window is left open (nil) to isolate the turma check.
func TestInscricaoService_Create_TurmaEnrollmentWindow(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	scheduleID := uuid.New()
	ptrT := func(v time.Time) *time.Time { return &v }

	newSvc := func(schedule models.CourseSchedule) *services.InscricaoService {
		inscricaoRepo := &MockInscricaoRepository{
			CreateFunc:              func(ctx context.Context, i *models.Inscricao) error { return nil },
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) { return false, nil },
		}
		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (string, *time.Time, *time.Time, bool, error) {
				return string(models.StatusCursoOpened), nil, nil, false, nil
			},
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return &models.Curso{
					ID:              1,
					LocationClasses: []models.LocationClass{{Schedules: []models.CourseSchedule{schedule}}},
				}, nil
			},
		}
		return services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})
	}

	accepting := true
	base := models.CourseSchedule{ID: scheduleID, Vacancies: 10, AcceptingEnrollments: &accepting}
	enroll := func(svc *services.InscricaoService) error {
		return svc.Create(ctx, &models.Inscricao{
			CPF: "12345678900", CursoID: 1, Name: "Teste", Email: "t@e.com", ScheduleID: &scheduleID,
		})
	}

	t.Run("turma aberta permite inscrição", func(t *testing.T) {
		s := base
		s.EnrollmentStartDate = ptrT(now.Add(-24 * time.Hour))
		s.EnrollmentEndDate = ptrT(now.Add(24 * time.Hour))
		if err := enroll(newSvc(s)); err != nil {
			t.Fatalf("esperava sucesso, veio erro: %v", err)
		}
	})

	t.Run("turma encerrada bloqueia inscrição", func(t *testing.T) {
		s := base
		s.EnrollmentStartDate = ptrT(now.Add(-48 * time.Hour))
		s.EnrollmentEndDate = ptrT(now.Add(-24 * time.Hour))
		err := enroll(newSvc(s))
		if err == nil || !strings.Contains(err.Error(), "encerraram") {
			t.Fatalf("esperava erro de inscrições encerradas, veio: %v", err)
		}
	})

	t.Run("turma ainda não aberta bloqueia inscrição", func(t *testing.T) {
		s := base
		s.EnrollmentStartDate = ptrT(now.Add(24 * time.Hour))
		s.EnrollmentEndDate = ptrT(now.Add(48 * time.Hour))
		err := enroll(newSvc(s))
		if err == nil || !strings.Contains(err.Error(), "ainda não iniciaram") {
			t.Fatalf("esperava erro de inscrições não iniciadas, veio: %v", err)
		}
	})

	t.Run("turma sem janela definida não bloqueia (compat legado)", func(t *testing.T) {
		if err := enroll(newSvc(base)); err != nil {
			t.Fatalf("esperava sucesso para turma sem datas, veio erro: %v", err)
		}
	})
}
