package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

// TestCursoEnrollmentPeriod_Integration exercises the full service+repository
// path against a real Postgres: two turmas with different enrollment windows
// must persist per-turma, and the course-level period must be derived
// (earliest opening / latest closing). Gated by RUN_REPOSITORY_INTEGRATION.
func TestCursoEnrollmentPeriod_Integration(t *testing.T) {
	db := getDBForIntegration(t)
	if db == nil {
		return
	}
	ctx := context.Background()
	repo := repository.NewCursoRepository(db)
	svc := services.NewCursoService(repo)

	d := func(m, day int) time.Time { return time.Date(2026, time.Month(m), day, 9, 0, 0, 0, time.UTC) }
	p := func(v time.Time) *time.Time { return &v }

	// Turma 1: inscrições 01/03–10/03 ; Turma 2: inscrições 05/03–20/03
	// Curso derivado esperado: início 01/03 (menor) / fim 20/03 (maior).
	curso := &models.Curso{
		Titulo:     "[INTEGRATION] enrollment per turma",
		Modalidade: models.ModalidadePresencial,
		Status:     models.StatusCursoDraft,
		LocationClasses: []models.LocationClass{{
			Address:      "Rua de Teste, 123 - Centro",
			Neighborhood: "Centro",
			Schedules: []models.CourseSchedule{
				{
					Vacancies:           30,
					EnrollmentStartDate: p(d(3, 1)), EnrollmentEndDate: p(d(3, 10)),
					ClassStartDate: d(3, 15), ClassEndDate: d(4, 15),
					ClassTime: "09h-12h", ClassDays: "Seg,Ter",
				},
				{
					Vacancies:           30,
					EnrollmentStartDate: p(d(3, 5)), EnrollmentEndDate: p(d(3, 20)),
					ClassStartDate: d(3, 25), ClassEndDate: d(4, 25),
					ClassTime: "14h-17h", ClassDays: "Qua,Qui",
				},
			},
		}},
	}

	id, err := svc.Create(ctx, curso)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() {
		db.Where("curso_id = ?", id).Delete(&models.LocationClass{})
		db.Delete(&models.Curso{}, id)
	})

	saved, err := repo.GetByID(ctx, id)
	if err != nil || saved == nil {
		t.Fatalf("GetByID: err=%v saved=%v", err, saved)
	}
	if len(saved.LocationClasses) != 1 || len(saved.LocationClasses[0].Schedules) != 2 {
		t.Fatalf("expected 1 location with 2 schedules, got %+v", saved.LocationClasses)
	}

	// Cada turma deve ter persistido a própria janela de inscrição.
	for i, sc := range saved.LocationClasses[0].Schedules {
		if sc.EnrollmentStartDate == nil || sc.EnrollmentEndDate == nil {
			t.Fatalf("turma %d: janela de inscrição não persistida (start=%v end=%v)",
				i+1, sc.EnrollmentStartDate, sc.EnrollmentEndDate)
		}
	}

	// Período do curso derivado das turmas (menor abertura / maior encerramento).
	if saved.EnrollmentStartDate == nil || !saved.EnrollmentStartDate.Equal(d(3, 1)) {
		t.Errorf("curso enrollment_start: got %v, want %v", saved.EnrollmentStartDate, d(3, 1))
	}
	if saved.EnrollmentEndDate == nil || !saved.EnrollmentEndDate.Equal(d(3, 20)) {
		t.Errorf("curso enrollment_end: got %v, want %v", saved.EnrollmentEndDate, d(3, 20))
	}
}
