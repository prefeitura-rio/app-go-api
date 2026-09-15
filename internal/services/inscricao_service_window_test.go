package services_test

import (
	"context"
	"errors"
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

// TestInscricaoService_Create_EnrollmentRuleErrorType pins the error type the
// handler relies on to answer 400 instead of 500: a citizen hitting a closed
// turma is not a server fault, and the portal shows the message verbatim.
func TestInscricaoService_Create_EnrollmentRuleErrorType(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	scheduleID := uuid.New()
	ptrT := func(v time.Time) *time.Time { return &v }
	accepting := true

	newSvc := func(courseStatus string, courseStart, courseEnd *time.Time, schedule models.CourseSchedule) *services.InscricaoService {
		inscricaoRepo := &MockInscricaoRepository{
			CreateFunc:              func(ctx context.Context, i *models.Inscricao) error { return nil },
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) { return false, nil },
		}
		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (string, *time.Time, *time.Time, bool, error) {
				return courseStatus, courseStart, courseEnd, false, nil
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

	enroll := func(svc *services.InscricaoService) error {
		return svc.Create(ctx, &models.Inscricao{
			CPF: "12345678900", CursoID: 1, Name: "Teste", Email: "t@e.com", ScheduleID: &scheduleID,
		})
	}

	closedTurma := models.CourseSchedule{
		ID: scheduleID, Vacancies: 10, AcceptingEnrollments: &accepting,
		EnrollmentStartDate: ptrT(now.Add(-48 * time.Hour)),
		EnrollmentEndDate:   ptrT(now.Add(-24 * time.Hour)),
	}
	openTurma := models.CourseSchedule{
		ID: scheduleID, Vacancies: 10, AcceptingEnrollments: &accepting,
		EnrollmentStartDate: ptrT(now.Add(-24 * time.Hour)),
		EnrollmentEndDate:   ptrT(now.Add(24 * time.Hour)),
	}

	cases := []struct {
		name        string
		svc         *services.InscricaoService
		wantMessage string
	}{
		{
			name:        "turma encerrada",
			svc:         newSvc(string(models.StatusCursoOpened), nil, nil, closedTurma),
			wantMessage: "as inscrições desta turma já encerraram",
		},
		{
			name:        "curso fora do período",
			svc:         newSvc(string(models.StatusCursoOpened), ptrT(now.Add(24*time.Hour)), ptrT(now.Add(48*time.Hour)), openTurma),
			wantMessage: "período de inscrições ainda não iniciou",
		},
		{
			name:        "curso não aberto",
			svc:         newSvc(string(models.StatusCursoDraft), nil, nil, openTurma),
			wantMessage: "curso não está aberto para inscrições",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := enroll(tc.svc)
			if err == nil {
				t.Fatal("esperava erro de regra de negócio, veio nil")
			}

			var ruleErr *services.EnrollmentRuleError
			if !errors.As(err, &ruleErr) {
				t.Fatalf("esperava *services.EnrollmentRuleError (handler responde 400), veio %T: %v", err, err)
			}
			if err.Error() != tc.wantMessage {
				t.Fatalf("mensagem deve chegar íntegra ao cidadão: esperava %q, veio %q", tc.wantMessage, err.Error())
			}
		})
	}
}
