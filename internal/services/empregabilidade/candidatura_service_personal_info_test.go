package empregabilidade_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
)

const testStaleThreshold = 15 * time.Minute

// Mock Citizen Snapshot Repo (the citizen_snapshots cache)
type MockCitizenSnapshotRepo struct {
	snapshots map[string]*models.CitizenSnapshot
	getError  error
}

func NewMockCitizenSnapshotRepo() *MockCitizenSnapshotRepo {
	return &MockCitizenSnapshotRepo{snapshots: make(map[string]*models.CitizenSnapshot)}
}

func (m *MockCitizenSnapshotRepo) GetByCPF(ctx context.Context, cpf string) (*models.CitizenSnapshot, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	return m.snapshots[cpf], nil
}

func (m *MockCitizenSnapshotRepo) GetByCPFs(ctx context.Context, cpfs []string) (map[string]*models.CitizenSnapshot, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	out := make(map[string]*models.CitizenSnapshot)
	for _, cpf := range cpfs {
		if s, ok := m.snapshots[cpf]; ok {
			out[cpf] = s
		}
	}
	return out, nil
}

// Mock Citizen Data Fetcher (RMI on-demand sync), recording every CPF it was asked to sync
type MockCitizenFetcher struct {
	syncedCPFs []string
	fresh      map[string]*models.CitizenSnapshot
	syncError  error
}

func NewMockCitizenFetcher() *MockCitizenFetcher {
	return &MockCitizenFetcher{fresh: make(map[string]*models.CitizenSnapshot)}
}

func (m *MockCitizenFetcher) SyncCitizenOnDemand(ctx context.Context, cpf string) (*models.CitizenSnapshot, error) {
	m.syncedCPFs = append(m.syncedCPFs, cpf)
	if m.syncError != nil {
		return nil, m.syncError
	}
	return m.fresh[cpf], nil
}

func (m *MockCitizenFetcher) StaleThreshold() time.Duration { return testStaleThreshold }

func snapshotSyncedAgo(celular string, age time.Duration) *models.CitizenSnapshot {
	return &models.CitizenSnapshot{
		Celular:      celular,
		LastSyncedAt: time.Now().Add(-age),
	}
}

// A citizen who only ever applied to a vaga is not covered by the enrollment scan
// in the sync worker, so the application itself must seed the snapshot.
func TestCandidaturaService_Create_SyncsCitizenSnapshot(t *testing.T) {
	mockCandidaturaRepo := NewMockCandidaturaRepo()
	mockVagaRepo := NewMockVagaRepo()
	fetcher := NewMockCitizenFetcher()

	vagaID := uuid.New()
	mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{ID: vagaID, Status: empregabilidade.StatusVagaPublicadoAtivo}

	service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, NewMockCurriculoService(), NewMockCitizenSnapshotRepo(), fetcher, newDisabledEmailSvc())

	_, err := service.Create(context.Background(), &empregabilidade.Candidatura{CPF: "12345678901", IDVaga: vagaID})
	if err != nil {
		t.Fatalf("Expected successful creation, got error: %v", err)
	}

	if len(fetcher.syncedCPFs) != 1 || fetcher.syncedCPFs[0] != "12345678901" {
		t.Errorf("Expected citizen snapshot to be synced for the applicant, got %v", fetcher.syncedCPFs)
	}
}

// The sync is best-effort: RMI being down must not cost the citizen their application.
func TestCandidaturaService_Create_SucceedsWhenSyncFails(t *testing.T) {
	mockCandidaturaRepo := NewMockCandidaturaRepo()
	mockVagaRepo := NewMockVagaRepo()
	fetcher := NewMockCitizenFetcher()
	fetcher.syncError = errors.New("RMI unavailable")

	vagaID := uuid.New()
	mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{ID: vagaID, Status: empregabilidade.StatusVagaPublicadoAtivo}

	service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, NewMockCurriculoService(), NewMockCitizenSnapshotRepo(), fetcher, newDisabledEmailSvc())

	id, err := service.Create(context.Background(), &empregabilidade.Candidatura{CPF: "12345678901", IDVaga: vagaID})
	if err != nil {
		t.Fatalf("Expected creation to survive a sync failure, got error: %v", err)
	}
	if id == uuid.Nil {
		t.Error("Expected non-nil UUID")
	}
}

