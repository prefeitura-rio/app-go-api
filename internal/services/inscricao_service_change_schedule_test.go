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

// changeScheduleFixture is a course with one location holding two schedules
// and a remote class holding one, plus an approved enrollment of the owner in
// the first schedule. All data is synthetic.
type changeScheduleFixture struct {
	ownerCPF         string
	inscricaoID      uuid.UUID
	locationID       uuid.UUID
	currentSchedule  uuid.UUID
	targetSchedule   uuid.UUID
	remoteClassID    uuid.UUID
	remoteScheduleID uuid.UUID
	curso            *models.Curso
}

func newChangeScheduleFixture(targetStart time.Time) *changeScheduleFixture {
	accepting := true
	created := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	updated := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	f := &changeScheduleFixture{
		ownerCPF:         "12345678900",
		inscricaoID:      uuid.New(),
		locationID:       uuid.New(),
		currentSchedule:  uuid.New(),
		targetSchedule:   uuid.New(),
		remoteClassID:    uuid.New(),
		remoteScheduleID: uuid.New(),
	}
	remoteStart := time.Now().Add(30 * 24 * time.Hour).UTC().Truncate(time.Second)
	remoteTime := "19:00 - 21:00"
	f.curso = &models.Curso{
		ID: 1,
		LocationClasses: []models.LocationClass{
			{
				ID:               f.locationID,
				CursoID:          1,
				Address:          "Rua Sintética, 100",
				Neighborhood:     "Centro",
				NeighborhoodZone: "Centro",
				CreatedAt:        created,
				UpdatedAt:        updated,
				Schedules: []models.CourseSchedule{
					{
						ID: f.currentSchedule, LocationID: f.locationID, Vacancies: 20, AcceptingEnrollments: &accepting,
						ClassStartDate: time.Now().Add(10 * 24 * time.Hour).UTC().Truncate(time.Second),
						ClassEndDate:   time.Now().Add(20 * 24 * time.Hour).UTC().Truncate(time.Second),
						ClassTime:      "09:00 - 12:00", ClassDays: "Seg, Qua",
						CreatedAt: created, UpdatedAt: updated,
					},
					{
						ID: f.targetSchedule, LocationID: f.locationID, Vacancies: 10, AcceptingEnrollments: &accepting,
						ClassStartDate: targetStart.UTC().Truncate(time.Second),
						ClassEndDate:   targetStart.Add(20 * 24 * time.Hour).UTC().Truncate(time.Second),
						ClassTime:      "14:00 - 17:00", ClassDays: "Ter, Qui",
						CreatedAt: created, UpdatedAt: updated,
					},
				},
			},
		},
		RemoteClass: &models.RemoteClass{
			ID:        f.remoteClassID,
			CursoID:   1,
			CreatedAt: created,
			UpdatedAt: updated,
			Schedules: []models.RemoteSchedule{
				{
					ID: f.remoteScheduleID, RemoteClassID: f.remoteClassID, Vacancies: 50, AcceptingEnrollments: &accepting,
					ClassStartDate: &remoteStart, ClassTime: &remoteTime,
					CreatedAt: created, UpdatedAt: updated,
				},
			},
		},
	}
	return f
}

// forgedEnrolledUnit is what a citizen could hand-craft in the request body.
func forgedEnrolledUnit() *models.EnrolledUnit {
	return &models.EnrolledUnit{
		ID:      "unidade-inventada",
		CursoID: 999,
		Address: "Endereço inventado",
		Schedules: []models.EnrolledUnitSchedule{
			{ID: uuid.NewString(), Vacancies: 9999, ClassStartDate: time.Now().Add(365 * 24 * time.Hour).Format(time.RFC3339)},
		},
	}
}

type changeScheduleDeps struct {
	persisted   *models.Inscricao
	updateCalls int
	counts      map[uuid.UUID]int64
	countErr    error
	status      models.StatusInscricao
	getErr      error
	notFound    bool
	updateErr   error
	nilConfig   bool
}

