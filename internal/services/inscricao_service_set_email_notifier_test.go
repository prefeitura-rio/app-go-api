package services_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/config"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

// MockEmailNotifier records which Send methods were called. Methods signal the
// provided channel (if non-nil) so async callers can be detected in tests.
type MockEmailNotifier struct {
	mu       sync.Mutex
	called   map[string]int
	notifyCh chan string
}

func newMockEmailNotifier() *MockEmailNotifier {
	return &MockEmailNotifier{
		called:   make(map[string]int),
		notifyCh: make(chan string, 8),
	}
}

func (m *MockEmailNotifier) record(method string) {
	m.mu.Lock()
	m.called[method]++
	m.mu.Unlock()
	select {
	case m.notifyCh <- method:
	default:
	}
}

func (m *MockEmailNotifier) callCount(method string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.called[method]
}

// waitForCall blocks until the given method is observed or the timeout expires.
// Returns true if the call was detected within the timeout.
func (m *MockEmailNotifier) waitForCall(method string, timeout time.Duration) bool {
	deadline := time.After(timeout)
	for {
		select {
		case got := <-m.notifyCh:
			if got == method {
				return true
			}
		case <-deadline:
			return false
		}
	}
}

func (m *MockEmailNotifier) SendEnrollmentCreatedEmail(_ context.Context, _ *models.Inscricao, _ *models.Curso) error {
	m.record("enrollment.created")
	return nil
}
func (m *MockEmailNotifier) SendEnrollmentApprovedEmail(_ context.Context, _ *models.Inscricao, _ *models.Curso) error {
	m.record("enrollment.approved")
	return nil
}
func (m *MockEmailNotifier) SendEnrollmentRejectedEmail(_ context.Context, _ *models.Inscricao, _ *models.Curso) error {
	m.record("enrollment.rejected")
	return nil
}
func (m *MockEmailNotifier) SendScheduleChangedEmail(_ context.Context, _ *models.Inscricao, _ *models.Curso) error {
	m.record("schedule.changed")
	return nil
}
func (m *MockEmailNotifier) SendCandidaturaEnviadaEmail(_ context.Context, _ *empregabilidade.Candidatura) error {
	m.record("candidatura.enviada")
	return nil
}
func (m *MockEmailNotifier) SendCandidaturaAprovadaEmail(_ context.Context, _ *empregabilidade.Candidatura) error {
	m.record("candidatura.aprovada")
	return nil
}
func (m *MockEmailNotifier) SendCandidaturaReprovadaEmail(_ context.Context, _ *empregabilidade.Candidatura) error {
	m.record("candidatura.reprovada")
	return nil
}

var _ services.EmailNotifier = (*MockEmailNotifier)(nil)

// TestInscricaoService_SetEmailNotifier verifies that SetEmailNotifier replaces
// the active email notifier and that subsequent status-change actions route
// through the newly injected notifier.
func TestInscricaoService_SetEmailNotifier(t *testing.T) {
	ctx := context.Background()

	newSvcWithNilNotifier := func(inscricaoRepo MockInscricaoRepository) *services.InscricaoService {
		return services.NewInscricaoServiceWithInterface(
			&inscricaoRepo,
			&MockCursoRepository{
				GetByIDFunc: func(_ context.Context, _ int) (*models.Curso, error) {
					return &models.Curso{ID: 1, Titulo: "Curso Test"}, nil
				},
			},
			nil, nil,
			nil, // start with no notifier
			&config.AppConfig{},
		)
	}

	t.Run("nil notifier — UpdateStatus does not call any email method", func(t *testing.T) {
		id := uuid.New()
		repo := MockInscricaoRepository{
			GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{ID: id, CursoID: 1, Status: models.StatusInscricaoPending}, nil
			},
		}
		svc := newSvcWithNilNotifier(repo)

		// No notifier set — should not panic and should not call anything.
		if err := svc.UpdateStatus(ctx, id, models.StatusInscricaoApproved, "", ""); err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}
		// Give the goroutine a chance to run in case the nil guard is missing.
		time.Sleep(20 * time.Millisecond)
	})

	t.Run("SetEmailNotifier — approved status triggers SendEnrollmentApprovedEmail", func(t *testing.T) {
		id := uuid.New()
		repo := MockInscricaoRepository{
			GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{ID: id, CursoID: 1, Status: models.StatusInscricaoPending}, nil
			},
		}
		svc := newSvcWithNilNotifier(repo)

		mock := newMockEmailNotifier()
		svc.SetEmailNotifier(mock)

		if err := svc.UpdateStatus(ctx, id, models.StatusInscricaoApproved, "", ""); err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}
		if !mock.waitForCall("enrollment.approved", 200*time.Millisecond) {
			t.Error("expected SendEnrollmentApprovedEmail to be called after SetEmailNotifier")
		}
	})

	t.Run("SetEmailNotifier — rejected status triggers SendEnrollmentRejectedEmail", func(t *testing.T) {
		id := uuid.New()
		repo := MockInscricaoRepository{
			GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{ID: id, CursoID: 1, Status: models.StatusInscricaoPending}, nil
			},
		}
		svc := newSvcWithNilNotifier(repo)

		mock := newMockEmailNotifier()
		svc.SetEmailNotifier(mock)

		if err := svc.UpdateStatus(ctx, id, models.StatusInscricaoRejected, "motivo", ""); err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}
		if !mock.waitForCall("enrollment.rejected", 200*time.Millisecond) {
			t.Error("expected SendEnrollmentRejectedEmail to be called after SetEmailNotifier")
		}
	})

	t.Run("SetEmailNotifier — same status transition does not send email", func(t *testing.T) {
		id := uuid.New()
		repo := MockInscricaoRepository{
			GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.Inscricao, error) {
				// already approved
				return &models.Inscricao{ID: id, CursoID: 1, Status: models.StatusInscricaoApproved}, nil
			},
		}
		svc := newSvcWithNilNotifier(repo)

		mock := newMockEmailNotifier()
		svc.SetEmailNotifier(mock)

		if err := svc.UpdateStatus(ctx, id, models.StatusInscricaoApproved, "", ""); err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}
		time.Sleep(50 * time.Millisecond)
		if mock.callCount("enrollment.approved") != 0 {
			t.Error("email should not be sent when status does not change")
		}
	})

	t.Run("SetEmailNotifier — replacing notifier uses new one going forward", func(t *testing.T) {
		id := uuid.New()
		repo := MockInscricaoRepository{
			GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{ID: id, CursoID: 1, Status: models.StatusInscricaoPending}, nil
			},
		}
		svc := newSvcWithNilNotifier(repo)

		first := newMockEmailNotifier()
		second := newMockEmailNotifier()

		svc.SetEmailNotifier(first)
		svc.SetEmailNotifier(second) // replace

		if err := svc.UpdateStatus(ctx, id, models.StatusInscricaoApproved, "", ""); err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}
		if !second.waitForCall("enrollment.approved", 200*time.Millisecond) {
			t.Error("expected second notifier to be called")
		}
		time.Sleep(50 * time.Millisecond)
		if first.callCount("enrollment.approved") != 0 {
			t.Error("first notifier should not be called after it was replaced")
		}
	})
}
