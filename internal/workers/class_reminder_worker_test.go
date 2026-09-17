package workers

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/prefeitura-rio/app-go-api/internal/config"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

type mockClassReminderLister struct {
	mu         sync.Mutex
	calls      [][2]time.Time
	inscricoes map[string][]*models.Inscricao
	err        error
}

func (m *mockClassReminderLister) ListApprovedStartingBetween(_ context.Context, start, end time.Time) ([]*models.Inscricao, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, [2]time.Time{start, end})
	if m.err != nil {
		return nil, m.err
	}
	if m.inscricoes == nil {
		return nil, nil
	}
	return m.inscricoes[start.Format("2006-01-02")], nil
}

type mockClassReminderNotifier struct {
	mu     sync.Mutex
	called int
	ids    []uuid.UUID
}

func (m *mockClassReminderNotifier) SendEnrollmentCreatedEmail(context.Context, *models.Inscricao, *models.Curso) error {
	return nil
}
func (m *mockClassReminderNotifier) SendEnrollmentApprovedEmail(context.Context, *models.Inscricao, *models.Curso) error {
	return nil
}
func (m *mockClassReminderNotifier) SendEnrollmentRejectedEmail(context.Context, *models.Inscricao, *models.Curso) error {
	return nil
}
func (m *mockClassReminderNotifier) SendEnrollmentConcludedEmail(context.Context, *models.Inscricao, *models.Curso) error {
	return nil
}
func (m *mockClassReminderNotifier) SendEnrollmentClassReminderEmail(_ context.Context, inscricao *models.Inscricao, _ *models.Curso) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.called++
	m.ids = append(m.ids, inscricao.ID)
	return nil
}
func (m *mockClassReminderNotifier) SendScheduleChangedEmail(context.Context, *models.Inscricao, *models.Curso) error {
	return nil
}
func (m *mockClassReminderNotifier) SendCandidaturaEnviadaEmail(context.Context, *empregabilidade.Candidatura) error {
	return nil
}
func (m *mockClassReminderNotifier) SendCandidaturaAprovadaEmail(context.Context, *empregabilidade.Candidatura) error {
	return nil
}
func (m *mockClassReminderNotifier) SendCandidaturaReprovadaEmail(context.Context, *empregabilidade.Candidatura) error {
	return nil
}
func (m *mockClassReminderNotifier) SendCandidaturaProximaEtapaEmail(context.Context, *empregabilidade.Candidatura, string) error {
	return nil
}

var _ services.EmailNotifier = (*mockClassReminderNotifier)(nil)

func TestTargetClassDateForOffset(t *testing.T) {
	now := time.Date(2026, 9, 17, 10, 0, 0, 0, time.Local)

	assert.Equal(t, time.Date(2026, 9, 18, 0, 0, 0, 0, time.Local), targetClassDateForOffset(now, -1))
	assert.Equal(t, time.Date(2026, 9, 15, 0, 0, 0, 0, time.Local), targetClassDateForOffset(now, 2))
}

func TestTemporalClassStartJobs_OnlyCatalogTemporalEmails(t *testing.T) {
	jobs := temporalClassStartJobs()
	require.Len(t, jobs, 1)
	assert.Equal(t, "enrollment.class_reminder_d1", jobs[0].Rule.ID)
	assert.Equal(t, -1, jobs[0].Rule.TemporalOffsetDays)
	assert.True(t, jobs[0].Rule.SendEnabled)
}

