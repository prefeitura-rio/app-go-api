package services_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/prefeitura-rio/app-go-api/internal/config"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

// enrollmentTable simulates the inscricoes table shared by concurrent requests
// and the per-turma lock of InscricaoRepository.WithScheduleLock (DEV-199).
// The count sleeps before answering to widen the window between the vacancy
// check and the write, which is where the TOCTOU happened.
type enrollmentTable struct {
	mu    sync.Mutex
	rows  map[uuid.UUID]*models.Inscricao
	locks sync.Map // schedule id -> *sync.Mutex
}

func newEnrollmentTable() *enrollmentTable {
	return &enrollmentTable{rows: make(map[uuid.UUID]*models.Inscricao)}
}

func (t *enrollmentTable) withScheduleLock(ctx context.Context, scheduleID uuid.UUID, fn func(ctx context.Context) error) error {
	l, _ := t.locks.LoadOrStore(scheduleID, &sync.Mutex{})
	mu := l.(*sync.Mutex)
	mu.Lock()
	defer mu.Unlock()
	return fn(ctx)
}

func (t *enrollmentTable) countOccupied(scheduleID uuid.UUID) int64 {
	t.mu.Lock()
	var n int64
	for _, row := range t.rows {
		if row.ScheduleID != nil && *row.ScheduleID == scheduleID &&
			(row.Status == models.StatusInscricaoApproved || row.Status == models.StatusInscricaoConcluded) {
			n++
		}
	}
	t.mu.Unlock()
	time.Sleep(2 * time.Millisecond)
	return n
}

func (t *enrollmentTable) insert(inscricao *models.Inscricao) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if inscricao.ID == uuid.Nil {
		inscricao.ID = uuid.New()
	}
	row := *inscricao
	t.rows[row.ID] = &row
}

func (t *enrollmentTable) get(id uuid.UUID) *models.Inscricao {
	t.mu.Lock()
	defer t.mu.Unlock()
	row, ok := t.rows[id]
	if !ok {
		return nil
	}
	cp := *row
	return &cp
}

func (t *enrollmentTable) countByStatus(scheduleID uuid.UUID, status models.StatusInscricao) int {
	t.mu.Lock()
	defer t.mu.Unlock()
	n := 0
	for _, row := range t.rows {
		if row.ScheduleID != nil && *row.ScheduleID == scheduleID && row.Status == status {
			n++
		}
	}
	return n
}

func (t *enrollmentTable) inscricaoRepo() *MockInscricaoRepository {
	return &MockInscricaoRepository{
		CreateFunc: func(_ context.Context, inscricao *models.Inscricao) error {
			t.insert(inscricao)
			return nil
		},
		GetByIDFunc: func(_ context.Context, id uuid.UUID) (*models.Inscricao, error) {
			return t.get(id), nil
		},
		UpdateFunc: func(_ context.Context, inscricao *models.Inscricao) error {
			t.insert(inscricao)
			return nil
		},
		WithScheduleLockFunc: t.withScheduleLock,
	}
}

func autoApproveCourse(scheduleID uuid.UUID, vacancies int) *models.Curso {
	accepting := true
	return &models.Curso{
		ID: 1,
		LocationClasses: []models.LocationClass{{
			ID: uuid.New(),
			Schedules: []models.CourseSchedule{{
				ID: scheduleID, Vacancies: vacancies, AcceptingEnrollments: &accepting,
			}},
		}},
	}
}

func (t *enrollmentTable) cursoRepo(curso *models.Curso) *MockCursoRepository {
	return &MockCursoRepository{
		ValidateForEnrollmentFunc: func(context.Context, int) (string, *time.Time, *time.Time, bool, error) {
			return string(models.StatusCursoOpened), nil, nil, true, nil
		},
		GetByIDFunc: func(context.Context, int) (*models.Curso, error) {
			return curso, nil
		},
		CountEnrollmentsByScheduleIDFunc: func(_ context.Context, sid uuid.UUID) (int64, error) {
			return t.countOccupied(sid), nil
		},
		CountEnrollmentsByScheduleIDsFunc: func(_ context.Context, ids []uuid.UUID) (map[uuid.UUID]int64, error) {
			counts := make(map[uuid.UUID]int64, len(ids))
			for _, id := range ids {
				counts[id] = t.countOccupied(id)
			}
			return counts, nil
		},
	}
}

