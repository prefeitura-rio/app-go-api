package empregabilidade_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	coreservices "github.com/prefeitura-rio/app-go-api/internal/services"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
)

// mockEmpEmailNotifier records which Send* methods were called.
// A buffered channel signals async goroutine calls to the test.
type mockEmpEmailNotifier struct {
	mu       sync.Mutex
	called   map[string]int
	notifyCh chan string
}

func newMockEmpEmailNotifier() *mockEmpEmailNotifier {
	return &mockEmpEmailNotifier{
		called:   make(map[string]int),
		notifyCh: make(chan string, 8),
	}
}

func (m *mockEmpEmailNotifier) record(method string) {
	m.mu.Lock()
	m.called[method]++
	m.mu.Unlock()
	select {
	case m.notifyCh <- method:
	default:
	}
}

func (m *mockEmpEmailNotifier) callCount(method string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.called[method]
}

func (m *mockEmpEmailNotifier) waitForCall(method string, timeout time.Duration) bool {
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

func (m *mockEmpEmailNotifier) SendEnrollmentCreatedEmail(_ context.Context, _ *models.Inscricao, _ *models.Curso) error {
	m.record("enrollment.created")
	return nil
}
func (m *mockEmpEmailNotifier) SendEnrollmentApprovedEmail(_ context.Context, _ *models.Inscricao, _ *models.Curso) error {
	m.record("enrollment.approved")
	return nil
}
func (m *mockEmpEmailNotifier) SendEnrollmentRejectedEmail(_ context.Context, _ *models.Inscricao, _ *models.Curso) error {
	m.record("enrollment.rejected")
	return nil
}
func (m *mockEmpEmailNotifier) SendScheduleChangedEmail(_ context.Context, _ *models.Inscricao, _ *models.Curso) error {
	m.record("schedule.changed")
	return nil
}
func (m *mockEmpEmailNotifier) SendCandidaturaEnviadaEmail(_ context.Context, _ *empregabilidade.Candidatura) error {
	m.record("candidatura.enviada")
	return nil
}
func (m *mockEmpEmailNotifier) SendCandidaturaAprovadaEmail(_ context.Context, _ *empregabilidade.Candidatura) error {
	m.record("candidatura.aprovada")
	return nil
}
func (m *mockEmpEmailNotifier) SendCandidaturaReprovadaEmail(_ context.Context, _ *empregabilidade.Candidatura) error {
	m.record("candidatura.reprovada")
	return nil
}

var _ coreservices.EmailNotifier = (*mockEmpEmailNotifier)(nil)

// newSvcWithDisabledEmail builds a CandidaturaService wired with a disabled
// EmailNotificationService (no real sends) so existing behaviour is unchanged.
func newSvcWithDisabledEmail(repo *MockCandidaturaRepo, vagaRepo *MockVagaRepo) *services.CandidaturaService {
	return services.NewCandidaturaService(repo, vagaRepo, NewMockCurriculoService(), nil, nil, newDisabledEmailSvc())
}

// TestCandidaturaService_SetEmailNotifier verifies that SetEmailNotifier
// replaces the active email notifier and that subsequent operations route
// through the newly injected notifier.
func TestCandidaturaService_SetEmailNotifier(t *testing.T) {
	ctx := context.Background()

	makeVaga := func() (uuid.UUID, *MockVagaRepo) {
		vagaID := uuid.New()
		vagaRepo := NewMockVagaRepo()
		vagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
			ID:     vagaID,
			Status: empregabilidade.StatusVagaPublicadoAtivo,
		}
		return vagaID, vagaRepo
	}

	t.Run("Create — initial disabled notifier does not record calls on mock", func(t *testing.T) {
		vagaID, vagaRepo := makeVaga()
		repo := NewMockCandidaturaRepo()
		svc := newSvcWithDisabledEmail(repo, vagaRepo)

		mock := newMockEmpEmailNotifier()
		// Do NOT call SetEmailNotifier — initial notifier is the disabled service.

		_, err := svc.Create(ctx, &empregabilidade.Candidatura{CPF: "11111111111", IDVaga: vagaID})
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		time.Sleep(50 * time.Millisecond)
		if mock.callCount("candidatura.enviada") != 0 {
			t.Error("mock should not be called when it was never injected")
		}
	})

	t.Run("Create — after SetEmailNotifier triggers SendCandidaturaEnviadaEmail", func(t *testing.T) {
		vagaID, vagaRepo := makeVaga()
		repo := NewMockCandidaturaRepo()
		svc := newSvcWithDisabledEmail(repo, vagaRepo)

		mock := newMockEmpEmailNotifier()
		svc.SetEmailNotifier(mock)

		_, err := svc.Create(ctx, &empregabilidade.Candidatura{CPF: "22222222222", IDVaga: vagaID})
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		if !mock.waitForCall("candidatura.enviada", 200*time.Millisecond) {
			t.Error("expected SendCandidaturaEnviadaEmail to be called after SetEmailNotifier")
		}
	})

	t.Run("UpdateStatus approved — after SetEmailNotifier triggers SendCandidaturaAprovadaEmail", func(t *testing.T) {
		vagaID, vagaRepo := makeVaga()
		repo := NewMockCandidaturaRepo()
		svc := newSvcWithDisabledEmail(repo, vagaRepo)

		// Seed a candidatura in the repo so UpdateStatus can find it.
		candidaturaID := uuid.New()
		repo.candidaturas[candidaturaID] = &empregabilidade.Candidatura{
			ID:     candidaturaID,
			CPF:    "33333333333",
			IDVaga: vagaID,
			Status: empregabilidade.StatusCandidaturaEnviada,
		}

		mock := newMockEmpEmailNotifier()
		svc.SetEmailNotifier(mock)

		if err := svc.UpdateStatus(ctx, candidaturaID, empregabilidade.StatusCandidaturaAprovada); err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}
		if !mock.waitForCall("candidatura.aprovada", 200*time.Millisecond) {
			t.Error("expected SendCandidaturaAprovadaEmail to be called after SetEmailNotifier")
		}
	})

	t.Run("UpdateStatus reprovada — after SetEmailNotifier triggers SendCandidaturaReprovadaEmail", func(t *testing.T) {
		vagaID, vagaRepo := makeVaga()
		repo := NewMockCandidaturaRepo()
		svc := newSvcWithDisabledEmail(repo, vagaRepo)

		candidaturaID := uuid.New()
		repo.candidaturas[candidaturaID] = &empregabilidade.Candidatura{
			ID:     candidaturaID,
			CPF:    "44444444444",
			IDVaga: vagaID,
			Status: empregabilidade.StatusCandidaturaEnviada,
		}

		mock := newMockEmpEmailNotifier()
		svc.SetEmailNotifier(mock)

		if err := svc.UpdateStatus(ctx, candidaturaID, empregabilidade.StatusCandidaturaReprovada); err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}
		if !mock.waitForCall("candidatura.reprovada", 200*time.Millisecond) {
			t.Error("expected SendCandidaturaReprovadaEmail to be called after SetEmailNotifier")
		}
	})

	t.Run("SetEmailNotifier — replacing notifier uses new one going forward", func(t *testing.T) {
		vagaID, vagaRepo := makeVaga()
		repo := NewMockCandidaturaRepo()
		svc := newSvcWithDisabledEmail(repo, vagaRepo)

		first := newMockEmpEmailNotifier()
		second := newMockEmpEmailNotifier()

		svc.SetEmailNotifier(first)
		svc.SetEmailNotifier(second) // replace

		_, err := svc.Create(ctx, &empregabilidade.Candidatura{CPF: "55555555555", IDVaga: vagaID})
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		if !second.waitForCall("candidatura.enviada", 200*time.Millisecond) {
			t.Error("expected second notifier to be called")
		}
		time.Sleep(50 * time.Millisecond)
		if first.callCount("candidatura.enviada") != 0 {
			t.Error("first notifier should not be called after it was replaced")
		}
	})
}