func TestClassReminderWorker_runCycle_SendsOnlyOwnTemporalEmail(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer redisClient.Close()

	d1ID := uuid.New()
	fixedNow := time.Date(2026, 9, 17, 10, 0, 0, 0, time.Local)
	d1Day := targetClassDateForOffset(fixedNow, -1)

	lister := &mockClassReminderLister{
		inscricoes: map[string][]*models.Inscricao{
			d1Day.Format("2006-01-02"): {
				{ID: d1ID, Email: "d1@test.com", Curso: &models.Curso{ID: 1, Titulo: "Curso D-1"}},
			},
		},
	}
	notifier := &mockClassReminderNotifier{}

	worker := NewClassReminderWorker(lister, notifier, redisClient, &config.ClassReminderSettings{
		Enabled:      true,
		SyncInterval: time.Hour,
	})
	worker.now = func() time.Time { return fixedNow }

	err = worker.runCycle(context.Background())
	require.NoError(t, err)

	// Only the D-1 rule is scanned (one temporal email in the catalog)
	lister.mu.Lock()
	assert.Equal(t, 1, len(lister.calls))
	lister.mu.Unlock()

	assert.Equal(t, 1, notifier.called)
	assert.Equal(t, d1ID, notifier.ids[0])

	err = worker.runCycle(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, notifier.called) // dedup
}

func TestClassReminderWorker_runCycle_SkipsWithoutCurso(t *testing.T) {
	fixedNow := time.Date(2026, 9, 17, 10, 0, 0, 0, time.Local)
	d1Day := targetClassDateForOffset(fixedNow, -1)

	lister := &mockClassReminderLister{
		inscricoes: map[string][]*models.Inscricao{
			d1Day.Format("2006-01-02"): {
				{ID: uuid.New(), Email: "aluno@test.com", Curso: nil},
			},
		},
	}
	notifier := &mockClassReminderNotifier{}

	worker := NewClassReminderWorker(lister, notifier, nil, &config.ClassReminderSettings{
		Enabled:      true,
		SyncInterval: time.Hour,
	})
	worker.now = func() time.Time { return fixedNow }

	err := worker.runCycle(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, notifier.called)
}

func TestClassReminderWorker_runCycle_SkipsWhenSendDisabled(t *testing.T) {
	fixedNow := time.Date(2026, 9, 17, 10, 0, 0, 0, time.Local)
	d1Day := targetClassDateForOffset(fixedNow, -1)
	id := uuid.New()

	lister := &mockClassReminderLister{
		inscricoes: map[string][]*models.Inscricao{
			d1Day.Format("2006-01-02"): {
				{ID: id, Email: "a@b.com", Curso: &models.Curso{ID: 1}},
			},
		},
	}
	notifier := &mockClassReminderNotifier{}

	worker := NewClassReminderWorker(lister, notifier, nil, &config.ClassReminderSettings{
		Enabled:      true,
		SyncInterval: time.Hour,
	})
	worker.now = func() time.Time { return fixedNow }
	worker.jobs = []TemporalEmailJob{{
		Rule: services.EmailCommunicationRule{
			ID:                 "enrollment.class_reminder_d1",
			Kind:               services.EmailTriggerTemporal,
			TemporalAnchor:     services.EmailAnchorClassStart,
			TemporalOffsetDays: -1,
			SendEnabled:        false,
		},
	}}

	err := worker.runCycle(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, notifier.called)
}

func TestClassReminderWorker_Start_Disabled(t *testing.T) {
	worker := NewClassReminderWorker(nil, nil, nil, &config.ClassReminderSettings{
		Enabled:      false,
		SyncInterval: time.Hour,
	})
	err := worker.Start(context.Background())
	assert.NoError(t, err)
}

func TestStartOfDay(t *testing.T) {
	ts := time.Date(2026, 9, 17, 15, 30, 45, 0, time.Local)
	got := startOfDay(ts)
	assert.Equal(t, 2026, got.Year())
	assert.Equal(t, time.September, got.Month())
	assert.Equal(t, 17, got.Day())
	assert.Equal(t, 0, got.Hour())
}

func TestClassReminderWorker_sendForRule_Unknown(t *testing.T) {
	worker := NewClassReminderWorker(nil, &mockClassReminderNotifier{}, nil, &config.ClassReminderSettings{Enabled: true})
	err := worker.sendForRule(context.Background(), services.EmailCommunicationRule{ID: "future.email_d2"}, &models.Inscricao{
		Curso: &models.Curso{ID: 1},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no email sender configured")
}