func (f *changeScheduleFixture) service(d *changeScheduleDeps) *services.InscricaoService {
	inscricaoRepo := &MockInscricaoRepository{
		GetByIDFunc: func(_ context.Context, id uuid.UUID) (*models.Inscricao, error) {
			if d.getErr != nil {
				return nil, d.getErr
			}
			if d.notFound {
				return nil, nil
			}
			status := d.status
			if status == "" {
				status = models.StatusInscricaoApproved
			}
			current := f.currentSchedule
			return &models.Inscricao{
				ID:         f.inscricaoID,
				CPF:        f.ownerCPF,
				CursoID:    f.curso.ID,
				Status:     status,
				ScheduleID: &current,
			}, nil
		},
		UpdateFunc: func(_ context.Context, inscricao *models.Inscricao) error {
			d.updateCalls++
			if d.updateErr != nil {
				return d.updateErr
			}
			d.persisted = inscricao
			return nil
		},
	}
	cursoRepo := &MockCursoRepository{
		GetByIDFunc: func(_ context.Context, id int) (*models.Curso, error) {
			if id != f.curso.ID {
				return nil, nil
			}
			return f.curso, nil
		},
		CountEnrollmentsByScheduleIDsFunc: func(_ context.Context, ids []uuid.UUID) (map[uuid.UUID]int64, error) {
			if d.countErr != nil {
				return nil, d.countErr
			}
			return d.counts, nil
		},
		// The schedule lookups by id alone do not tell which course a schedule
		// belongs to; ChangeSchedule must resolve it from the enrollment's course.
		GetCourseScheduleByIDFunc: func(_ context.Context, id uuid.UUID) (*models.CourseSchedule, error) {
			accepting := true
			return &models.CourseSchedule{ID: id, Vacancies: 100, AcceptingEnrollments: &accepting}, nil
		},
		GetRemoteScheduleByIDFunc: func(_ context.Context, id uuid.UUID) (*models.RemoteSchedule, error) {
			accepting := true
			return &models.RemoteSchedule{ID: id, Vacancies: 100, AcceptingEnrollments: &accepting}, nil
		},
	}
	cfg := &config.AppConfig{Enrollment: config.EnrollmentSettings{ScheduleChangeDeadlineHours: 48}}
	if d.nilConfig {
		cfg = nil
	}
	return services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, cfg)
}

func assertRuleError(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error containing %q, got nil", want)
	}
	var ruleErr *services.EnrollmentRuleError
	if !errors.As(err, &ruleErr) {
		t.Errorf("expected *services.EnrollmentRuleError, got %T (%v)", err, err)
	}
	if !strings.Contains(err.Error(), want) {
		t.Errorf("expected error containing %q, got %q", want, err.Error())
	}
}

