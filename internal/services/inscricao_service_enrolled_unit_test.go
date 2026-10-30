package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/prefeitura-rio/app-go-api/internal/config"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

// TestInscricaoService_Create_EnrolledUnitFromValidatedSchedule covers DEV-134:
// on the citizen Create flow the enrolled_unit sent in the body is never
// persisted as-is. It is rebuilt from the course data for the schedule_id that
// passed validateScheduleID, and dropped when there is no validated schedule.
func TestInscricaoService_Create_EnrolledUnitFromValidatedSchedule(t *testing.T) {
	ctx := context.Background()
	accepting := true
	created := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	updated := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)

	locationID := uuid.New()
	scheduleID := uuid.New()
	otherScheduleID := uuid.New()
	remoteScheduleID := uuid.New()

	location := models.LocationClass{
		ID:               locationID,
		CursoID:          1,
		Address:          "Rua Sintética, 100",
		Neighborhood:     "Centro",
		NeighborhoodZone: "Centro",
		CreatedAt:        created,
		UpdatedAt:        updated,
		Schedules: []models.CourseSchedule{
			{
				ID: scheduleID, LocationID: locationID, Vacancies: 20, AcceptingEnrollments: &accepting,
				ClassStartDate: time.Date(2026, 12, 1, 9, 0, 0, 0, time.UTC),
				ClassEndDate:   time.Date(2026, 12, 20, 9, 0, 0, 0, time.UTC),
				ClassTime:      "09:00 - 12:00", ClassDays: "Seg, Qua",
				CreatedAt: created, UpdatedAt: updated,
			},
			{
				ID: otherScheduleID, LocationID: locationID, Vacancies: 15, AcceptingEnrollments: &accepting,
				ClassStartDate: time.Date(2027, 1, 10, 14, 0, 0, 0, time.UTC),
				ClassEndDate:   time.Date(2027, 1, 30, 14, 0, 0, 0, time.UTC),
				ClassTime:      "14:00 - 17:00", ClassDays: "Ter, Qui",
				CreatedAt: created, UpdatedAt: updated,
			},
		},
	}
	remote := &models.RemoteClass{
		ID:      uuid.New(),
		CursoID: 1,
		Schedules: []models.RemoteSchedule{
			{ID: remoteScheduleID, Vacancies: 50, AcceptingEnrollments: &accepting},
		},
	}

	forged := func() *models.EnrolledUnit {
		return &models.EnrolledUnit{
			ID:      "qualquer",
			CursoID: 999,
			Address: "Endereço inventado",
			Schedules: []models.EnrolledUnitSchedule{
				{ID: uuid.NewString(), Vacancies: 9999, ClassStartDate: "2030-01-01"},
			},
		}
	}

	newSvc := func(persisted **models.Inscricao) *services.InscricaoService {
		inscricaoRepo := &MockInscricaoRepository{
			CreateFunc: func(ctx context.Context, i *models.Inscricao) error {
				*persisted = i
				return nil
			},
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) { return false, nil },
		}
		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (string, *time.Time, *time.Time, bool, error) {
				return string(models.StatusCursoOpened), nil, nil, false, nil
			},
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return &models.Curso{ID: 1, LocationClasses: []models.LocationClass{location}, RemoteClass: remote}, nil
			},
		}
		return services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})
	}

	t.Run("enrolled_unit without schedule_id is dropped", func(t *testing.T) {
		var persisted *models.Inscricao
		err := newSvc(&persisted).Create(ctx, &models.Inscricao{
			CPF: "12345678900", CursoID: 1, Name: "Teste", EnrolledUnit: forged(),
		})
		if err != nil {
			t.Fatalf("esperava sucesso, veio erro: %v", err)
		}
		if persisted == nil {
			t.Fatal("inscrição não foi persistida")
		}
		if persisted.EnrolledUnit != nil {
			t.Fatalf("enrolled_unit do corpo não deveria ser persistido sem schedule_id, veio: %+v", persisted.EnrolledUnit)
		}
	})

	t.Run("enrolled_unit is rebuilt from the validated location schedule", func(t *testing.T) {
		var persisted *models.Inscricao
		err := newSvc(&persisted).Create(ctx, &models.Inscricao{
			CPF: "12345678900", CursoID: 1, Name: "Teste", ScheduleID: &scheduleID, EnrolledUnit: forged(),
		})
		if err != nil {
			t.Fatalf("esperava sucesso, veio erro: %v", err)
		}
		unit := persisted.EnrolledUnit
		if unit == nil {
			t.Fatal("esperava enrolled_unit montado a partir da turma validada")
		}
		if unit.ID != locationID.String() || unit.CursoID != 1 || unit.Address != location.Address ||
			unit.Neighborhood != location.Neighborhood || unit.NeighborhoodZone != location.NeighborhoodZone {
			t.Fatalf("enrolled_unit não corresponde ao local do curso: %+v", unit)
		}
		if unit.CreatedAt != created.Format(time.RFC3339) || unit.UpdatedAt != updated.Format(time.RFC3339) {
			t.Fatalf("datas do local inesperadas: %q %q", unit.CreatedAt, unit.UpdatedAt)
		}
		if len(unit.Schedules) != 2 {
			t.Fatalf("esperava as 2 turmas do local, veio %d", len(unit.Schedules))
		}
		got := unit.Schedules[0]
		want := location.Schedules[0]
		if got.ID != scheduleID.String() || got.LocationID != locationID.String() || got.Vacancies != want.Vacancies ||
			got.ClassStartDate != want.ClassStartDate.Format(time.RFC3339) ||
			got.ClassEndDate != want.ClassEndDate.Format(time.RFC3339) ||
			got.ClassTime != want.ClassTime || got.ClassDays != want.ClassDays {
			t.Fatalf("turma não corresponde ao curso: %+v", got)
		}
		if unit.Schedules[1].ID != otherScheduleID.String() {
			t.Fatalf("segunda turma inesperada: %+v", unit.Schedules[1])
		}
	})

	t.Run("remote schedule keeps no enrolled_unit", func(t *testing.T) {
		var persisted *models.Inscricao
		err := newSvc(&persisted).Create(ctx, &models.Inscricao{
			CPF: "12345678900", CursoID: 1, Name: "Teste", ScheduleID: &remoteScheduleID, EnrolledUnit: forged(),
		})
		if err != nil {
			t.Fatalf("esperava sucesso, veio erro: %v", err)
		}
		if persisted.EnrolledUnit != nil {
			t.Fatalf("turma remota não tem local; enrolled_unit do corpo não deveria ser persistido: %+v", persisted.EnrolledUnit)
		}
	})

	t.Run("schedule from another course is still rejected", func(t *testing.T) {
		var persisted *models.Inscricao
		foreign := uuid.New()
		err := newSvc(&persisted).Create(ctx, &models.Inscricao{
			CPF: "12345678900", CursoID: 1, Name: "Teste", ScheduleID: &foreign, EnrolledUnit: forged(),
		})
		if err == nil {
			t.Fatal("esperava erro para turma que não pertence ao curso")
		}
		if persisted != nil {
			t.Fatal("inscrição não deveria ter sido persistida")
		}
	})
}