// TestInscricaoService_CreatePaths_ConcurrentAutoApprove_RespectsVacancies fires
// simultaneous enrollments into one auto-approve turma and checks that exactly
// `vacancies` of them are approved; the others stay pending (DEV-199).
func TestInscricaoService_CreatePaths_ConcurrentAutoApprove_RespectsVacancies(t *testing.T) {
	const (
		vacancies = 5
		requests  = 20
	)
	paths := []struct {
		name   string
		create func(*services.InscricaoService, context.Context, *models.Inscricao) error
	}{
		{"Create", (*services.InscricaoService).Create},
		{"CreateByAdmin", (*services.InscricaoService).CreateByAdmin},
		{"CreateManual", (*services.InscricaoService).CreateManual},
	}

	for _, path := range paths {
		t.Run(path.name, func(t *testing.T) {
			scheduleID := uuid.New()
			table := newEnrollmentTable()
			svc := services.NewInscricaoServiceWithInterface(
				table.inscricaoRepo(), table.cursoRepo(autoApproveCourse(scheduleID, vacancies)),
				nil, nil, nil, &config.AppConfig{},
			)

			var wg sync.WaitGroup
			start := make(chan struct{})
			errs := make(chan error, requests)
			for i := 0; i < requests; i++ {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					sid := scheduleID
					inscricao := &models.Inscricao{
						CPF:        fmt.Sprintf("000000%05d", i),
						CursoID:    1,
						Name:       "Pessoa Sintética",
						ScheduleID: &sid,
					}
					<-start
					errs <- path.create(svc, context.Background(), inscricao)
				}(i)
			}
			close(start)
			wg.Wait()
			close(errs)
			for err := range errs {
				require.NoError(t, err)
			}

			assert.Equal(t, vacancies, table.countByStatus(scheduleID, models.StatusInscricaoApproved),
				"approved enrollments must not exceed the turma vacancies")
			assert.Equal(t, requests-vacancies, table.countByStatus(scheduleID, models.StatusInscricaoPending))
		})
	}
}

// TestInscricaoService_ChangeSchedule_Concurrent_RespectsVacancies moves several
// approved enrollments into the same turma at once: only as many as it has
// vacancies get in, the others are refused with "não há vagas" (DEV-199).
func TestInscricaoService_ChangeSchedule_Concurrent_RespectsVacancies(t *testing.T) {
	const (
		vacancies = 3
		requests  = 10
	)
	f := newChangeScheduleFixture(time.Now().Add(7 * 24 * time.Hour))
	f.curso.LocationClasses[0].Schedules[1].Vacancies = vacancies
	table := newEnrollmentTable()

	ids := make([]uuid.UUID, requests)
	cpfs := make([]string, requests)
	for i := range ids {
		ids[i] = uuid.New()
		cpfs[i] = fmt.Sprintf("111111%05d", i)
		current := f.currentSchedule
		table.insert(&models.Inscricao{
			ID: ids[i], CPF: cpfs[i], CursoID: f.curso.ID,
			Status: models.StatusInscricaoApproved, ScheduleID: &current,
		})
	}

	svc := services.NewInscricaoServiceWithInterface(
		table.inscricaoRepo(), table.cursoRepo(f.curso), nil, nil, nil,
		&config.AppConfig{Enrollment: config.EnrollmentSettings{ScheduleChangeDeadlineHours: 48}},
	)

	var wg sync.WaitGroup
	start := make(chan struct{})
	errs := make(chan error, requests)
	for i := 0; i < requests; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			target := f.targetSchedule
			<-start
			_, err := svc.ChangeSchedule(context.Background(), ids[i], cpfs[i], &models.ScheduleChangeRequest{ScheduleID: &target})
			errs <- err
		}(i)
	}
	close(start)
	wg.Wait()
	close(errs)

	moved, full := 0, 0
	for err := range errs {
		switch {
		case err == nil:
			moved++
		case strings.Contains(err.Error(), "não há vagas"):
			full++
		default:
			t.Fatalf("unexpected error: %v", err)
		}
	}
	assert.Equal(t, vacancies, moved)
	assert.Equal(t, requests-vacancies, full)
	assert.Equal(t, vacancies, table.countByStatus(f.targetSchedule, models.StatusInscricaoApproved),
		"the turma must not hold more approved enrollments than its vacancies")
}

// TestInscricaoService_Create_LockError_FallsBackToPending keeps the fail-safe:
// when the turma lock cannot be taken, the vacancy cannot be verified, so the
// enrollment is stored pending instead of failing or being approved.
func TestInscricaoService_Create_LockError_FallsBackToPending(t *testing.T) {
	scheduleID := uuid.New()
	var created []*models.Inscricao
	inscricaoRepo := &MockInscricaoRepository{
		CreateFunc: func(_ context.Context, inscricao *models.Inscricao) error {
			cp := *inscricao
			created = append(created, &cp)
			return nil
		},
		WithScheduleLockFunc: func(context.Context, uuid.UUID, func(context.Context) error) error {
			return errors.New("erro ao bloquear turma para controle de vagas: conexão perdida")
		},
	}
	cursoRepo := newEnrollmentTable().cursoRepo(autoApproveCourse(scheduleID, 10))
	svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

	err := svc.Create(context.Background(), &models.Inscricao{CPF: "00000000191", CursoID: 1, ScheduleID: &scheduleID})

	require.NoError(t, err)
	require.Len(t, created, 1)
	assert.Equal(t, models.StatusInscricaoPending, created[0].Status)
}