// TestInscricaoService_ChangeSchedule_EnrolledUnitFromCourse covers DEV-136:
// the class change takes the turma only from a schedule_id of the enrollment's
// own course, validated against the course data, and rebuilds the
// enrolled_unit on the server instead of copying the request body.
func TestInscricaoService_ChangeSchedule_EnrolledUnitFromCourse(t *testing.T) {
	ctx := context.Background()

	t.Run("forged_enrolled_unit_is_replaced_by_the_course_unit", func(t *testing.T) {
		f := newChangeScheduleFixture(time.Now().Add(15 * 24 * time.Hour))
		d := &changeScheduleDeps{counts: map[uuid.UUID]int64{f.targetSchedule: 3}}

		result, err := f.service(d).ChangeSchedule(ctx, f.inscricaoID, f.ownerCPF, &models.ScheduleChangeRequest{
			ScheduleID:   &f.targetSchedule,
			EnrolledUnit: forgedEnrolledUnit(),
		})
		if err != nil {
			t.Fatalf("ChangeSchedule failed: %v", err)
		}
		if d.updateCalls != 1 || d.persisted == nil {
			t.Fatalf("expected one Update, got %d", d.updateCalls)
		}
		if result != d.persisted {
			t.Errorf("expected the persisted enrollment to be returned")
		}
		if d.persisted.ScheduleID == nil || *d.persisted.ScheduleID != f.targetSchedule {
			t.Fatalf("schedule_id = %v, want %s", d.persisted.ScheduleID, f.targetSchedule)
		}
		unit := d.persisted.EnrolledUnit
		if unit == nil {
			t.Fatal("expected enrolled_unit built from the course, got nil")
		}
		if unit.ID != f.locationID.String() || unit.CursoID != f.curso.ID || unit.Address != "Rua Sintética, 100" {
			t.Errorf("enrolled_unit = {ID:%s CursoID:%d Address:%s}, want the course location", unit.ID, unit.CursoID, unit.Address)
		}
		if len(unit.Schedules) != 2 {
			t.Fatalf("expected the 2 schedules of the location, got %d", len(unit.Schedules))
		}
		var found bool
		for _, s := range unit.Schedules {
			if s.Vacancies == 9999 {
				t.Errorf("forged schedule leaked into enrolled_unit: %+v", s)
			}
			if s.ID == f.targetSchedule.String() {
				found = true
				if s.Vacancies != 10 || s.ClassTime != "14:00 - 17:00" {
					t.Errorf("target schedule = %+v, want the course data", s)
				}
			}
		}
		if !found {
			t.Errorf("target schedule %s missing from enrolled_unit", f.targetSchedule)
		}
	})

	t.Run("enrolled_unit_is_not_required_in_the_body", func(t *testing.T) {
		f := newChangeScheduleFixture(time.Now().Add(15 * 24 * time.Hour))
		d := &changeScheduleDeps{counts: map[uuid.UUID]int64{}}

		_, err := f.service(d).ChangeSchedule(ctx, f.inscricaoID, f.ownerCPF, &models.ScheduleChangeRequest{
			ScheduleID: &f.targetSchedule,
		})
		if err != nil {
			t.Fatalf("ChangeSchedule failed without enrolled_unit: %v", err)
		}
		if d.persisted == nil || d.persisted.EnrolledUnit == nil || d.persisted.EnrolledUnit.ID != f.locationID.String() {
			t.Fatalf("expected enrolled_unit of location %s, got %+v", f.locationID, d.persisted)
		}
	})

	t.Run("remote_schedule_builds_the_remote_unit", func(t *testing.T) {
		f := newChangeScheduleFixture(time.Now().Add(15 * 24 * time.Hour))
		d := &changeScheduleDeps{counts: map[uuid.UUID]int64{f.remoteScheduleID: 1}}

		_, err := f.service(d).ChangeSchedule(ctx, f.inscricaoID, f.ownerCPF, &models.ScheduleChangeRequest{
			ScheduleID:   &f.remoteScheduleID,
			EnrolledUnit: forgedEnrolledUnit(),
		})
		if err != nil {
			t.Fatalf("ChangeSchedule to remote schedule failed: %v", err)
		}
		unit := d.persisted.EnrolledUnit
		if unit == nil || unit.ID != f.remoteClassID.String() || unit.CursoID != f.curso.ID {
			t.Fatalf("expected the remote class unit %s, got %+v", f.remoteClassID, unit)
		}
		if len(unit.Schedules) != 1 || unit.Schedules[0].ID != f.remoteScheduleID.String() || unit.Schedules[0].Vacancies != 50 {
			t.Errorf("remote unit schedules = %+v, want the course remote schedule", unit.Schedules)
		}
	})

	t.Run("missing_schedule_id_is_rejected_even_with_enrolled_unit", func(t *testing.T) {
		f := newChangeScheduleFixture(time.Now().Add(15 * 24 * time.Hour))
		d := &changeScheduleDeps{counts: map[uuid.UUID]int64{}}

		_, err := f.service(d).ChangeSchedule(ctx, f.inscricaoID, f.ownerCPF, &models.ScheduleChangeRequest{
			EnrolledUnit: forgedEnrolledUnit(),
		})
		assertRuleError(t, err, "selecione a nova turma")
		if d.updateCalls != 0 {
			t.Errorf("expected no Update, got %d", d.updateCalls)
		}
	})

	t.Run("schedule_of_another_course_is_rejected", func(t *testing.T) {
		f := newChangeScheduleFixture(time.Now().Add(15 * 24 * time.Hour))
		d := &changeScheduleDeps{counts: map[uuid.UUID]int64{}}
		foreign := uuid.New()

		_, err := f.service(d).ChangeSchedule(ctx, f.inscricaoID, f.ownerCPF, &models.ScheduleChangeRequest{
			ScheduleID:   &foreign,
			EnrolledUnit: forgedEnrolledUnit(),
		})
		assertRuleError(t, err, "não pertence a este curso")
		if d.updateCalls != 0 {
			t.Errorf("expected no Update, got %d", d.updateCalls)
		}
	})

	t.Run("schedule_not_accepting_enrollments_is_rejected", func(t *testing.T) {
		f := newChangeScheduleFixture(time.Now().Add(15 * 24 * time.Hour))
		closed := false
		f.curso.LocationClasses[0].Schedules[1].AcceptingEnrollments = &closed
		d := &changeScheduleDeps{counts: map[uuid.UUID]int64{}}

		_, err := f.service(d).ChangeSchedule(ctx, f.inscricaoID, f.ownerCPF, &models.ScheduleChangeRequest{
			ScheduleID: &f.targetSchedule,
		})
		assertRuleError(t, err, "não está aceitando inscrições")
		if d.updateCalls != 0 {
			t.Errorf("expected no Update, got %d", d.updateCalls)
		}
	})

	t.Run("full_schedule_is_rejected_by_course_vacancies", func(t *testing.T) {
		f := newChangeScheduleFixture(time.Now().Add(15 * 24 * time.Hour))
		// The target schedule has 10 vacancies in the course data.
		d := &changeScheduleDeps{counts: map[uuid.UUID]int64{f.targetSchedule: 10}}

		_, err := f.service(d).ChangeSchedule(ctx, f.inscricaoID, f.ownerCPF, &models.ScheduleChangeRequest{
			ScheduleID: &f.targetSchedule,
		})
		if err == nil || !strings.Contains(err.Error(), "não há vagas") {
			t.Fatalf("expected 'não há vagas' error, got %v", err)
		}
		if d.updateCalls != 0 {
			t.Errorf("expected no Update, got %d", d.updateCalls)
		}
	})

	t.Run("deadline_uses_the_real_class_start_not_the_body", func(t *testing.T) {
		// The real schedule starts in 24h (inside the 48h deadline); the
		// forged body claims a start one year away.
		f := newChangeScheduleFixture(time.Now().Add(24 * time.Hour))
		d := &changeScheduleDeps{counts: map[uuid.UUID]int64{}}

		_, err := f.service(d).ChangeSchedule(ctx, f.inscricaoID, f.ownerCPF, &models.ScheduleChangeRequest{
			ScheduleID:   &f.targetSchedule,
			EnrolledUnit: forgedEnrolledUnit(),
		})
		if err == nil || !strings.Contains(err.Error(), "menos de 48 horas") {
			t.Fatalf("expected deadline error, got %v", err)
		}
		if d.updateCalls != 0 {
			t.Errorf("expected no Update, got %d", d.updateCalls)
		}
	})

	t.Run("body_class_start_date_is_not_required", func(t *testing.T) {
		// Old clients may send an enrolled_unit without dates; the real
		// schedule date is what counts.
		f := newChangeScheduleFixture(time.Now().Add(15 * 24 * time.Hour))
		d := &changeScheduleDeps{counts: map[uuid.UUID]int64{}}

		_, err := f.service(d).ChangeSchedule(ctx, f.inscricaoID, f.ownerCPF, &models.ScheduleChangeRequest{
			ScheduleID:   &f.targetSchedule,
			EnrolledUnit: &models.EnrolledUnit{Schedules: []models.EnrolledUnitSchedule{{ClassStartDate: "data-invalida"}}},
		})
		if err != nil {
			t.Fatalf("ChangeSchedule failed: %v", err)
		}
	})

	t.Run("remote_schedule_without_class_start_is_rejected", func(t *testing.T) {
		f := newChangeScheduleFixture(time.Now().Add(15 * 24 * time.Hour))
		f.curso.RemoteClass.Schedules[0].ClassStartDate = nil
		d := &changeScheduleDeps{counts: map[uuid.UUID]int64{}}

		_, err := f.service(d).ChangeSchedule(ctx, f.inscricaoID, f.ownerCPF, &models.ScheduleChangeRequest{
			ScheduleID: &f.remoteScheduleID,
		})
		assertRuleError(t, err, "data de início")
		if d.updateCalls != 0 {
			t.Errorf("expected no Update, got %d", d.updateCalls)
		}
	})

	t.Run("vacancy_count_error_is_returned", func(t *testing.T) {
		f := newChangeScheduleFixture(time.Now().Add(15 * 24 * time.Hour))
		d := &changeScheduleDeps{countErr: errors.New("count query failed")}

		_, err := f.service(d).ChangeSchedule(ctx, f.inscricaoID, f.ownerCPF, &models.ScheduleChangeRequest{
			ScheduleID: &f.targetSchedule,
		})
		if err == nil || !strings.Contains(err.Error(), "erro ao verificar vagas") {
			t.Fatalf("expected vacancy count error, got %v", err)
		}
		if d.updateCalls != 0 {
			t.Errorf("expected no Update, got %d", d.updateCalls)
		}
	})

	t.Run("other_citizen_cannot_change", func(t *testing.T) {
		f := newChangeScheduleFixture(time.Now().Add(15 * 24 * time.Hour))
		d := &changeScheduleDeps{counts: map[uuid.UUID]int64{}}

		_, err := f.service(d).ChangeSchedule(ctx, f.inscricaoID, "98765432100", &models.ScheduleChangeRequest{
			ScheduleID: &f.targetSchedule,
		})
		if err == nil || err.Error() != "você só pode alterar suas próprias inscrições" {
			t.Fatalf("expected ownership error, got %v", err)
		}
		if d.updateCalls != 0 {
			t.Errorf("expected no Update, got %d", d.updateCalls)
		}
	})
}

