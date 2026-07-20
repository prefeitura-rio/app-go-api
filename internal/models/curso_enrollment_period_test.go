package models_test

import (
	"testing"
	"time"

	"github.com/prefeitura-rio/app-go-api/internal/models"
)

func TestCurso_ApplyDerivedEnrollmentPeriod(t *testing.T) {
	base := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	d := func(days int) time.Time { return base.Add(time.Duration(days) * 24 * time.Hour) }

	t.Run("presencial derives earliest start and latest end across turmas", func(t *testing.T) {
		c := models.Curso{
			Modalidade: models.ModalidadePresencial,
			LocationClasses: []models.LocationClass{
				{Schedules: []models.CourseSchedule{
					{EnrollmentStartDate: ptr(d(5)), EnrollmentEndDate: ptr(d(10))},
					{EnrollmentStartDate: ptr(d(2)), EnrollmentEndDate: ptr(d(8))},
				}},
				{Schedules: []models.CourseSchedule{
					{EnrollmentStartDate: ptr(d(3)), EnrollmentEndDate: ptr(d(20))},
				}},
			},
		}
		c.ApplyDerivedEnrollmentPeriod()
		if c.EnrollmentStartDate == nil || !c.EnrollmentStartDate.Equal(d(2)) {
			t.Errorf("start = %v, want %v", c.EnrollmentStartDate, d(2))
		}
		if c.EnrollmentEndDate == nil || !c.EnrollmentEndDate.Equal(d(20)) {
			t.Errorf("end = %v, want %v", c.EnrollmentEndDate, d(20))
		}
	})

	t.Run("online derives from remote schedules", func(t *testing.T) {
		c := models.Curso{
			Modalidade: models.ModalidadeRemoto,
			RemoteClass: &models.RemoteClass{Schedules: []models.RemoteSchedule{
				{EnrollmentStartDate: ptr(d(4)), EnrollmentEndDate: ptr(d(9))},
				{EnrollmentStartDate: ptr(d(1)), EnrollmentEndDate: ptr(d(6))},
			}},
		}
		c.ApplyDerivedEnrollmentPeriod()
		if c.EnrollmentStartDate == nil || !c.EnrollmentStartDate.Equal(d(1)) {
			t.Errorf("start = %v, want %v", c.EnrollmentStartDate, d(1))
		}
		if c.EnrollmentEndDate == nil || !c.EnrollmentEndDate.Equal(d(9)) {
			t.Errorf("end = %v, want %v", c.EnrollmentEndDate, d(9))
		}
	})

	t.Run("livre formacao online keeps its own course dates", func(t *testing.T) {
		c := models.Curso{
			Modalidade:          models.ModalidadeLivreFormacaoOnline,
			EnrollmentStartDate: ptr(d(0)),
			EnrollmentEndDate:   ptr(d(30)),
		}
		c.ApplyDerivedEnrollmentPeriod()
		if !c.EnrollmentStartDate.Equal(d(0)) || !c.EnrollmentEndDate.Equal(d(30)) {
			t.Errorf("livre formacao dates changed: %v..%v", c.EnrollmentStartDate, c.EnrollmentEndDate)
		}
	})

	t.Run("no turma enrollment dates is a no-op", func(t *testing.T) {
		c := models.Curso{
			Modalidade:          models.ModalidadePresencial,
			EnrollmentStartDate: ptr(d(0)),
			EnrollmentEndDate:   ptr(d(30)),
			LocationClasses: []models.LocationClass{
				{Schedules: []models.CourseSchedule{{}}},
			},
		}
		c.ApplyDerivedEnrollmentPeriod()
		if !c.EnrollmentStartDate.Equal(d(0)) || !c.EnrollmentEndDate.Equal(d(30)) {
			t.Errorf("dates changed on no-op: %v..%v", c.EnrollmentStartDate, c.EnrollmentEndDate)
		}
	})
}