// TestInscricaoService_Create_CountError_FallsBackToPending keeps the existing
// fail-safe for a count error inside the locked section: the locked
// transaction is abandoned and the enrollment is stored pending.
func TestInscricaoService_Create_CountError_FallsBackToPending(t *testing.T) {
	scheduleID := uuid.New()
	var created []*models.Inscricao
	inscricaoRepo := &MockInscricaoRepository{
		CreateFunc: func(_ context.Context, inscricao *models.Inscricao) error {
			cp := *inscricao
			created = append(created, &cp)
			return nil
		},
	}
	cursoRepo := newEnrollmentTable().cursoRepo(autoApproveCourse(scheduleID, 10))
	cursoRepo.CountEnrollmentsByScheduleIDFunc = func(context.Context, uuid.UUID) (int64, error) {
		return 0, errors.New("db down")
	}
	svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

	err := svc.Create(context.Background(), &models.Inscricao{CPF: "00000000191", CursoID: 1, ScheduleID: &scheduleID})

	require.NoError(t, err)
	require.Len(t, created, 1)
	assert.Equal(t, models.StatusInscricaoPending, created[0].Status)
}

// TestInscricaoService_Create_InsertErrorUnderLock_IsReturned makes sure a
// failed insert inside the locked transaction is reported and not retried as
// pending (it would hit the same error, or duplicate the enrollment).
func TestInscricaoService_Create_InsertErrorUnderLock_IsReturned(t *testing.T) {
	scheduleID := uuid.New()
	calls := 0
	inscricaoRepo := &MockInscricaoRepository{
		CreateFunc: func(context.Context, *models.Inscricao) error {
			calls++
			return errors.New("erro ao criar inscrição: duplicate key")
		},
	}
	cursoRepo := newEnrollmentTable().cursoRepo(autoApproveCourse(scheduleID, 10))
	svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

	err := svc.Create(context.Background(), &models.Inscricao{CPF: "00000000191", CursoID: 1, ScheduleID: &scheduleID})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao criar inscrição")
	assert.Equal(t, 1, calls)
}

// TestInscricaoService_Create_NoVacancyLimitOrNoAutoApprove_DoesNotLock keeps
// the lock off the paths that do not consume a vacancy.
func TestInscricaoService_Create_NoVacancyLimitOrNoAutoApprove_DoesNotLock(t *testing.T) {
	cases := []struct {
		name        string
		vacancies   int
		autoApprove bool
		want        models.StatusInscricao
	}{
		{"turma without vacancy limit", 0, true, models.StatusInscricaoApproved},
		{"course without auto-approve", 10, false, models.StatusInscricaoPending},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			scheduleID := uuid.New()
			var created *models.Inscricao
			inscricaoRepo := &MockInscricaoRepository{
				CreateFunc: func(_ context.Context, inscricao *models.Inscricao) error {
					cp := *inscricao
					created = &cp
					return nil
				},
				WithScheduleLockFunc: func(context.Context, uuid.UUID, func(context.Context) error) error {
					t.Fatal("the turma lock must not be taken when no vacancy is consumed")
					return nil
				},
			}
			cursoRepo := newEnrollmentTable().cursoRepo(autoApproveCourse(scheduleID, tc.vacancies))
			cursoRepo.ValidateForEnrollmentFunc = func(context.Context, int) (string, *time.Time, *time.Time, bool, error) {
				return string(models.StatusCursoOpened), nil, nil, tc.autoApprove, nil
			}
			svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

			err := svc.Create(context.Background(), &models.Inscricao{CPF: "00000000191", CursoID: 1, ScheduleID: &scheduleID})

			require.NoError(t, err)
			require.NotNil(t, created)
			assert.Equal(t, tc.want, created.Status)
		})
	}
}

// TestInscricaoService_ChangeSchedule_LockError_DoesNotUpdate makes sure the turma
// change is refused, and nothing is written, when the lock cannot be taken.
func TestInscricaoService_ChangeSchedule_LockError_DoesNotUpdate(t *testing.T) {
	f := newChangeScheduleFixture(time.Now().Add(7 * 24 * time.Hour))
	inscricaoRepo := &MockInscricaoRepository{
		GetByIDFunc: func(context.Context, uuid.UUID) (*models.Inscricao, error) {
			current := f.currentSchedule
			return &models.Inscricao{ID: f.inscricaoID, CPF: f.ownerCPF, CursoID: f.curso.ID,
				Status: models.StatusInscricaoApproved, ScheduleID: &current}, nil
		},
		UpdateFunc: func(context.Context, *models.Inscricao) error {
			t.Fatal("Update must not run without the turma lock")
			return nil
		},
		WithScheduleLockFunc: func(context.Context, uuid.UUID, func(context.Context) error) error {
			return errors.New("erro ao bloquear turma para controle de vagas: conexão perdida")
		},
	}
	cursoRepo := newEnrollmentTable().cursoRepo(f.curso)
	svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil,
		&config.AppConfig{Enrollment: config.EnrollmentSettings{ScheduleChangeDeadlineHours: 48}})

	target := f.targetSchedule
	got, err := svc.ChangeSchedule(context.Background(), f.inscricaoID, f.ownerCPF, &models.ScheduleChangeRequest{ScheduleID: &target})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "erro ao verificar vagas")
}
