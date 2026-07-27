package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

// mustParse is a tiny helper for building expected/input instants in tests.
func mustParse(t *testing.T, s string) time.Time {
	t.Helper()
	v, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatalf("parse %q: %v", s, err)
	}
	return v
}

// TestNormalizeCurso_MidnightUTCDates verifies that day-granular date fields stored
// at midnight UTC are re-anchored to the intended BRT day (start-of-day for openings,
// end-of-day for closings) so they display correctly under -03:00 serialization,
// while values carrying a real time-of-day are left untouched.
func TestNormalizeCurso_MidnightUTCDates(t *testing.T) {
	ctx := context.Background()

	newSvc := func(capture **models.Curso) *services.CursoService {
		repo := &MockCursoRepository{
			CreateFunc: func(_ context.Context, curso *models.Curso) (int, error) {
				*capture = curso
				return 1, nil
			},
			CreateLocationClassesFunc: func(_ context.Context, _ []models.LocationClass) error {
				return nil
			},
		}
		return services.NewCursoServiceWithInterface(repo)
	}

	utc := func(tm time.Time) string { return tm.UTC().Format(time.RFC3339) }

	t.Run("midnight-UTC dates re-anchored to BRT day", func(t *testing.T) {
		var created *models.Curso
		svc := newSvc(&created)

		curso := &models.Curso{
			Titulo:     "Curso TZ",
			Status:     models.StatusCursoOpened,
			Modalidade: models.ModalidadePresencial,
			LocationClasses: []models.LocationClass{
				{
					Address:      "Rua de Teste, 123",
					Neighborhood: "Centro",
					Schedules: []models.CourseSchedule{
						{
							Vacancies:           30,
							EnrollmentStartDate: ptr(mustParse(t, "2027-03-01T00:00:00Z")),
							EnrollmentEndDate:   ptr(mustParse(t, "2027-03-08T00:00:00Z")),
							ClassStartDate:      mustParse(t, "2027-03-10T00:00:00Z"),
							ClassEndDate:        mustParse(t, "2027-03-20T00:00:00Z"),
							ClassTime:           "09:00-12:00",
							ClassDays:           "Segunda a Sexta",
						},
					},
				},
			},
		}

		if _, err := svc.Create(ctx, curso); err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		sched := created.LocationClasses[0].Schedules[0]
		// start-of-day BRT == 03:00Z ; end-of-day BRT (23:59:59-03) == 02:59:59Z do dia seguinte
		if got := utc(sched.ClassStartDate); got != "2027-03-10T03:00:00Z" {
			t.Errorf("ClassStartDate: got %s, want 2027-03-10T03:00:00Z", got)
		}
		if got := utc(sched.ClassEndDate); got != "2027-03-21T02:59:59Z" {
			t.Errorf("ClassEndDate: got %s, want 2027-03-21T02:59:59Z", got)
		}
		if got := utc(*sched.EnrollmentStartDate); got != "2027-03-01T03:00:00Z" {
			t.Errorf("EnrollmentStartDate: got %s, want 2027-03-01T03:00:00Z", got)
		}
		if got := utc(*sched.EnrollmentEndDate); got != "2027-03-09T02:59:59Z" {
			t.Errorf("EnrollmentEndDate: got %s, want 2027-03-09T02:59:59Z", got)
		}

		// Course-level enrollment window is derived from the (already normalized) turmas.
		if got := utc(*created.EnrollmentStartDate); got != "2027-03-01T03:00:00Z" {
			t.Errorf("course EnrollmentStartDate: got %s, want 2027-03-01T03:00:00Z", got)
		}
		if got := utc(*created.EnrollmentEndDate); got != "2027-03-09T02:59:59Z" {
			t.Errorf("course EnrollmentEndDate: got %s, want 2027-03-09T02:59:59Z", got)
		}
	})

	t.Run("dates with real time-of-day are left untouched", func(t *testing.T) {
		var created *models.Curso
		svc := newSvc(&created)

		// 13:00Z = 10:00 BRT, 22:00Z = 19:00 BRT — same calendar day in BRT, no shift needed.
		start := mustParse(t, "2027-03-10T13:00:00Z")
		end := mustParse(t, "2027-03-20T22:00:00Z")

		curso := &models.Curso{
			Titulo:     "Curso Hora Real",
			Status:     models.StatusCursoOpened,
			Modalidade: models.ModalidadePresencial,
			LocationClasses: []models.LocationClass{
				{
					Address:      "Rua de Teste, 123",
					Neighborhood: "Centro",
					Schedules: []models.CourseSchedule{
						{
							Vacancies:      30,
							ClassStartDate: start,
							ClassEndDate:   end,
							ClassTime:      "10:00-13:00",
							ClassDays:      "Segunda a Sexta",
						},
					},
				},
			},
		}

		if _, err := svc.Create(ctx, curso); err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		sched := created.LocationClasses[0].Schedules[0]
		if !sched.ClassStartDate.Equal(start) {
			t.Errorf("ClassStartDate mudou: got %s, want %s", utc(sched.ClassStartDate), utc(start))
		}
		if !sched.ClassEndDate.Equal(end) {
			t.Errorf("ClassEndDate mudou: got %s, want %s", utc(sched.ClassEndDate), utc(end))
		}
	})
}

func ptr(t time.Time) *time.Time { return &t }
