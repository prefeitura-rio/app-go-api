package models_test

import (
	"testing"
	"time"

	"github.com/prefeitura-rio/app-go-api/internal/models"
)

// remoteVacancy builds a remote schedule with a class window and remaining vacancies.
func remoteSchedule(remaining int, start, end *time.Time) models.RemoteSchedule {
	s := models.RemoteSchedule{Vacancies: remaining, ClassStartDate: start, ClassEndDate: end}
	s.RemainingVacancies = remaining
	return s
}

// locationSchedule builds a location schedule with a class window and remaining vacancies.
func locationSchedule(remaining int, start, end time.Time) models.CourseSchedule {
	s := models.CourseSchedule{Vacancies: remaining, ClassStartDate: start, ClassEndDate: end}
	s.RemainingVacancies = remaining
	return s
}

func TestCurso_IsEnrollmentAvailable(t *testing.T) {
	now := time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)
	past := now.Add(-30 * 24 * time.Hour)
	future := now.Add(30 * 24 * time.Hour)
	classStart := now.Add(-10 * 24 * time.Hour)
	classEnd := now.Add(10 * 24 * time.Hour)
	classEndPast := now.Add(-5 * 24 * time.Hour)

	tests := []struct {
		name  string
		curso models.Curso
		want  bool
	}{
		{
			name: "presencial accepting_enrollments with vacancies is available",
			curso: models.Curso{
				Status:              models.StatusCursoPublished,
				Modalidade:          models.ModalidadePresencial,
				EnrollmentStartDate: ptr(past),
				EnrollmentEndDate:   ptr(future),
				LocationClasses: []models.LocationClass{
					{Schedules: []models.CourseSchedule{locationSchedule(5, classStart, classEnd)}},
				},
			},
			want: true,
		},
		{
			// Regression from the ticket (course id 2804): accepting_enrollments but
			// every turma is sold out must not count as available.
			name: "presencial accepting_enrollments but sold out is NOT available",
			curso: models.Curso{
				Status:              models.StatusCursoPublished,
				Modalidade:          models.ModalidadePresencial,
				EnrollmentStartDate: ptr(past),
				EnrollmentEndDate:   ptr(future),
				LocationClasses: []models.LocationClass{
					{Schedules: []models.CourseSchedule{locationSchedule(0, classStart, classEnd)}},
				},
			},
			want: false,
		},
		{
			name: "remoto accepting_enrollments with vacancies is available",
			curso: models.Curso{
				Status:              models.StatusCursoPublished,
				Modalidade:          models.ModalidadeRemoto,
				EnrollmentStartDate: ptr(past),
				EnrollmentEndDate:   ptr(future),
				RemoteClass: &models.RemoteClass{
					Schedules: []models.RemoteSchedule{remoteSchedule(3, ptr(classStart), ptr(classEnd))},
				},
			},
			want: true,
		},
		{
			name: "remoto accepting_enrollments but sold out is NOT available",
			curso: models.Curso{
				Status:              models.StatusCursoPublished,
				Modalidade:          models.ModalidadeRemoto,
				EnrollmentStartDate: ptr(past),
				EnrollmentEndDate:   ptr(future),
				RemoteClass: &models.RemoteClass{
					Schedules: []models.RemoteSchedule{remoteSchedule(0, ptr(classStart), ptr(classEnd))},
				},
			},
			want: false,
		},
		{
			name: "closed is NOT available",
			curso: models.Curso{
				Status:     models.StatusCursoClosed,
				Modalidade: models.ModalidadePresencial,
				LocationClasses: []models.LocationClass{
					{Schedules: []models.CourseSchedule{locationSchedule(5, classStart, classEnd)}},
				},
			},
			want: false,
		},
		{
			name: "canceled is NOT available",
			curso: models.Curso{
				Status:     models.StatusCursoCanceled,
				Modalidade: models.ModalidadePresencial,
				LocationClasses: []models.LocationClass{
					{Schedules: []models.CourseSchedule{locationSchedule(5, classStart, classEnd)}},
				},
			},
			want: false,
		},
		{
			name: "finished (past class dates) is NOT available",
			curso: models.Curso{
				Status:              models.StatusCursoPublished,
				Modalidade:          models.ModalidadePresencial,
				EnrollmentStartDate: ptr(past),
				EnrollmentEndDate:   ptr(past.Add(5 * 24 * time.Hour)),
				LocationClasses: []models.LocationClass{
					{Schedules: []models.CourseSchedule{locationSchedule(5, classStart, classEndPast)}},
				},
			},
			want: false,
		},
		{
			name: "scheduled (enrollment not open yet) is NOT available",
			curso: models.Curso{
				Status:              models.StatusCursoPublished,
				Modalidade:          models.ModalidadePresencial,
				EnrollmentStartDate: ptr(future),
				LocationClasses: []models.LocationClass{
					{Schedules: []models.CourseSchedule{locationSchedule(5, classStart, classEnd)}},
				},
			},
			want: false,
		},
		{
			name: "published, enrollment window passed, is NOT available",
			curso: models.Curso{
				Status:              models.StatusCursoPublished,
				Modalidade:          models.ModalidadePresencial,
				EnrollmentStartDate: ptr(past),
				EnrollmentEndDate:   ptr(past.Add(24 * time.Hour)),
				LocationClasses: []models.LocationClass{
					{Schedules: []models.CourseSchedule{locationSchedule(5, classStart, classEnd)}},
				},
			},
			want: false,
		},
		{
			name: "draft is NOT available",
			curso: models.Curso{
				Status:     models.StatusCursoDraft,
				Modalidade: models.ModalidadePresencial,
				LocationClasses: []models.LocationClass{
					{Schedules: []models.CourseSchedule{locationSchedule(5, classStart, classEnd)}},
				},
			},
			want: false,
		},
		{
			// LIVRE_FORMACAO_ONLINE has no turmas, so the vacancy gate does not apply.
			name: "LIVRE_FORMACAO_ONLINE within window is available (no vacancy gate)",
			curso: models.Curso{
				Status:              models.StatusCursoPublished,
				Modalidade:          models.ModalidadeLivreFormacaoOnline,
				EnrollmentStartDate: ptr(past),
				EnrollmentEndDate:   ptr(future),
			},
			want: true,
		},
		{
			name: "presencial with one sold-out and one open turma is available",
			curso: models.Curso{
				Status:              models.StatusCursoPublished,
				Modalidade:          models.ModalidadePresencial,
				EnrollmentStartDate: ptr(past),
				EnrollmentEndDate:   ptr(future),
				LocationClasses: []models.LocationClass{
					{Schedules: []models.CourseSchedule{
						locationSchedule(0, classStart, classEnd),
						locationSchedule(2, classStart, classEnd),
					}},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.curso
			if got := c.IsEnrollmentAvailable(now); got != tt.want {
				t.Errorf("IsEnrollmentAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}