// TestInscricaoService_ChangeSchedule_Guards keeps the checks that run before
// and after the turma validation: enrollment lookup, ownership, terminal
// status, single-turma courses, remote turma rules, default deadline and
// persistence errors.
func TestInscricaoService_ChangeSchedule_Guards(t *testing.T) {
	ctx := context.Background()
	future := time.Now().Add(15 * 24 * time.Hour)

	tests := []struct {
		name     string
		deps     changeScheduleDeps
		setup    func(f *changeScheduleFixture)
		schedule func(f *changeScheduleFixture) uuid.UUID
		wantErr  string
		wantRule bool
	}{
		{name: "enrollment_lookup_error", deps: changeScheduleDeps{getErr: errors.New("database timeout")}, wantErr: "erro ao buscar inscrição"},
		{name: "enrollment_not_found", deps: changeScheduleDeps{notFound: true}, wantErr: "inscrição não encontrada"},
		{name: "cancelled_enrollment", deps: changeScheduleDeps{status: models.StatusInscricaoCancelled}, wantErr: "não é possível trocar de turma para inscrições com status"},
		{name: "concluded_enrollment", deps: changeScheduleDeps{status: models.StatusInscricaoConcluded}, wantErr: "não é possível trocar de turma para inscrições com status"},
		{
			name: "course_with_a_single_turma",
			setup: func(f *changeScheduleFixture) {
				f.curso.RemoteClass = nil
				f.curso.LocationClasses[0].Schedules = f.curso.LocationClasses[0].Schedules[1:]
			},
			wantErr: "apenas uma turma",
		},
		{
			name: "remote_turma_not_accepting_enrollments",
			setup: func(f *changeScheduleFixture) {
				closed := false
				f.curso.RemoteClass.Schedules[0].AcceptingEnrollments = &closed
			},
			schedule: func(f *changeScheduleFixture) uuid.UUID { return f.remoteScheduleID },
			wantErr:  "não está aceitando inscrições",
			wantRule: true,
		},
		{
			name:     "remote_turma_full",
			schedule: func(f *changeScheduleFixture) uuid.UUID { return f.remoteScheduleID },
			wantErr:  "não há vagas",
		},
		{name: "update_error", deps: changeScheduleDeps{updateErr: errors.New("update failed")}, wantErr: "erro ao atualizar inscrição"},
		{name: "default_deadline_without_config", deps: changeScheduleDeps{nilConfig: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newChangeScheduleFixture(future)
			if tt.setup != nil {
				tt.setup(f)
			}
			d := tt.deps
			d.counts = map[uuid.UUID]int64{f.remoteScheduleID: 50}
			scheduleID := f.targetSchedule
			if tt.schedule != nil {
				scheduleID = tt.schedule(f)
			}

			_, err := f.service(&d).ChangeSchedule(ctx, f.inscricaoID, f.ownerCPF, &models.ScheduleChangeRequest{
				ScheduleID: &scheduleID,
			})
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("ChangeSchedule failed: %v", err)
				}
				if d.updateCalls != 1 {
					t.Errorf("expected one Update, got %d", d.updateCalls)
				}
				return
			}
			if tt.wantRule {
				assertRuleError(t, err, tt.wantErr)
			} else if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
			}
			if tt.deps.updateErr == nil && d.updateCalls != 0 {
				t.Errorf("expected no Update, got %d", d.updateCalls)
			}
		})
	}
}