func TestCandidaturaService_EnrichWithPersonalInfo(t *testing.T) {
	newService := func(repo *MockCitizenSnapshotRepo, fetcher *MockCitizenFetcher) *services.CandidaturaService {
		if fetcher == nil {
			return services.NewCandidaturaService(NewMockCandidaturaRepo(), NewMockVagaRepo(), NewMockCurriculoService(), repo, nil, newDisabledEmailSvc())
		}
		return services.NewCandidaturaService(NewMockCandidaturaRepo(), NewMockVagaRepo(), NewMockCurriculoService(), repo, fetcher, newDisabledEmailSvc())
	}

	t.Run("Refreshes a stale snapshot", func(t *testing.T) {
		repo := NewMockCitizenSnapshotRepo()
		repo.snapshots["12345678901"] = snapshotSyncedAgo("", 2*time.Hour)

		fetcher := NewMockCitizenFetcher()
		fetcher.fresh["12345678901"] = snapshotSyncedAgo("5521987719458", 0)

		c := &empregabilidade.Candidatura{CPF: "12345678901"}
		newService(repo, fetcher).EnrichWithPersonalInfo(context.Background(), c)

		if len(fetcher.syncedCPFs) != 1 {
			t.Fatalf("Expected the stale snapshot to be refreshed, syncs: %v", fetcher.syncedCPFs)
		}
		if c.PersonalInfo == nil || c.PersonalInfo.Celular != "5521987719458" {
			t.Errorf("Expected the refreshed phone to be served, got %+v", c.PersonalInfo)
		}
	})

	t.Run("Leaves a fresh snapshot untouched", func(t *testing.T) {
		repo := NewMockCitizenSnapshotRepo()
		repo.snapshots["12345678901"] = snapshotSyncedAgo("5521987719458", time.Minute)

		fetcher := NewMockCitizenFetcher()

		c := &empregabilidade.Candidatura{CPF: "12345678901"}
		newService(repo, fetcher).EnrichWithPersonalInfo(context.Background(), c)

		if len(fetcher.syncedCPFs) != 0 {
			t.Errorf("Expected no sync for a fresh snapshot, got %v", fetcher.syncedCPFs)
		}
		if c.PersonalInfo == nil || c.PersonalInfo.Celular != "5521987719458" {
			t.Errorf("Expected the cached phone to be served, got %+v", c.PersonalInfo)
		}
	})

	t.Run("Syncs when no snapshot exists", func(t *testing.T) {
		repo := NewMockCitizenSnapshotRepo()
		fetcher := NewMockCitizenFetcher()
		fetcher.fresh["12345678901"] = snapshotSyncedAgo("5521987719458", 0)

		c := &empregabilidade.Candidatura{CPF: "12345678901"}
		newService(repo, fetcher).EnrichWithPersonalInfo(context.Background(), c)

		if c.PersonalInfo == nil || c.PersonalInfo.Celular != "5521987719458" {
			t.Errorf("Expected personal_info to be populated from the on-demand sync, got %+v", c.PersonalInfo)
		}
	})

	t.Run("Falls back to the stale snapshot when the refresh fails", func(t *testing.T) {
		repo := NewMockCitizenSnapshotRepo()
		repo.snapshots["12345678901"] = snapshotSyncedAgo("5521999999999", 2*time.Hour)

		fetcher := NewMockCitizenFetcher()
		fetcher.syncError = errors.New("RMI unavailable")

		c := &empregabilidade.Candidatura{CPF: "12345678901"}
		newService(repo, fetcher).EnrichWithPersonalInfo(context.Background(), c)

		if c.PersonalInfo == nil || c.PersonalInfo.Celular != "5521999999999" {
			t.Errorf("Expected the stale snapshot to still be served, got %+v", c.PersonalInfo)
		}
	})

	t.Run("Does nothing when no fetcher is wired", func(t *testing.T) {
		repo := NewMockCitizenSnapshotRepo()
		repo.snapshots["12345678901"] = snapshotSyncedAgo("5521987719458", 2*time.Hour)

		c := &empregabilidade.Candidatura{CPF: "12345678901"}
		newService(repo, nil).EnrichWithPersonalInfo(context.Background(), c)

		if c.PersonalInfo == nil || c.PersonalInfo.Celular != "5521987719458" {
			t.Errorf("Expected the cached snapshot to be served as-is, got %+v", c.PersonalInfo)
		}
	})
}

func TestCandidaturaService_EnrichMultipleWithPersonalInfo(t *testing.T) {
	repo := NewMockCitizenSnapshotRepo()
	repo.snapshots["11111111111"] = snapshotSyncedAgo("5521111111111", time.Minute) // fresh
	repo.snapshots["22222222222"] = snapshotSyncedAgo("", 2*time.Hour)              // stale, phone missing
	// 33333333333 has no snapshot at all

	fetcher := NewMockCitizenFetcher()
	fetcher.fresh["22222222222"] = snapshotSyncedAgo("5521222222222", 0)
	fetcher.fresh["33333333333"] = snapshotSyncedAgo("5521333333333", 0)

	service := services.NewCandidaturaService(NewMockCandidaturaRepo(), NewMockVagaRepo(), NewMockCurriculoService(), repo, fetcher, newDisabledEmailSvc())

	candidaturas := []*empregabilidade.Candidatura{
		{CPF: "11111111111"},
		{CPF: "22222222222"},
		{CPF: "33333333333"},
	}
	service.EnrichMultipleWithPersonalInfo(context.Background(), candidaturas)

	if len(fetcher.syncedCPFs) != 2 {
		t.Errorf("Expected only the stale and the missing CPF to be synced, got %v", fetcher.syncedCPFs)
	}

	expected := map[string]string{
		"11111111111": "5521111111111",
		"22222222222": "5521222222222",
		"33333333333": "5521333333333",
	}
	for _, c := range candidaturas {
		if c.PersonalInfo == nil {
			t.Errorf("Expected personal_info for CPF %s", c.CPF)
			continue
		}
		if c.PersonalInfo.Celular != expected[c.CPF] {
			t.Errorf("CPF %s: expected phone %s, got %s", c.CPF, expected[c.CPF], c.PersonalInfo.Celular)
		}
	}
}
