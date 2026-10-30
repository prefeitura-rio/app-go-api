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

// TestInscricaoService_ChangeSchedule_SameSchedule_RejectedAsAlreadyEnrolled
// covers DEV-141: asking to move to the turma the enrollment is already in is
// refused with its own message ("você já está nesta turma"), never with "não há
// vagas" (the vacancy count includes the enrollment itself, so a full turma used
// to answer "sem vagas"). The enrollment is not changed and no vacancy is
// counted. All data is synthetic.
func TestInscricaoService_ChangeSchedule_SameSchedule_RejectedAsAlreadyEnrolled(t *testing.T) {
	ctx := context.Background()
	accepting := true
	created := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	start := time.Now().Add(30 * 24 * time.Hour).UTC().Truncate(time.Second)

	ownerCPF := "12345678900"
	inscricaoID := uuid.New()
	locationID := uuid.New()
	currentSchedule := uuid.New()
	otherSchedule := uuid.New()

	curso := &models.Curso{
		ID: 1,
		LocationClasses: []models.LocationClass{{
			ID: locationID, CursoID: 1, Address: "Rua Sintética, 100",
			CreatedAt: created, UpdatedAt: created,
			Schedules: []models.CourseSchedule{
				{
					ID: currentSchedule, LocationID: locationID, Vacancies: 1, AcceptingEnrollments: &accepting,
					ClassStartDate: start, ClassEndDate: start.Add(20 * 24 * time.Hour),
					CreatedAt: created, UpdatedAt: created,
				},
				{
					ID: otherSchedule, LocationID: locationID, Vacancies: 10, AcceptingEnrollments: &accepting,
					ClassStartDate: start, ClassEndDate: start.Add(20 * 24 * time.Hour),
					CreatedAt: created, UpdatedAt: created,
				},
			},
		}},
	}

	tests := []struct {
		name   string
		status models.StatusInscricao
	}{
		{"approved enrollment in a full turma", models.StatusInscricaoApproved},
		{"pending enrollment in a full turma", models.StatusInscricaoPending},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheduleID := currentSchedule
			updated := false
			counted := false

			inscricaoRepo := &MockInscricaoRepository{
				GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
					return &models.Inscricao{
						ID: inscricaoID, CPF: ownerCPF, CursoID: 1, Status: tt.status, ScheduleID: &scheduleID,
					}, nil
				},
				UpdateFunc: func(ctx context.Context, i *models.Inscricao) error {
					updated = true
					return nil
				},
			}
			cursoRepo := &MockCursoRepository{
				GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) { return curso, nil },
				CountEnrollmentsByScheduleIDsFunc: func(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]int64, error) {
					counted = true
					// The turma is full, and the enrollment itself is one of the counted.
					return map[uuid.UUID]int64{currentSchedule: 1}, nil
				},
				GetCourseScheduleByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.CourseSchedule, error) {
					s := curso.LocationClasses[0].Schedules[0]
					return &s, nil
				},
			}
			svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

			same := currentSchedule
			got, err := svc.ChangeSchedule(ctx, inscricaoID, ownerCPF, &models.ScheduleChangeRequest{
				ScheduleID: &same,
				EnrolledUnit: &models.EnrolledUnit{
					ID: locationID.String(), CursoID: 1,
					Schedules: []models.EnrolledUnitSchedule{
						{ID: currentSchedule.String(), ClassStartDate: start.Format(time.RFC3339)},
					},
				},
			})

			if err == nil {
				t.Fatalf("esperava recusa da troca para a turma atual, veio sucesso: %+v", got)
			}
			if strings.Contains(err.Error(), "não há vagas") {
				t.Fatalf("a troca para a turma atual não pode responder \"sem vagas\": %v", err)
			}
			if !strings.Contains(err.Error(), "você já está nesta turma") {
				t.Fatalf("esperava a mensagem \"você já está nesta turma\", veio: %v", err)
			}
			// The handler answers 400 for this message (rule error, not a server failure).
			if !strings.Contains(err.Error(), "não é possível trocar de turma") {
				t.Fatalf("a mensagem deve seguir o padrão das recusas de troca (400), veio: %v", err)
			}
			var ruleErr *services.EnrollmentRuleError
			if !errors.As(err, &ruleErr) {
				t.Fatalf("esperava EnrollmentRuleError, veio %T: %v", err, err)
			}
			if updated {
				t.Fatal("a inscrição não deveria ser alterada")
			}
			if counted {
				t.Fatal("as vagas não deveriam ser contadas para a turma atual")
			}
		})
	}
}
