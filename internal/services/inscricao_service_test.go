package services_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/config"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

// MockInscricaoRepository implements InscricaoRepositoryInterface for testing
type MockInscricaoRepository struct {
	CreateFunc               func(ctx context.Context, inscricao *models.Inscricao) error
	GetByIDFunc              func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error)
	UpdateFunc               func(ctx context.Context, inscricao *models.Inscricao) error
	DeleteFunc               func(ctx context.Context, id uuid.UUID) error
	ExistsByCPFAndCursoFunc  func(ctx context.Context, cpf string, cursoID int) (bool, error)
	GetByCursoIDFunc         func(ctx context.Context, cursoID int, filter map[string]interface{}, limit, offset int) ([]*models.Inscricao, int, error)
	UpdateStatusFunc         func(ctx context.Context, inscricaoID uuid.UUID, status models.StatusInscricao, reason, adminNotes string) error
	UpdateMultipleStatusFunc func(ctx context.Context, inscricaoIDs []uuid.UUID, status models.StatusInscricao, reason, adminNotes string) (int, error)
	GetSummaryByCursoIDFunc  func(ctx context.Context, cursoID int) (*models.EnrollmentSummary, error)
	ListByCPFFunc            func(ctx context.Context, cpf string, filter map[string]interface{}, offset, limit int) ([]*models.Inscricao, int, error)
	UpdateCertificateFunc    func(ctx context.Context, inscricaoID uuid.UUID, certificateURL string) error
}

func (m *MockInscricaoRepository) Create(ctx context.Context, inscricao *models.Inscricao) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, inscricao)
	}
	inscricao.ID = uuid.New()
	return nil
}

func (m *MockInscricaoRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockInscricaoRepository) Update(ctx context.Context, inscricao *models.Inscricao) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, inscricao)
	}
	return nil
}

func (m *MockInscricaoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockInscricaoRepository) ExistsByCPFAndCurso(ctx context.Context, cpf string, cursoID int) (bool, error) {
	if m.ExistsByCPFAndCursoFunc != nil {
		return m.ExistsByCPFAndCursoFunc(ctx, cpf, cursoID)
	}
	return false, nil
}

func (m *MockInscricaoRepository) GetByCursoID(ctx context.Context, cursoID int, filter map[string]interface{}, limit, offset int) ([]*models.Inscricao, int, error) {
	if m.GetByCursoIDFunc != nil {
		return m.GetByCursoIDFunc(ctx, cursoID, filter, limit, offset)
	}
	return []*models.Inscricao{}, 0, nil
}

func (m *MockInscricaoRepository) UpdateStatus(ctx context.Context, inscricaoID uuid.UUID, status models.StatusInscricao, reason, adminNotes string) error {
	if m.UpdateStatusFunc != nil {
		return m.UpdateStatusFunc(ctx, inscricaoID, status, reason, adminNotes)
	}
	return nil
}

func (m *MockInscricaoRepository) UpdateMultipleStatus(ctx context.Context, inscricaoIDs []uuid.UUID, status models.StatusInscricao, reason, adminNotes string) (int, error) {
	if m.UpdateMultipleStatusFunc != nil {
		return m.UpdateMultipleStatusFunc(ctx, inscricaoIDs, status, reason, adminNotes)
	}
	return len(inscricaoIDs), nil
}

func (m *MockInscricaoRepository) GetSummaryByCursoID(ctx context.Context, cursoID int) (*models.EnrollmentSummary, error) {
	if m.GetSummaryByCursoIDFunc != nil {
		return m.GetSummaryByCursoIDFunc(ctx, cursoID)
	}
	return &models.EnrollmentSummary{}, nil
}

func (m *MockInscricaoRepository) ListByCPF(ctx context.Context, cpf string, filter map[string]interface{}, offset, limit int) ([]*models.Inscricao, int, error) {
	if m.ListByCPFFunc != nil {
		return m.ListByCPFFunc(ctx, cpf, filter, offset, limit)
	}
	return []*models.Inscricao{}, 0, nil
}

func (m *MockInscricaoRepository) UpdateCertificate(ctx context.Context, inscricaoID uuid.UUID, certificateURL string) error {
	if m.UpdateCertificateFunc != nil {
		return m.UpdateCertificateFunc(ctx, inscricaoID, certificateURL)
	}
	return nil
}

// MockCitizenDataFetcher for testing
type MockCitizenDataFetcher struct {
	SyncFunc func(ctx context.Context, cpf string) (*models.CitizenSnapshot, error)
}

func (m *MockCitizenDataFetcher) SyncCitizenOnDemand(ctx context.Context, cpf string) (*models.CitizenSnapshot, error) {
	if m.SyncFunc != nil {
		return m.SyncFunc(ctx, cpf)
	}
	return nil, nil
}

// TestInscricaoService_Create tests the Create method
func TestInscricaoService_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("Create valid inscricao with auto-approve", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				if inscricao.Status != models.StatusInscricaoApproved {
					t.Errorf("Expected status APPROVED, got %s", inscricao.Status)
				}
				return nil
			},
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, nil
			},
		}

		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (status string, enrollmentStart *time.Time, enrollmentEnd *time.Time, autoApprove bool, err error) {
				return string(models.StatusCursoOpened), nil, nil, true, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricao := &models.Inscricao{
			CPF:     "12345678900",
			CursoID: 1,
			Name:    "Test User",
			Email:   "test@example.com",
		}

		err := svc.Create(ctx, inscricao)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	})

	t.Run("Create valid inscricao without auto-approve", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				if inscricao.Status != models.StatusInscricaoPending {
					t.Errorf("Expected status PENDING, got %s", inscricao.Status)
				}
				return nil
			},
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, nil
			},
		}

		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (status string, enrollmentStart *time.Time, enrollmentEnd *time.Time, autoApprove bool, err error) {
				return string(models.StatusCursoOpened), nil, nil, false, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricao := &models.Inscricao{
			CPF:     "12345678900",
			CursoID: 1,
			Name:    "Test User",
			Email:   "test@example.com",
		}

		err := svc.Create(ctx, inscricao)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	})

	t.Run("Create fails when curso not opened", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{}

		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (status string, enrollmentStart *time.Time, enrollmentEnd *time.Time, autoApprove bool, err error) {
				return string(models.StatusCursoDraft), nil, nil, false, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricao := &models.Inscricao{
			CPF:     "12345678900",
			CursoID: 1,
		}

		err := svc.Create(ctx, inscricao)
		if err == nil {
			t.Error("Expected error when curso not opened")
		}
	})

	t.Run("Create fails when enrollment period not started", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{}

		futureDate := time.Now().Add(24 * time.Hour)
		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (status string, enrollmentStart *time.Time, enrollmentEnd *time.Time, autoApprove bool, err error) {
				return string(models.StatusCursoOpened), &futureDate, nil, false, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricao := &models.Inscricao{
			CPF:     "12345678900",
			CursoID: 1,
		}

		err := svc.Create(ctx, inscricao)
		if err == nil {
			t.Error("Expected error when enrollment period not started")
		}
	})

	t.Run("Create fails when enrollment period ended", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{}

		pastDate := time.Now().Add(-24 * time.Hour)
		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (status string, enrollmentStart *time.Time, enrollmentEnd *time.Time, autoApprove bool, err error) {
				return string(models.StatusCursoOpened), nil, &pastDate, false, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricao := &models.Inscricao{
			CPF:     "12345678900",
			CursoID: 1,
		}

		err := svc.Create(ctx, inscricao)
		if err == nil {
			t.Error("Expected error when enrollment period ended")
		}
	})

	t.Run("Create fails when CPF already enrolled", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return true, nil
			},
		}

		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (status string, enrollmentStart *time.Time, enrollmentEnd *time.Time, autoApprove bool, err error) {
				return string(models.StatusCursoOpened), nil, nil, false, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricao := &models.Inscricao{
			CPF:     "12345678900",
			CursoID: 1,
		}

		err := svc.Create(ctx, inscricao)
		if err == nil {
			t.Error("Expected error when CPF already enrolled")
		}
	})

	t.Run("Create with citizen data fetcher", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				if inscricao.Email != "fetched@example.com" {
					t.Errorf("Expected email from fetcher, got %s", inscricao.Email)
				}
				if inscricao.Phone != "1234567890" {
					t.Errorf("Expected phone from fetcher, got %s", inscricao.Phone)
				}
				return nil
			},
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, nil
			},
		}

		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (status string, enrollmentStart *time.Time, enrollmentEnd *time.Time, autoApprove bool, err error) {
				return string(models.StatusCursoOpened), nil, nil, false, nil
			},
		}

		citizenFetcher := &MockCitizenDataFetcher{
			SyncFunc: func(ctx context.Context, cpf string) (*models.CitizenSnapshot, error) {
				return &models.CitizenSnapshot{
					CPF:     cpf,
					Email:   "fetched@example.com",
					Celular: "1234567890",
					Nome:    "Fetched Name",
				}, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, citizenFetcher, nil, &config.AppConfig{})

		inscricao := &models.Inscricao{
			CPF:     "12345678900",
			CursoID: 1,
		}

		err := svc.Create(ctx, inscricao)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	})
}

// TestInscricaoService_UpdateStatus tests the UpdateStatus method
func TestInscricaoService_UpdateStatus(t *testing.T) {
	ctx := context.Background()

	t.Run("UpdateStatus from pending to approved", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CursoID: 1,
					Status:  models.StatusInscricaoPending,
				}, nil
			},
			UpdateStatusFunc: func(ctx context.Context, id uuid.UUID, status models.StatusInscricao, reason, adminNotes string) error {
				if status != models.StatusInscricaoApproved {
					t.Errorf("Expected status APPROVED, got %s", status)
				}
				return nil
			},
		}

		cursoRepo := &MockCursoRepository{
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return &models.Curso{ID: 1, Titulo: "Test Course"}, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.UpdateStatus(ctx, inscricaoID, models.StatusInscricaoApproved, "", "")
		if err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}
	})

	t.Run("UpdateStatus fails when inscricao not found", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return nil, nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.UpdateStatus(ctx, inscricaoID, models.StatusInscricaoApproved, "", "")
		if err == nil {
			t.Error("Expected error when inscricao not found")
		}
	})

	t.Run("UpdateStatus with repository error", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CursoID: 1,
					Status:  models.StatusInscricaoPending,
				}, nil
			},
			UpdateStatusFunc: func(ctx context.Context, id uuid.UUID, status models.StatusInscricao, reason, adminNotes string) error {
				return errors.New("database error")
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.UpdateStatus(ctx, inscricaoID, models.StatusInscricaoApproved, "", "")
		if err == nil {
			t.Error("Expected database error")
		}
	})

	t.Run("UpdateStatus with GetByID error", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return nil, errors.New("database connection error")
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.UpdateStatus(ctx, inscricaoID, models.StatusInscricaoApproved, "", "")
		if err == nil {
			t.Error("Expected database error")
		}
		if !errors.Is(err, errors.New("database connection error")) && !strings.Contains(err.Error(), "erro ao verificar inscrição") {
			t.Errorf("Expected verification error, got: %v", err)
		}
	})

	t.Run("UpdateStatus status unchanged no email", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CursoID: 1,
					Status:  models.StatusInscricaoApproved, // Same status
				}, nil
			},
			UpdateStatusFunc: func(ctx context.Context, id uuid.UUID, status models.StatusInscricao, reason, adminNotes string) error {
				return nil
			},
		}

		cursoRepo := &MockCursoRepository{
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return &models.Curso{ID: 1, Titulo: "Test Course"}, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		// Update to same status - should not send email
		err := svc.UpdateStatus(ctx, inscricaoID, models.StatusInscricaoApproved, "", "")
		if err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}
	})

	t.Run("UpdateStatus to rejected status", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CursoID: 1,
					Status:  models.StatusInscricaoPending,
				}, nil
			},
			UpdateStatusFunc: func(ctx context.Context, id uuid.UUID, status models.StatusInscricao, reason, adminNotes string) error {
				if status != models.StatusInscricaoRejected {
					t.Errorf("Expected status REJECTED, got %s", status)
				}
				return nil
			},
		}

		cursoRepo := &MockCursoRepository{
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return &models.Curso{ID: 1, Titulo: "Test Course"}, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.UpdateStatus(ctx, inscricaoID, models.StatusInscricaoRejected, "Test reason", "Admin notes")
		if err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}
	})
}

// TestInscricaoService_UpdateMultipleStatus tests the UpdateMultipleStatus method
func TestInscricaoService_UpdateMultipleStatus(t *testing.T) {
	ctx := context.Background()

	t.Run("UpdateMultipleStatus success", func(t *testing.T) {
		ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      id,
					CursoID: 1,
					Status:  models.StatusInscricaoPending,
				}, nil
			},
			UpdateMultipleStatusFunc: func(ctx context.Context, inscricaoIDs []uuid.UUID, status models.StatusInscricao, reason, adminNotes string) (int, error) {
				return len(inscricaoIDs), nil
			},
		}

		cursoRepo := &MockCursoRepository{
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return &models.Curso{ID: 1, Titulo: "Test Course"}, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		count, err := svc.UpdateMultipleStatus(ctx, ids, models.StatusInscricaoApproved, "", "")
		if err != nil {
			t.Fatalf("UpdateMultipleStatus failed: %v", err)
		}
		if count != 3 {
			t.Errorf("Expected count 3, got %d", count)
		}
	})

	t.Run("UpdateMultipleStatus with empty list fails", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{}
		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		_, err := svc.UpdateMultipleStatus(ctx, []uuid.UUID{}, models.StatusInscricaoApproved, "", "")
		if err == nil {
			t.Error("Expected error when list is empty")
		}
	})

	t.Run("UpdateMultipleStatus with repository error", func(t *testing.T) {
		ids := []uuid.UUID{uuid.New(), uuid.New()}

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      id,
					CursoID: 1,
					Status:  models.StatusInscricaoPending,
				}, nil
			},
			UpdateMultipleStatusFunc: func(ctx context.Context, inscricaoIDs []uuid.UUID, status models.StatusInscricao, reason, adminNotes string) (int, error) {
				return 0, errors.New("bulk update failed")
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		count, err := svc.UpdateMultipleStatus(ctx, ids, models.StatusInscricaoApproved, "", "")
		if err == nil {
			t.Error("Expected database error")
		}
		if count != 0 {
			t.Errorf("Expected count 0 on error, got %d", count)
		}
	})

	t.Run("UpdateMultipleStatus partial success", func(t *testing.T) {
		ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      id,
					CursoID: 1,
					Status:  models.StatusInscricaoPending,
				}, nil
			},
			UpdateMultipleStatusFunc: func(ctx context.Context, inscricaoIDs []uuid.UUID, status models.StatusInscricao, reason, adminNotes string) (int, error) {
				// Partial success: only 2 out of 3 updated
				return 2, nil
			},
		}

		cursoRepo := &MockCursoRepository{
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return &models.Curso{ID: 1, Titulo: "Test Course"}, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		count, err := svc.UpdateMultipleStatus(ctx, ids, models.StatusInscricaoApproved, "", "")
		if err != nil {
			t.Fatalf("UpdateMultipleStatus failed: %v", err)
		}
		if count != 2 {
			t.Errorf("Expected count 2, got %d", count)
		}
	})

	t.Run("UpdateMultipleStatus with GetByID error during email collection", func(t *testing.T) {
		ids := []uuid.UUID{uuid.New(), uuid.New()}
		callCount := 0

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				callCount++
				if callCount == 1 {
					// First call succeeds
					return &models.Inscricao{
						ID:      id,
						CursoID: 1,
						Status:  models.StatusInscricaoPending,
					}, nil
				}
				// Second call fails
				return nil, errors.New("database error")
			},
			UpdateMultipleStatusFunc: func(ctx context.Context, inscricaoIDs []uuid.UUID, status models.StatusInscricao, reason, adminNotes string) (int, error) {
				return len(inscricaoIDs), nil
			},
		}

		cursoRepo := &MockCursoRepository{
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return &models.Curso{ID: 1, Titulo: "Test Course"}, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		// Should still succeed despite GetByID error during email collection
		count, err := svc.UpdateMultipleStatus(ctx, ids, models.StatusInscricaoApproved, "", "")
		if err != nil {
			t.Fatalf("UpdateMultipleStatus should succeed even with email collection errors: %v", err)
		}
		if count != 2 {
			t.Errorf("Expected count 2, got %d", count)
		}
	})

	t.Run("UpdateMultipleStatus to rejected status", func(t *testing.T) {
		ids := []uuid.UUID{uuid.New(), uuid.New()}

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      id,
					CursoID: 1,
					Status:  models.StatusInscricaoPending,
				}, nil
			},
			UpdateMultipleStatusFunc: func(ctx context.Context, inscricaoIDs []uuid.UUID, status models.StatusInscricao, reason, adminNotes string) (int, error) {
				if status != models.StatusInscricaoRejected {
					t.Errorf("Expected status REJECTED, got %s", status)
				}
				return len(inscricaoIDs), nil
			},
		}

		cursoRepo := &MockCursoRepository{
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return &models.Curso{ID: 1, Titulo: "Test Course"}, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		count, err := svc.UpdateMultipleStatus(ctx, ids, models.StatusInscricaoRejected, "Rejected reason", "Admin notes")
		if err != nil {
			t.Fatalf("UpdateMultipleStatus failed: %v", err)
		}
		if count != 2 {
			t.Errorf("Expected count 2, got %d", count)
		}
	})

	t.Run("UpdateMultipleStatus status not approved or rejected", func(t *testing.T) {
		ids := []uuid.UUID{uuid.New(), uuid.New()}

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      id,
					CursoID: 1,
					Status:  models.StatusInscricaoPending,
				}, nil
			},
			UpdateMultipleStatusFunc: func(ctx context.Context, inscricaoIDs []uuid.UUID, status models.StatusInscricao, reason, adminNotes string) (int, error) {
				return len(inscricaoIDs), nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		// Update to pending (not approved/rejected) - should not trigger email collection
		count, err := svc.UpdateMultipleStatus(ctx, ids, models.StatusInscricaoPending, "", "")
		if err != nil {
			t.Fatalf("UpdateMultipleStatus failed: %v", err)
		}
		if count != 2 {
			t.Errorf("Expected count 2, got %d", count)
		}
	})

	t.Run("UpdateMultipleStatus to concluded status", func(t *testing.T) {
		ids := []uuid.UUID{uuid.New(), uuid.New()}

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      id,
					CursoID: 1,
					Status:  models.StatusInscricaoApproved,
				}, nil
			},
			UpdateMultipleStatusFunc: func(ctx context.Context, inscricaoIDs []uuid.UUID, status models.StatusInscricao, reason, adminNotes string) (int, error) {
				if status != models.StatusInscricaoConcluded {
					t.Errorf("Expected status CONCLUDED, got %s", status)
				}
				return len(inscricaoIDs), nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		count, err := svc.UpdateMultipleStatus(ctx, ids, models.StatusInscricaoConcluded, "", "Curso concluído")
		if err != nil {
			t.Fatalf("UpdateMultipleStatus failed: %v", err)
		}
		if count != 2 {
			t.Errorf("Expected count 2, got %d", count)
		}
	})

	t.Run("UpdateMultipleStatus without email service doesn't call GetByID", func(t *testing.T) {
		ids := []uuid.UUID{uuid.New(), uuid.New()}

		callCount := 0

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				callCount++
				return &models.Inscricao{
					ID:      id,
					CursoID: 1,
					Status:  models.StatusInscricaoPending,
				}, nil
			},
			UpdateMultipleStatusFunc: func(ctx context.Context, inscricaoIDs []uuid.UUID, status models.StatusInscricao, reason, adminNotes string) (int, error) {
				return len(inscricaoIDs), nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		// No email notification service
		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		count, err := svc.UpdateMultipleStatus(ctx, ids, models.StatusInscricaoApproved, "", "")
		if err != nil {
			t.Fatalf("UpdateMultipleStatus failed: %v", err)
		}
		if count != 2 {
			t.Errorf("Expected count 2, got %d", count)
		}
		// GetByID should not be called when email service is nil
		if callCount != 0 {
			t.Errorf("Expected GetByID not to be called without email service, called %d times", callCount)
		}
	})

	t.Run("UpdateMultipleStatus with zero updates", func(t *testing.T) {
		ids := []uuid.UUID{uuid.New()}

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      id,
					CursoID: 1,
					Status:  models.StatusInscricaoPending,
				}, nil
			},
			UpdateMultipleStatusFunc: func(ctx context.Context, inscricaoIDs []uuid.UUID, status models.StatusInscricao, reason, adminNotes string) (int, error) {
				// Simulate zero rows updated (maybe due to concurrent update or constraint)
				return 0, nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		count, err := svc.UpdateMultipleStatus(ctx, ids, models.StatusInscricaoApproved, "", "")
		if err != nil {
			t.Fatalf("UpdateMultipleStatus failed: %v", err)
		}
		if count != 0 {
			t.Errorf("Expected count 0, got %d", count)
		}
	})

	t.Run("UpdateMultipleStatus with all same status (no email sent)", func(t *testing.T) {
		ids := []uuid.UUID{uuid.New(), uuid.New()}

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      id,
					CursoID: 1,
					Status:  models.StatusInscricaoApproved, // Already approved
				}, nil
			},
			UpdateMultipleStatusFunc: func(ctx context.Context, inscricaoIDs []uuid.UUID, status models.StatusInscricao, reason, adminNotes string) (int, error) {
				return len(inscricaoIDs), nil
			},
		}

		cursoRepo := &MockCursoRepository{
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return &models.Curso{ID: 1, Titulo: "Test Course"}, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		// Update to same status - should collect emails but skip sending
		count, err := svc.UpdateMultipleStatus(ctx, ids, models.StatusInscricaoApproved, "", "")
		if err != nil {
			t.Fatalf("UpdateMultipleStatus failed: %v", err)
		}
		if count != 2 {
			t.Errorf("Expected count 2, got %d", count)
		}
	})

	t.Run("UpdateMultipleStatus with curso lookup failure during email", func(t *testing.T) {
		ids := []uuid.UUID{uuid.New()}

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      id,
					CursoID: 1,
					Status:  models.StatusInscricaoPending,
					Email:   "user@example.com",
				}, nil
			},
			UpdateMultipleStatusFunc: func(ctx context.Context, inscricaoIDs []uuid.UUID, status models.StatusInscricao, reason, adminNotes string) (int, error) {
				return 1, nil
			},
		}

		cursoRepo := &MockCursoRepository{
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return nil, errors.New("curso not found")
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		// Should still succeed even if curso lookup fails during email sending
		count, err := svc.UpdateMultipleStatus(ctx, ids, models.StatusInscricaoApproved, "", "")
		if err != nil {
			t.Fatalf("UpdateMultipleStatus should succeed even if curso lookup fails: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected count 1, got %d", count)
		}
	})
}

// TestInscricaoService_Delete tests the Delete method
func TestInscricaoService_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("Delete existing inscricao", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{ID: inscricaoID}, nil
			},
			DeleteFunc: func(ctx context.Context, id uuid.UUID) error {
				return nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.Delete(ctx, inscricaoID)
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}
	})

	t.Run("Delete fails when inscricao not found", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return nil, nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.Delete(ctx, inscricaoID)
		if err == nil {
			t.Error("Expected error when inscricao not found")
		}
	})
}

// TestInscricaoService_UpdateCertificate tests the UpdateCertificate method
func TestInscricaoService_UpdateCertificate(t *testing.T) {
	ctx := context.Background()

	t.Run("UpdateCertificate for approved inscricao", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CursoID: 1,
					Status:  models.StatusInscricaoApproved,
				}, nil
			},
			UpdateCertificateFunc: func(ctx context.Context, id uuid.UUID, certificateURL string) error {
				if certificateURL != "https://example.com/cert.pdf" {
					t.Errorf("Expected certificate URL, got %s", certificateURL)
				}
				return nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.UpdateCertificate(ctx, 1, inscricaoID, "https://example.com/cert.pdf")
		if err != nil {
			t.Fatalf("UpdateCertificate failed: %v", err)
		}
	})

	t.Run("UpdateCertificate fails when inscricao from different curso", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CursoID: 2, // Different curso ID
					Status:  models.StatusInscricaoApproved,
				}, nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.UpdateCertificate(ctx, 1, inscricaoID, "https://example.com/cert.pdf")
		if err == nil {
			t.Error("Expected error when inscricao from different curso")
		}
	})

	t.Run("UpdateCertificate fails for pending inscricao", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CursoID: 1,
					Status:  models.StatusInscricaoPending,
				}, nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.UpdateCertificate(ctx, 1, inscricaoID, "https://example.com/cert.pdf")
		if err == nil {
			t.Error("Expected error for pending inscricao")
		}
	})
}

// TestInscricaoService_GetSummaryByCursoID tests the GetSummaryByCursoID method
func TestInscricaoService_GetSummaryByCursoID(t *testing.T) {
	ctx := context.Background()

	t.Run("GetSummaryByCursoID success", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			GetSummaryByCursoIDFunc: func(ctx context.Context, cursoID int) (*models.EnrollmentSummary, error) {
				return &models.EnrollmentSummary{
					Total:    100,
					Pending:  10,
					Approved: 80,
					Rejected: 10,
				}, nil
			},
		}

		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (status string, enrollmentStart *time.Time, enrollmentEnd *time.Time, autoApprove bool, err error) {
				return string(models.StatusCursoOpened), nil, nil, false, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		summary, err := svc.GetSummaryByCursoID(ctx, 1)
		if err != nil {
			t.Fatalf("GetSummaryByCursoID failed: %v", err)
		}
		if summary.Total != 100 {
			t.Errorf("Expected total 100, got %d", summary.Total)
		}
	})

	t.Run("GetSummaryByCursoID fails when curso not found", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{}

		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (status string, enrollmentStart *time.Time, enrollmentEnd *time.Time, autoApprove bool, err error) {
				return "", nil, nil, false, errors.New("curso not found")
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		_, err := svc.GetSummaryByCursoID(ctx, 999)
		if err == nil {
			t.Error("Expected error when curso not found")
		}
	})
}

// TestInscricaoService_GetByID tests the GetByID method
func TestInscricaoService_GetByID(t *testing.T) {
	ctx := context.Background()

	t.Run("GetByID existing inscricao", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CursoID: 1,
					CPF:     "12345678900",
				}, nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricao, err := svc.GetByID(ctx, inscricaoID)
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}
		if inscricao == nil {
			t.Fatal("Expected inscricao, got nil")
		}
		if inscricao.ID != inscricaoID {
			t.Errorf("Expected ID %s, got %s", inscricaoID, inscricao.ID)
		}
	})

	t.Run("GetByID non-existing inscricao", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return nil, nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricao, err := svc.GetByID(ctx, uuid.New())
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}
		if inscricao != nil {
			t.Error("Expected nil for non-existing inscricao")
		}
	})
}

// TestInscricaoService_GetByCursoID tests the GetByCursoID method
func TestInscricaoService_GetByCursoID(t *testing.T) {
	ctx := context.Background()

	t.Run("GetByCursoID with results", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			GetByCursoIDFunc: func(ctx context.Context, cursoID int, filter map[string]interface{}, limit, offset int) ([]*models.Inscricao, int, error) {
				return []*models.Inscricao{
					{ID: uuid.New(), CursoID: 1},
					{ID: uuid.New(), CursoID: 1},
				}, 2, nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricoes, total, err := svc.GetByCursoID(ctx, 1, nil, 1, 10)
		if err != nil {
			t.Fatalf("GetByCursoID failed: %v", err)
		}
		if total != 2 {
			t.Errorf("Expected total 2, got %d", total)
		}
		if len(inscricoes) != 2 {
			t.Errorf("Expected 2 inscricoes, got %d", len(inscricoes))
		}
	})

	t.Run("GetByCursoID with pagination", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			GetByCursoIDFunc: func(ctx context.Context, cursoID int, filter map[string]interface{}, limit, offset int) ([]*models.Inscricao, int, error) {
				if limit != 5 || offset != 10 {
					t.Errorf("Expected limit=5 offset=10, got limit=%d offset=%d", limit, offset)
				}
				return []*models.Inscricao{}, 0, nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		_, _, err := svc.GetByCursoID(ctx, 1, nil, 3, 5)
		if err != nil {
			t.Fatalf("GetByCursoID failed: %v", err)
		}
	})
}

// TestInscricaoService_ListByCPF tests the ListByCPF method
func TestInscricaoService_ListByCPF(t *testing.T) {
	ctx := context.Background()

	t.Run("ListByCPF with results", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			ListByCPFFunc: func(ctx context.Context, cpf string, filter map[string]interface{}, offset, limit int) ([]*models.Inscricao, int, error) {
				if cpf != "12345678900" {
					t.Errorf("Expected CPF 12345678900, got %s", cpf)
				}
				return []*models.Inscricao{
					{ID: uuid.New(), CPF: "12345678900"},
				}, 1, nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricoes, total, err := svc.ListByCPF(ctx, "12345678900", nil, 0, 10)
		if err != nil {
			t.Fatalf("ListByCPF failed: %v", err)
		}
		if total != 1 {
			t.Errorf("Expected total 1, got %d", total)
		}
		if len(inscricoes) != 1 {
			t.Errorf("Expected 1 inscricao, got %d", len(inscricoes))
		}
	})

	t.Run("ListByCPF with repository error", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			ListByCPFFunc: func(ctx context.Context, cpf string, filter map[string]interface{}, offset, limit int) ([]*models.Inscricao, int, error) {
				return nil, 0, errors.New("database error")
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		_, _, err := svc.ListByCPF(ctx, "12345678900", nil, 0, 10)
		if err == nil {
			t.Error("Expected database error")
		}
	})
}

// TestInscricaoService_UpdateInscricao tests the UpdateInscricao method
func TestInscricaoService_UpdateInscricao(t *testing.T) {
	ctx := context.Background()

	t.Run("UpdateInscricao success", func(t *testing.T) {
		inscricaoID := uuid.New()
		newName := "Updated Name"
		newEmail := "updated@example.com"

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CursoID: 1,
					Name:    "Original Name",
					Email:   "original@example.com",
				}, nil
			},
			UpdateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				if inscricao.Name != newName {
					t.Errorf("Expected name %s, got %s", newName, inscricao.Name)
				}
				if inscricao.Email != newEmail {
					t.Errorf("Expected email %s, got %s", newEmail, inscricao.Email)
				}
				return nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		updateData := &models.InscricaoUpdateRequest{
			Name:  &newName,
			Email: &newEmail,
		}

		err := svc.UpdateInscricao(ctx, inscricaoID, 1, updateData)
		if err != nil {
			t.Fatalf("UpdateInscricao failed: %v", err)
		}
	})

	t.Run("UpdateInscricao fails when inscricao not found", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return nil, nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		updateData := &models.InscricaoUpdateRequest{}

		err := svc.UpdateInscricao(ctx, inscricaoID, 1, updateData)
		if err == nil {
			t.Error("Expected error when inscricao not found")
		}
	})

	t.Run("UpdateInscricao fails when curso ID mismatch", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CursoID: 2, // Different curso
				}, nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		updateData := &models.InscricaoUpdateRequest{}

		err := svc.UpdateInscricao(ctx, inscricaoID, 1, updateData)
		if err == nil {
			t.Error("Expected error when curso ID mismatch")
		}
	})
}

// TestInscricaoService_CreateByAdmin tests the CreateByAdmin method used by CSV import jobs.
// It must bypass course status validation while preserving all other validations.
func TestInscricaoService_CreateByAdmin(t *testing.T) {
	ctx := context.Background()

	t.Run("CreateByAdmin succeeds for closed course", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, nil
			},
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				return nil
			},
		}
		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (string, *time.Time, *time.Time, bool, error) {
				return string(models.StatusCursoClosed), nil, nil, true, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.CreateByAdmin(ctx, &models.Inscricao{CPF: "12345678900", Name: "Teste", CursoID: 1})
		if err != nil {
			t.Errorf("CreateByAdmin should succeed for closed course, got: %v", err)
		}
	})

	t.Run("CreateByAdmin succeeds for draft course", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, nil
			},
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				return nil
			},
		}
		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (string, *time.Time, *time.Time, bool, error) {
				return string(models.StatusCursoDraft), nil, nil, false, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.CreateByAdmin(ctx, &models.Inscricao{CPF: "12345678900", Name: "Teste", CursoID: 1})
		if err != nil {
			t.Errorf("CreateByAdmin should succeed for draft course, got: %v", err)
		}
	})

	t.Run("CreateByAdmin fails when CPF already enrolled", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return true, nil
			},
		}
		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (string, *time.Time, *time.Time, bool, error) {
				return string(models.StatusCursoClosed), nil, nil, false, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.CreateByAdmin(ctx, &models.Inscricao{CPF: "12345678900", Name: "Teste", CursoID: 1})
		if err == nil {
			t.Error("Expected error for duplicate CPF")
		}
	})

	t.Run("CreateByAdmin sets status approved when auto_approve is true", func(t *testing.T) {
		var created *models.Inscricao
		inscricaoRepo := &MockInscricaoRepository{
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, nil
			},
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				created = inscricao
				return nil
			},
		}
		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (string, *time.Time, *time.Time, bool, error) {
				return string(models.StatusCursoClosed), nil, nil, true, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.CreateByAdmin(ctx, &models.Inscricao{CPF: "12345678900", Name: "Teste", CursoID: 1})
		if err != nil {
			t.Fatalf("CreateByAdmin failed: %v", err)
		}
		if created.Status != models.StatusInscricaoApproved {
			t.Errorf("Expected status approved, got %s", created.Status)
		}
	})

	t.Run("CreateByAdmin sets status pending when auto_approve is false", func(t *testing.T) {
		var created *models.Inscricao
		inscricaoRepo := &MockInscricaoRepository{
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, nil
			},
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				created = inscricao
				return nil
			},
		}
		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (string, *time.Time, *time.Time, bool, error) {
				return string(models.StatusCursoClosed), nil, nil, false, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.CreateByAdmin(ctx, &models.Inscricao{CPF: "12345678900", Name: "Teste", CursoID: 1})
		if err != nil {
			t.Fatalf("CreateByAdmin failed: %v", err)
		}
		if created.Status != models.StatusInscricaoPending {
			t.Errorf("Expected status pending, got %s", created.Status)
		}
	})

	t.Run("CreateByAdmin preserves secretaria-provided contact over RMI snapshot", func(t *testing.T) {
		var created *models.Inscricao
		inscricaoRepo := &MockInscricaoRepository{
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, nil
			},
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				created = inscricao
				return nil
			},
		}
		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (string, *time.Time, *time.Time, bool, error) {
				return string(models.StatusCursoClosed), nil, nil, false, nil
			},
		}
		// RMI returns divergent contact data; it must not overwrite the órgão-provided contact.
		citizenFetcher := &MockCitizenDataFetcher{
			SyncFunc: func(ctx context.Context, cpf string) (*models.CitizenSnapshot, error) {
				return &models.CitizenSnapshot{
					CPF:     cpf,
					Email:   "rmi@rmi.com",
					Celular: "21999999999",
					Nome:    "Nome RMI",
				}, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, citizenFetcher, nil, &config.AppConfig{})

		err := svc.CreateByAdmin(ctx, &models.Inscricao{
			CPF:     "12345678900",
			CursoID: 1,
			Name:    "Nome Secretaria",
			Email:   "secretaria@example.com",
			Phone:   "21988887777",
		})
		if err != nil {
			t.Fatalf("CreateByAdmin failed: %v", err)
		}
		if created.Email != "secretaria@example.com" {
			t.Errorf("Expected secretaria email preserved, got %s", created.Email)
		}
		if created.Phone != "21988887777" {
			t.Errorf("Expected secretaria phone preserved, got %s", created.Phone)
		}
		// A name provided by the secretaria must not be replaced by the RMI name.
		if created.Name != "Nome Secretaria" {
			t.Errorf("Expected provided name preserved, got %s", created.Name)
		}
	})

	t.Run("CreateByAdmin fills missing name from RMI snapshot", func(t *testing.T) {
		var created *models.Inscricao
		inscricaoRepo := &MockInscricaoRepository{
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, nil
			},
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				created = inscricao
				return nil
			},
		}
		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (string, *time.Time, *time.Time, bool, error) {
				return string(models.StatusCursoClosed), nil, nil, false, nil
			},
		}
		citizenFetcher := &MockCitizenDataFetcher{
			SyncFunc: func(ctx context.Context, cpf string) (*models.CitizenSnapshot, error) {
				return &models.CitizenSnapshot{CPF: cpf, Nome: "Nome RMI"}, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, citizenFetcher, nil, &config.AppConfig{})

		err := svc.CreateByAdmin(ctx, &models.Inscricao{CPF: "12345678900", CursoID: 1, Name: ""})
		if err != nil {
			t.Fatalf("CreateByAdmin failed: %v", err)
		}
		if created.Name != "Nome RMI" {
			t.Errorf("Expected name filled from RMI snapshot, got %s", created.Name)
		}
	})

	t.Run("Create (public) fails for closed course", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{}
		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (string, *time.Time, *time.Time, bool, error) {
				return string(models.StatusCursoClosed), nil, nil, false, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.Create(ctx, &models.Inscricao{CPF: "12345678900", CursoID: 1})
		if err == nil {
			t.Error("Expected Create to fail for closed course")
		}
		if err.Error() != "curso não está aberto para inscrições" {
			t.Errorf("Unexpected error message: %s", err.Error())
		}
	})

	t.Run("CreateByAdmin fails when ValidateForEnrollment errors", func(t *testing.T) {
		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (string, *time.Time, *time.Time, bool, error) {
				return "", nil, nil, false, fmt.Errorf("db error")
			},
		}
		svc := services.NewInscricaoServiceWithInterface(&MockInscricaoRepository{}, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.CreateByAdmin(ctx, &models.Inscricao{CPF: "12345678900", CursoID: 1})
		if err == nil || err.Error() != "db error" {
			t.Errorf("Expected db error, got: %v", err)
		}
	})

	t.Run("CreateByAdmin bypasses enrollment period not started", func(t *testing.T) {
		future := time.Now().Add(24 * time.Hour)
		inscricaoRepo := &MockInscricaoRepository{
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, nil
			},
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				return nil
			},
		}
		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (string, *time.Time, *time.Time, bool, error) {
				return string(models.StatusCursoClosed), &future, nil, false, nil
			},
		}
		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.CreateByAdmin(ctx, &models.Inscricao{CPF: "12345678900", CursoID: 1})
		if err != nil {
			t.Errorf("Expected no error when enrollment period not started (admin bypass), got: %v", err)
		}
	})

	t.Run("CreateByAdmin bypasses enrollment period ended", func(t *testing.T) {
		past := time.Now().Add(-24 * time.Hour)
		inscricaoRepo := &MockInscricaoRepository{
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, nil
			},
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				return nil
			},
		}
		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (string, *time.Time, *time.Time, bool, error) {
				return string(models.StatusCursoClosed), nil, &past, false, nil
			},
		}
		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.CreateByAdmin(ctx, &models.Inscricao{CPF: "12345678900", CursoID: 1})
		if err != nil {
			t.Errorf("Expected no error when enrollment period ended (admin bypass), got: %v", err)
		}
	})

	t.Run("CreateByAdmin fails when ExistsByCPFAndCurso errors", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, fmt.Errorf("db error checking cpf")
			},
		}
		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (string, *time.Time, *time.Time, bool, error) {
				return string(models.StatusCursoClosed), nil, nil, false, nil
			},
		}
		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.CreateByAdmin(ctx, &models.Inscricao{CPF: "12345678900", CursoID: 1})
		if err == nil {
			t.Error("Expected error from repo")
		}
	})

	t.Run("CreateByAdmin fails when repo.Create errors", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, nil
			},
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				return fmt.Errorf("insert failed")
			},
		}
		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (string, *time.Time, *time.Time, bool, error) {
				return string(models.StatusCursoClosed), nil, nil, false, nil
			},
		}
		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.CreateByAdmin(ctx, &models.Inscricao{CPF: "12345678900", Name: "Teste", CursoID: 1})
		if err == nil || err.Error() != "insert failed" {
			t.Errorf("Expected insert failed error, got: %v", err)
		}
	})

	t.Run("CreateByAdmin does not backfill contact from citizen fetcher", func(t *testing.T) {
		// When the secretaria provided no contact, the enrollment keeps it empty (rendered as
		// "Não informado" in the Admin). It is NOT backfilled from RMI here — course-email
		// dispatch already falls back to the RMI snapshot when the enrollment has no email.
		var created *models.Inscricao
		inscricaoRepo := &MockInscricaoRepository{
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, nil
			},
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				created = inscricao
				return nil
			},
		}
		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (string, *time.Time, *time.Time, bool, error) {
				return string(models.StatusCursoClosed), nil, nil, false, nil
			},
		}
		citizenFetcher := &MockCitizenDataFetcher{
			SyncFunc: func(ctx context.Context, cpf string) (*models.CitizenSnapshot, error) {
				return &models.CitizenSnapshot{
					CPF:     cpf,
					Email:   "fetched@example.com",
					Celular: "21999990000",
					Nome:    "Nome do RMI",
				}, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, citizenFetcher, nil, &config.AppConfig{})

		err := svc.CreateByAdmin(ctx, &models.Inscricao{CPF: "12345678900", Name: "Original", CursoID: 1})
		if err != nil {
			t.Fatalf("CreateByAdmin failed: %v", err)
		}
		if created.Email != "" {
			t.Errorf("Expected empty email (no backfill from RMI), got %s", created.Email)
		}
		if created.Phone != "" {
			t.Errorf("Expected empty phone (no backfill from RMI), got %s", created.Phone)
		}
	})

	t.Run("CreateByAdmin continues when citizen fetcher errors", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, nil
			},
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				return nil
			},
		}
		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (string, *time.Time, *time.Time, bool, error) {
				return string(models.StatusCursoClosed), nil, nil, false, nil
			},
		}
		citizenFetcher := &MockCitizenDataFetcher{
			SyncFunc: func(ctx context.Context, cpf string) (*models.CitizenSnapshot, error) {
				return nil, fmt.Errorf("rmi unavailable")
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, citizenFetcher, nil, &config.AppConfig{})

		err := svc.CreateByAdmin(ctx, &models.Inscricao{CPF: "12345678900", Name: "Teste", CursoID: 1})
		if err != nil {
			t.Errorf("CreateByAdmin should not fail when citizen fetcher errors, got: %v", err)
		}
	})

	t.Run("CreateByAdmin sends approved email when auto_approve is true", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, nil
			},
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error { return nil },
		}
		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (string, *time.Time, *time.Time, bool, error) {
				return string(models.StatusCursoClosed), nil, nil, true, nil
			},
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return &models.Curso{ID: id, Titulo: "Curso Fechado"}, nil
			},
		}
		emailNotifier := newMockEmailNotifier()
		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, emailNotifier, &config.AppConfig{})

		err := svc.CreateByAdmin(ctx, &models.Inscricao{CPF: "12345678900", Name: "Teste", CursoID: 1})
		if err != nil {
			t.Fatalf("CreateByAdmin failed: %v", err)
		}
		if !emailNotifier.waitForCall("enrollment.approved", 500*time.Millisecond) {
			t.Error("Expected enrollment.approved email to be sent")
		}
	})

	t.Run("CreateByAdmin sends pending email when auto_approve is false", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, nil
			},
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error { return nil },
		}
		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (string, *time.Time, *time.Time, bool, error) {
				return string(models.StatusCursoClosed), nil, nil, false, nil
			},
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return &models.Curso{ID: id, Titulo: "Curso Fechado"}, nil
			},
		}
		emailNotifier := newMockEmailNotifier()
		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, emailNotifier, &config.AppConfig{})

		err := svc.CreateByAdmin(ctx, &models.Inscricao{CPF: "12345678900", Name: "Teste", CursoID: 1})
		if err != nil {
			t.Fatalf("CreateByAdmin failed: %v", err)
		}
		if !emailNotifier.waitForCall("enrollment.created", 500*time.Millisecond) {
			t.Error("Expected enrollment.created email to be sent")
		}
	})

	t.Run("CreateByAdmin fails when GetByID errors on schedule validation", func(t *testing.T) {
		scheduleID := uuid.New()
		inscricaoRepo := &MockInscricaoRepository{
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, nil
			},
		}
		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (string, *time.Time, *time.Time, bool, error) {
				return string(models.StatusCursoClosed), nil, nil, false, nil
			},
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return nil, fmt.Errorf("db error on GetByID")
			},
		}
		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.CreateByAdmin(ctx, &models.Inscricao{CPF: "12345678900", Name: "Teste", CursoID: 1, ScheduleID: &scheduleID})
		if err == nil {
			t.Error("Expected error when GetByID fails")
		}
	})

	t.Run("CreateByAdmin fails when curso not found on schedule validation", func(t *testing.T) {
		scheduleID := uuid.New()
		inscricaoRepo := &MockInscricaoRepository{
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, nil
			},
		}
		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (string, *time.Time, *time.Time, bool, error) {
				return string(models.StatusCursoClosed), nil, nil, false, nil
			},
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return nil, nil
			},
		}
		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.CreateByAdmin(ctx, &models.Inscricao{CPF: "12345678900", Name: "Teste", CursoID: 1, ScheduleID: &scheduleID})
		if err == nil || err.Error() != "curso não encontrado" {
			t.Errorf("Expected 'curso não encontrado', got: %v", err)
		}
	})
}

// TestInscricaoService_CreateManual_Extended covers additional paths in CreateManual
func TestInscricaoService_CreateManual_Extended(t *testing.T) {
	ctx := context.Background()

	t.Run("CreateManual fails when ValidateForEnrollment errors", func(t *testing.T) {
		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (string, *time.Time, *time.Time, bool, error) {
				return "", nil, nil, false, fmt.Errorf("db error")
			},
		}
		svc := services.NewInscricaoServiceWithInterface(&MockInscricaoRepository{}, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.CreateManual(ctx, &models.Inscricao{CPF: "12345678900", CursoID: 1})
		if err == nil || err.Error() != "db error" {
			t.Errorf("Expected db error, got: %v", err)
		}
	})

	t.Run("CreateManual fails when CPF already enrolled", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return true, nil
			},
		}
		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (string, *time.Time, *time.Time, bool, error) {
				return string(models.StatusCursoClosed), nil, nil, false, nil
			},
		}
		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.CreateManual(ctx, &models.Inscricao{CPF: "12345678900", CursoID: 1})
		if err == nil {
			t.Error("Expected error for duplicate CPF")
		}
	})

	t.Run("CreateManual fails when repo.Create errors", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, nil
			},
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				return fmt.Errorf("insert failed")
			},
		}
		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (string, *time.Time, *time.Time, bool, error) {
				return string(models.StatusCursoClosed), nil, nil, false, nil
			},
		}
		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.CreateManual(ctx, &models.Inscricao{CPF: "12345678900", Name: "Teste", CursoID: 1})
		if err == nil || err.Error() != "insert failed" {
			t.Errorf("Expected insert failed error, got: %v", err)
		}
	})

	t.Run("CreateManual sends approved email when auto_approve is true", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, nil
			},
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error { return nil },
		}
		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (string, *time.Time, *time.Time, bool, error) {
				return string(models.StatusCursoClosed), nil, nil, true, nil
			},
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return &models.Curso{ID: id, Titulo: "Curso Manual"}, nil
			},
		}
		emailNotifier := newMockEmailNotifier()
		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, emailNotifier, &config.AppConfig{})

		err := svc.CreateManual(ctx, &models.Inscricao{CPF: "12345678900", Name: "Teste", CursoID: 1})
		if err != nil {
			t.Fatalf("CreateManual failed: %v", err)
		}
		if !emailNotifier.waitForCall("enrollment.approved", 500*time.Millisecond) {
			t.Error("Expected enrollment.approved email to be sent")
		}
	})

	t.Run("CreateManual sends pending email when auto_approve is false", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, nil
			},
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error { return nil },
		}
		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (string, *time.Time, *time.Time, bool, error) {
				return string(models.StatusCursoClosed), nil, nil, false, nil
			},
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return &models.Curso{ID: id, Titulo: "Curso Manual"}, nil
			},
		}
		emailNotifier := newMockEmailNotifier()
		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, emailNotifier, &config.AppConfig{})

		err := svc.CreateManual(ctx, &models.Inscricao{CPF: "12345678900", Name: "Teste", CursoID: 1})
		if err != nil {
			t.Fatalf("CreateManual failed: %v", err)
		}
		if !emailNotifier.waitForCall("enrollment.created", 500*time.Millisecond) {
			t.Error("Expected enrollment.created email to be sent")
		}
	})

	t.Run("CreateManual fails when GetByID errors on schedule validation", func(t *testing.T) {
		scheduleID := uuid.New()
		inscricaoRepo := &MockInscricaoRepository{
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, nil
			},
		}
		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (string, *time.Time, *time.Time, bool, error) {
				return string(models.StatusCursoClosed), nil, nil, false, nil
			},
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return nil, fmt.Errorf("db error on GetByID")
			},
		}
		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.CreateManual(ctx, &models.Inscricao{CPF: "12345678900", Name: "Teste", CursoID: 1, ScheduleID: &scheduleID})
		if err == nil {
			t.Error("Expected error when GetByID fails")
		}
	})

	t.Run("CreateManual fails when curso not found on schedule validation", func(t *testing.T) {
		scheduleID := uuid.New()
		inscricaoRepo := &MockInscricaoRepository{
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, nil
			},
		}
		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (string, *time.Time, *time.Time, bool, error) {
				return string(models.StatusCursoClosed), nil, nil, false, nil
			},
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return nil, nil
			},
		}
		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.CreateManual(ctx, &models.Inscricao{CPF: "12345678900", Name: "Teste", CursoID: 1, ScheduleID: &scheduleID})
		if err == nil || err.Error() != "curso não encontrado" {
			t.Errorf("Expected 'curso não encontrado', got: %v", err)
		}
	})
}

// TestInscricaoService_CreateManual tests the CreateManual method
func TestInscricaoService_CreateManual(t *testing.T) {
	ctx := context.Background()

	t.Run("CreateManual bypasses enrollment period", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				if inscricao.Status != models.StatusInscricaoApproved {
					t.Errorf("Expected status APPROVED, got %s", inscricao.Status)
				}
				if inscricao.Email != "admin@example.com" {
					t.Errorf("Expected admin-provided email, got %s", inscricao.Email)
				}
				return nil
			},
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, nil
			},
		}

		// Simulate expired enrollment period
		pastDate := time.Now().Add(-24 * time.Hour)
		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (status string, enrollmentStart *time.Time, enrollmentEnd *time.Time, autoApprove bool, err error) {
				return string(models.StatusCursoOpened), nil, &pastDate, true, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricao := &models.Inscricao{
			CPF:     "12345678900",
			CursoID: 1,
			Name:    "Admin Created User",
			Email:   "admin@example.com",
		}

		err := svc.CreateManual(ctx, inscricao)
		if err != nil {
			t.Fatalf("CreateManual failed: %v", err)
		}
	})

	t.Run("CreateManual does not fetch citizen data", func(t *testing.T) {
		citizenDataFetcherCalled := false

		inscricaoRepo := &MockInscricaoRepository{
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				// Should preserve admin-provided data
				if inscricao.Email != "admin@example.com" {
					t.Errorf("Expected admin email to be preserved, got %s", inscricao.Email)
				}
				return nil
			},
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, nil
			},
		}

		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (status string, enrollmentStart *time.Time, enrollmentEnd *time.Time, autoApprove bool, err error) {
				return string(models.StatusCursoOpened), nil, nil, false, nil
			},
		}

		citizenFetcher := &MockCitizenDataFetcher{
			SyncFunc: func(ctx context.Context, cpf string) (*models.CitizenSnapshot, error) {
				citizenDataFetcherCalled = true
				return &models.CitizenSnapshot{
					Email: "should-not-override@example.com",
				}, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, citizenFetcher, nil, &config.AppConfig{})

		inscricao := &models.Inscricao{
			CPF:     "12345678900",
			CursoID: 1,
			Name:    "Manual User",
			Email:   "admin@example.com",
		}

		err := svc.CreateManual(ctx, inscricao)
		if err != nil {
			t.Fatalf("CreateManual failed: %v", err)
		}

		if citizenDataFetcherCalled {
			t.Error("Citizen data fetcher should not be called in CreateManual")
		}
	})

	t.Run("CreateManual succeeds even when curso not opened", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, nil
			},
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				return nil
			},
		}

		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (status string, enrollmentStart *time.Time, enrollmentEnd *time.Time, autoApprove bool, err error) {
				return string(models.StatusCursoDraft), nil, nil, false, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricao := &models.Inscricao{
			CPF:     "12345678900",
			Name:    "Teste Admin",
			CursoID: 1,
		}

		err := svc.CreateManual(ctx, inscricao)
		if err != nil {
			t.Errorf("CreateManual should succeed for closed course, got: %v", err)
		}
	})

	t.Run("CreateManual with schedule validation", func(t *testing.T) {
		scheduleID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				return nil
			},
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, nil
			},
		}

		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (status string, enrollmentStart *time.Time, enrollmentEnd *time.Time, autoApprove bool, err error) {
				return string(models.StatusCursoOpened), nil, nil, true, nil
			},
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				acceptingEnrollments := true
				return &models.Curso{
					ID: 1,
					LocationClasses: []models.LocationClass{
						{
							Schedules: []models.CourseSchedule{
								{
									ID:                   scheduleID,
									Vacancies:            10,
									AcceptingEnrollments: &acceptingEnrollments,
								},
							},
						},
					},
				}, nil
			},
			CountEnrollmentsByScheduleIDFunc: func(ctx context.Context, sid uuid.UUID) (int64, error) {
				return 5, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricao := &models.Inscricao{
			CPF:        "12345678900",
			CursoID:    1,
			Name:       "Test User",
			Email:      "test@example.com",
			ScheduleID: &scheduleID,
		}

		err := svc.CreateManual(ctx, inscricao)
		if err != nil {
			t.Fatalf("CreateManual failed: %v", err)
		}
	})

	t.Run("CreateManual with full schedule disables auto-approve", func(t *testing.T) {
		scheduleID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				// Should be pending since schedule is full
				if inscricao.Status != models.StatusInscricaoPending {
					t.Errorf("Expected status PENDING due to full schedule, got %s", inscricao.Status)
				}
				return nil
			},
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, nil
			},
		}

		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (status string, enrollmentStart *time.Time, enrollmentEnd *time.Time, autoApprove bool, err error) {
				return string(models.StatusCursoOpened), nil, nil, true, nil
			},
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				acceptingEnrollments := true
				return &models.Curso{
					ID: 1,
					LocationClasses: []models.LocationClass{
						{
							Schedules: []models.CourseSchedule{
								{
									ID:                   scheduleID,
									Vacancies:            10,
									AcceptingEnrollments: &acceptingEnrollments,
								},
							},
						},
					},
				}, nil
			},
			CountEnrollmentsByScheduleIDFunc: func(ctx context.Context, sid uuid.UUID) (int64, error) {
				return 10, nil // Full schedule
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricao := &models.Inscricao{
			CPF:        "12345678900",
			CursoID:    1,
			Name:       "Test User",
			Email:      "test@example.com",
			ScheduleID: &scheduleID,
		}

		err := svc.CreateManual(ctx, inscricao)
		if err != nil {
			t.Fatalf("CreateManual failed: %v", err)
		}
	})

	t.Run("CreateManual fails with invalid schedule", func(t *testing.T) {
		invalidScheduleID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, nil
			},
		}

		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (status string, enrollmentStart *time.Time, enrollmentEnd *time.Time, autoApprove bool, err error) {
				return string(models.StatusCursoOpened), nil, nil, false, nil
			},
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return &models.Curso{
					ID:              1,
					LocationClasses: []models.LocationClass{},
				}, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricao := &models.Inscricao{
			CPF:        "12345678900",
			CursoID:    1,
			ScheduleID: &invalidScheduleID,
		}

		err := svc.CreateManual(ctx, inscricao)
		if err == nil {
			t.Error("Expected error with invalid schedule ID")
		}
	})

	t.Run("CreateManual validates email with SanitizeEmail", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				// Valid email should be preserved
				if inscricao.Email != "test@example.com" {
					t.Errorf("Expected valid email 'test@example.com', got %s", inscricao.Email)
				}
				return nil
			},
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, nil
			},
		}

		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (status string, enrollmentStart *time.Time, enrollmentEnd *time.Time, autoApprove bool, err error) {
				return string(models.StatusCursoOpened), nil, nil, false, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricao := &models.Inscricao{
			CPF:     "12345678900",
			CursoID: 1,
			Name:    "Test User",
			Email:   "test@example.com", // Valid email
		}

		err := svc.CreateManual(ctx, inscricao)
		if err != nil {
			t.Fatalf("CreateManual failed: %v", err)
		}
	})

	t.Run("CreateManual handles invalid email", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				// Invalid email should be sanitized to empty string
				if inscricao.Email != "" {
					t.Errorf("Expected invalid email to be sanitized to empty string, got %s", inscricao.Email)
				}
				return nil
			},
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, nil
			},
		}

		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (status string, enrollmentStart *time.Time, enrollmentEnd *time.Time, autoApprove bool, err error) {
				return string(models.StatusCursoOpened), nil, nil, false, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricao := &models.Inscricao{
			CPF:     "12345678900",
			CursoID: 1,
			Name:    "Test User",
			Email:   "not-a-valid-email", // Invalid email format
		}

		err := svc.CreateManual(ctx, inscricao)
		if err != nil {
			t.Fatalf("CreateManual failed: %v", err)
		}
	})
}

// MockCitizenSnapshotRepository for testing - implements CitizenSnapshotRepositoryInterface
type MockCitizenSnapshotRepository struct {
	GetByCPFFunc               func(ctx context.Context, cpf string) (*models.CitizenSnapshot, error)
	GetByCPFsFunc              func(ctx context.Context, cpfs []string) (map[string]*models.CitizenSnapshot, error)
	UpsertFunc                 func(ctx context.Context, snapshot *models.CitizenSnapshot) error
	BatchUpsertFunc            func(ctx context.Context, snapshots []*models.CitizenSnapshot) error
	GetStaleSnapshotsFunc      func(ctx context.Context, staleThreshold time.Duration, limit int) ([]*models.CitizenSnapshot, error)
	GetCPFsWithEnrollmentsFunc func(ctx context.Context, staleThreshold time.Duration, limit int) ([]string, error)
	DeleteFunc                 func(ctx context.Context, cpf string) error
	CountFunc                  func(ctx context.Context) (int64, error)
}

func (m *MockCitizenSnapshotRepository) GetByCPF(ctx context.Context, cpf string) (*models.CitizenSnapshot, error) {
	if m.GetByCPFFunc != nil {
		return m.GetByCPFFunc(ctx, cpf)
	}
	return nil, nil
}

func (m *MockCitizenSnapshotRepository) GetByCPFs(ctx context.Context, cpfs []string) (map[string]*models.CitizenSnapshot, error) {
	if m.GetByCPFsFunc != nil {
		return m.GetByCPFsFunc(ctx, cpfs)
	}
	return make(map[string]*models.CitizenSnapshot), nil
}

func (m *MockCitizenSnapshotRepository) Upsert(ctx context.Context, snapshot *models.CitizenSnapshot) error {
	if m.UpsertFunc != nil {
		return m.UpsertFunc(ctx, snapshot)
	}
	return nil
}

func (m *MockCitizenSnapshotRepository) BatchUpsert(ctx context.Context, snapshots []*models.CitizenSnapshot) error {
	if m.BatchUpsertFunc != nil {
		return m.BatchUpsertFunc(ctx, snapshots)
	}
	return nil
}

func (m *MockCitizenSnapshotRepository) GetStaleSnapshots(ctx context.Context, staleThreshold time.Duration, limit int) ([]*models.CitizenSnapshot, error) {
	if m.GetStaleSnapshotsFunc != nil {
		return m.GetStaleSnapshotsFunc(ctx, staleThreshold, limit)
	}
	return []*models.CitizenSnapshot{}, nil
}

func (m *MockCitizenSnapshotRepository) GetCPFsWithEnrollments(ctx context.Context, staleThreshold time.Duration, limit int) ([]string, error) {
	if m.GetCPFsWithEnrollmentsFunc != nil {
		return m.GetCPFsWithEnrollmentsFunc(ctx, staleThreshold, limit)
	}
	return []string{}, nil
}

func (m *MockCitizenSnapshotRepository) Delete(ctx context.Context, cpf string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, cpf)
	}
	return nil
}

func (m *MockCitizenSnapshotRepository) Count(ctx context.Context) (int64, error) {
	if m.CountFunc != nil {
		return m.CountFunc(ctx)
	}
	return 0, nil
}

// TestInscricaoService_EnrichWithPersonalInfo tests the EnrichWithPersonalInfo method
func TestInscricaoService_EnrichWithPersonalInfo(t *testing.T) {
	ctx := context.Background()

	t.Run("EnrichWithPersonalInfo returns when citizenSnapshotRepo is nil", func(t *testing.T) {
		svc := services.NewInscricaoServiceWithInterface(
			&MockInscricaoRepository{},
			&MockCursoRepository{},
			nil, // nil snapshot repo
			nil,
			nil,
			&config.AppConfig{},
		)

		inscricao := &models.Inscricao{
			CPF: "12345678900",
		}

		// Should not panic with nil repo
		svc.EnrichWithPersonalInfo(ctx, inscricao)

		if inscricao.PersonalInfo != nil {
			t.Error("Expected PersonalInfo to remain nil when repo is nil")
		}
	})

	t.Run("EnrichWithPersonalInfo handles nil inscricao", func(t *testing.T) {
		svc := services.NewInscricaoServiceWithInterface(
			&MockInscricaoRepository{},
			&MockCursoRepository{},
			nil,
			nil,
			nil,
			&config.AppConfig{},
		)

		// Should not panic with nil inscricao
		svc.EnrichWithPersonalInfo(ctx, nil)
	})

	t.Run("EnrichWithPersonalInfo handles empty CPF", func(t *testing.T) {
		svc := services.NewInscricaoServiceWithInterface(
			&MockInscricaoRepository{},
			&MockCursoRepository{},
			nil,
			nil,
			nil,
			&config.AppConfig{},
		)

		inscricao := &models.Inscricao{
			CPF: "",
		}

		svc.EnrichWithPersonalInfo(ctx, inscricao)

		if inscricao.PersonalInfo != nil {
			t.Error("Expected PersonalInfo to remain nil for empty CPF")
		}
	})
}

// TestInscricaoService_EnrichMultipleWithPersonalInfo tests the EnrichMultipleWithPersonalInfo method
func TestInscricaoService_EnrichMultipleWithPersonalInfo(t *testing.T) {
	ctx := context.Background()

	t.Run("EnrichMultipleWithPersonalInfo handles empty list", func(t *testing.T) {
		svc := services.NewInscricaoServiceWithInterface(
			&MockInscricaoRepository{},
			&MockCursoRepository{},
			nil,
			nil,
			nil,
			&config.AppConfig{},
		)

		inscricoes := []*models.Inscricao{}

		// Should not panic with empty list
		svc.EnrichMultipleWithPersonalInfo(ctx, inscricoes)
	})

	t.Run("EnrichMultipleWithPersonalInfo returns when repo is nil", func(t *testing.T) {
		svc := services.NewInscricaoServiceWithInterface(
			&MockInscricaoRepository{},
			&MockCursoRepository{},
			nil,
			nil,
			nil,
			&config.AppConfig{},
		)

		inscricoes := []*models.Inscricao{
			{CPF: "11111111111"},
		}

		// Should return early when repo is nil
		svc.EnrichMultipleWithPersonalInfo(ctx, inscricoes)

		if inscricoes[0].PersonalInfo != nil {
			t.Error("Expected PersonalInfo to remain nil when repo is nil")
		}
	})
}

// TestInscricaoService_ChangeSchedule tests the ChangeSchedule method
func TestInscricaoService_ChangeSchedule(t *testing.T) {
	ctx := context.Background()

	t.Run("ChangeSchedule success", func(t *testing.T) {
		inscricaoID := uuid.New()
		newScheduleID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CPF:     "12345678900",
					CursoID: 1,
					Status:  models.StatusInscricaoApproved,
				}, nil
			},
			UpdateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				if inscricao.ScheduleID == nil || *inscricao.ScheduleID != newScheduleID {
					t.Error("Expected schedule ID to be updated")
				}
				return nil
			},
		}

		cursoRepo := &MockCursoRepository{
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return &models.Curso{
					ID: id,
					LocationClasses: []models.LocationClass{
						{Schedules: []models.CourseSchedule{{}, {}}},
					},
				}, nil
			},
			CountEnrollmentsByScheduleIDsFunc: func(ctx context.Context, scheduleIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
				return map[uuid.UUID]int64{
					newScheduleID: 5, // 5 enrolled, has space
				}, nil
			},
			GetCourseScheduleByIDFunc: func(ctx context.Context, scheduleID uuid.UUID) (*models.CourseSchedule, error) {
				acceptingEnrollments := true
				return &models.CourseSchedule{
					ID:                   scheduleID,
					Vacancies:            10,
					AcceptingEnrollments: &acceptingEnrollments,
				}, nil
			},
		}

		futureDate := time.Now().Add(72 * time.Hour) // 3 days from now
		request := &models.ScheduleChangeRequest{
			ScheduleID: &newScheduleID,
			EnrolledUnit: &models.EnrolledUnit{
				ID: "unit-1",
				Schedules: []models.EnrolledUnitSchedule{
					{
						ID:             newScheduleID.String(),
						ClassStartDate: futureDate.Format(time.RFC3339),
					},
				},
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{
			Enrollment: config.EnrollmentSettings{
				ScheduleChangeDeadlineHours: 48,
			},
		})

		result, err := svc.ChangeSchedule(ctx, inscricaoID, "12345678900", request)
		if err != nil {
			t.Fatalf("ChangeSchedule failed: %v", err)
		}
		if result == nil {
			t.Fatal("Expected inscricao result")
		}
	})

	t.Run("ChangeSchedule fails when user CPF doesn't match", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CPF:     "12345678900",
					CursoID: 1,
					Status:  models.StatusInscricaoApproved,
				}, nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		futureDate := time.Now().Add(72 * time.Hour)
		request := &models.ScheduleChangeRequest{
			EnrolledUnit: &models.EnrolledUnit{
				Schedules: []models.EnrolledUnitSchedule{
					{ClassStartDate: futureDate.Format(time.RFC3339)},
				},
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		_, err := svc.ChangeSchedule(ctx, inscricaoID, "99999999999", request)
		if err == nil {
			t.Error("Expected error when CPF doesn't match")
		}
	})

	t.Run("ChangeSchedule fails for cancelled enrollment", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CPF:     "12345678900",
					CursoID: 1,
					Status:  models.StatusInscricaoCancelled,
				}, nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		futureDate := time.Now().Add(72 * time.Hour)
		request := &models.ScheduleChangeRequest{
			EnrolledUnit: &models.EnrolledUnit{
				Schedules: []models.EnrolledUnitSchedule{
					{ClassStartDate: futureDate.Format(time.RFC3339)},
				},
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		_, err := svc.ChangeSchedule(ctx, inscricaoID, "12345678900", request)
		if err == nil {
			t.Error("Expected error for cancelled enrollment")
		}
	})

	t.Run("ChangeSchedule fails within deadline", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CPF:     "12345678900",
					CursoID: 1,
					Status:  models.StatusInscricaoApproved,
				}, nil
			},
		}

		cursoRepo := &MockCursoRepository{
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return &models.Curso{
					ID: id,
					LocationClasses: []models.LocationClass{
						{Schedules: []models.CourseSchedule{{}, {}}},
					},
				}, nil
			},
		}

		// Class starts in 24 hours, but deadline is 48 hours
		nearFutureDate := time.Now().Add(24 * time.Hour)
		request := &models.ScheduleChangeRequest{
			EnrolledUnit: &models.EnrolledUnit{
				Schedules: []models.EnrolledUnitSchedule{
					{ClassStartDate: nearFutureDate.Format(time.RFC3339)},
				},
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{
			Enrollment: config.EnrollmentSettings{
				ScheduleChangeDeadlineHours: 48,
			},
		})

		_, err := svc.ChangeSchedule(ctx, inscricaoID, "12345678900", request)
		if err == nil {
			t.Error("Expected error when changing schedule within deadline")
		}
	})

	t.Run("ChangeSchedule fails when schedule is full", func(t *testing.T) {
		inscricaoID := uuid.New()
		newScheduleID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CPF:     "12345678900",
					CursoID: 1,
					Status:  models.StatusInscricaoApproved,
				}, nil
			},
		}

		cursoRepo := &MockCursoRepository{
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return &models.Curso{
					ID: id,
					LocationClasses: []models.LocationClass{
						{Schedules: []models.CourseSchedule{{}, {}}},
					},
				}, nil
			},
			CountEnrollmentsByScheduleIDsFunc: func(ctx context.Context, scheduleIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
				return map[uuid.UUID]int64{
					newScheduleID: 10, // Full
				}, nil
			},
			GetCourseScheduleByIDFunc: func(ctx context.Context, scheduleID uuid.UUID) (*models.CourseSchedule, error) {
				acceptingEnrollments := true
				return &models.CourseSchedule{
					ID:                   scheduleID,
					Vacancies:            10,
					AcceptingEnrollments: &acceptingEnrollments,
				}, nil
			},
		}

		futureDate := time.Now().Add(72 * time.Hour)
		request := &models.ScheduleChangeRequest{
			ScheduleID: &newScheduleID,
			EnrolledUnit: &models.EnrolledUnit{
				Schedules: []models.EnrolledUnitSchedule{
					{ClassStartDate: futureDate.Format(time.RFC3339)},
				},
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{
			Enrollment: config.EnrollmentSettings{
				ScheduleChangeDeadlineHours: 48,
			},
		})

		_, err := svc.ChangeSchedule(ctx, inscricaoID, "12345678900", request)
		if err == nil {
			t.Error("Expected error when schedule is full")
		}
	})

	t.Run("ChangeSchedule with remote schedule", func(t *testing.T) {
		inscricaoID := uuid.New()
		remoteScheduleID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CPF:     "12345678900",
					CursoID: 1,
					Status:  models.StatusInscricaoApproved,
				}, nil
			},
			UpdateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				return nil
			},
		}

		cursoRepo := &MockCursoRepository{
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return &models.Curso{
					ID: id,
					RemoteClass: &models.RemoteClass{
						Schedules: []models.RemoteSchedule{{}, {}},
					},
				}, nil
			},
			CountEnrollmentsByScheduleIDsFunc: func(ctx context.Context, scheduleIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
				return map[uuid.UUID]int64{
					remoteScheduleID: 3,
				}, nil
			},
			GetCourseScheduleByIDFunc: func(ctx context.Context, scheduleID uuid.UUID) (*models.CourseSchedule, error) {
				return nil, nil // Not a course schedule
			},
			GetRemoteScheduleByIDFunc: func(ctx context.Context, scheduleID uuid.UUID) (*models.RemoteSchedule, error) {
				acceptingEnrollments := true
				return &models.RemoteSchedule{
					ID:                   scheduleID,
					Vacancies:            10,
					AcceptingEnrollments: &acceptingEnrollments,
				}, nil
			},
		}

		futureDate := time.Now().Add(72 * time.Hour)
		request := &models.ScheduleChangeRequest{
			ScheduleID: &remoteScheduleID,
			EnrolledUnit: &models.EnrolledUnit{
				Schedules: []models.EnrolledUnitSchedule{
					{ClassStartDate: futureDate.Format(time.RFC3339)},
				},
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{
			Enrollment: config.EnrollmentSettings{
				ScheduleChangeDeadlineHours: 48,
			},
		})

		_, err := svc.ChangeSchedule(ctx, inscricaoID, "12345678900", request)
		if err != nil {
			t.Fatalf("ChangeSchedule with remote schedule failed: %v", err)
		}
	})

	t.Run("ChangeSchedule fails when schedule not accepting enrollments", func(t *testing.T) {
		inscricaoID := uuid.New()
		scheduleID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CPF:     "12345678900",
					CursoID: 1,
					Status:  models.StatusInscricaoApproved,
				}, nil
			},
		}

		cursoRepo := &MockCursoRepository{
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return &models.Curso{
					ID: id,
					LocationClasses: []models.LocationClass{
						{Schedules: []models.CourseSchedule{{}, {}}},
					},
				}, nil
			},
			CountEnrollmentsByScheduleIDsFunc: func(ctx context.Context, scheduleIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
				return map[uuid.UUID]int64{scheduleID: 0}, nil
			},
			GetCourseScheduleByIDFunc: func(ctx context.Context, sid uuid.UUID) (*models.CourseSchedule, error) {
				acceptingEnrollments := false
				return &models.CourseSchedule{
					ID:                   sid,
					Vacancies:            10,
					AcceptingEnrollments: &acceptingEnrollments,
				}, nil
			},
		}

		futureDate := time.Now().Add(72 * time.Hour)
		request := &models.ScheduleChangeRequest{
			ScheduleID: &scheduleID,
			EnrolledUnit: &models.EnrolledUnit{
				Schedules: []models.EnrolledUnitSchedule{
					{ClassStartDate: futureDate.Format(time.RFC3339)},
				},
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{
			Enrollment: config.EnrollmentSettings{
				ScheduleChangeDeadlineHours: 48,
			},
		})

		_, err := svc.ChangeSchedule(ctx, inscricaoID, "12345678900", request)
		if err == nil {
			t.Error("Expected error when schedule not accepting enrollments")
		}
	})

	t.Run("ChangeSchedule fails with invalid date format", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CPF:     "12345678900",
					CursoID: 1,
					Status:  models.StatusInscricaoApproved,
				}, nil
			},
		}

		cursoRepo := &MockCursoRepository{
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return &models.Curso{
					ID: id,
					LocationClasses: []models.LocationClass{
						{Schedules: []models.CourseSchedule{{}, {}}},
					},
				}, nil
			},
		}

		request := &models.ScheduleChangeRequest{
			EnrolledUnit: &models.EnrolledUnit{
				Schedules: []models.EnrolledUnitSchedule{
					{ClassStartDate: "invalid-date-format"},
				},
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		_, err := svc.ChangeSchedule(ctx, inscricaoID, "12345678900", request)
		if err == nil {
			t.Error("Expected error with invalid date format")
		}
	})

	t.Run("ChangeSchedule uses default deadline when config missing", func(t *testing.T) {
		inscricaoID := uuid.New()
		scheduleID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CPF:     "12345678900",
					CursoID: 1,
					Status:  models.StatusInscricaoApproved,
				}, nil
			},
			UpdateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				return nil
			},
		}

		cursoRepo := &MockCursoRepository{
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return &models.Curso{
					ID: id,
					LocationClasses: []models.LocationClass{
						{Schedules: []models.CourseSchedule{{}, {}}},
					},
				}, nil
			},
			CountEnrollmentsByScheduleIDsFunc: func(ctx context.Context, scheduleIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
				return map[uuid.UUID]int64{scheduleID: 0}, nil
			},
			GetCourseScheduleByIDFunc: func(ctx context.Context, sid uuid.UUID) (*models.CourseSchedule, error) {
				acceptingEnrollments := true
				return &models.CourseSchedule{
					ID:                   sid,
					Vacancies:            10,
					AcceptingEnrollments: &acceptingEnrollments,
				}, nil
			},
		}

		// 72 hours should be enough for default 48 hour deadline
		futureDate := time.Now().Add(72 * time.Hour)
		request := &models.ScheduleChangeRequest{
			ScheduleID: &scheduleID,
			EnrolledUnit: &models.EnrolledUnit{
				Schedules: []models.EnrolledUnitSchedule{
					{ClassStartDate: futureDate.Format(time.RFC3339)},
				},
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, nil)

		_, err := svc.ChangeSchedule(ctx, inscricaoID, "12345678900", request)
		if err != nil {
			t.Fatalf("ChangeSchedule should use default 48h deadline: %v", err)
		}
	})

	t.Run("ChangeSchedule fails when course has only one schedule", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CPF:     "12345678900",
					CursoID: 1,
					Status:  models.StatusInscricaoApproved,
				}, nil
			},
		}

		cursoRepo := &MockCursoRepository{
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return &models.Curso{
					ID: id,
					LocationClasses: []models.LocationClass{
						{Schedules: []models.CourseSchedule{{}}},
					},
				}, nil
			},
		}

		futureDate := time.Now().Add(72 * time.Hour)
		request := &models.ScheduleChangeRequest{
			EnrolledUnit: &models.EnrolledUnit{
				Schedules: []models.EnrolledUnitSchedule{
					{ClassStartDate: futureDate.Format(time.RFC3339)},
				},
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		_, err := svc.ChangeSchedule(ctx, inscricaoID, "12345678900", request)
		if err == nil {
			t.Error("Expected error when course has only one schedule")
		}
		if err != nil && !strings.Contains(err.Error(), "apenas uma turma") {
			t.Errorf("Expected 'apenas uma turma' error, got: %v", err)
		}
	})
}

// Additional tests to improve coverage

// TestInscricaoService_UpdateCertificate_AdditionalCases tests missing branches
func TestInscricaoService_UpdateCertificate_AdditionalCases(t *testing.T) {
	ctx := context.Background()

	t.Run("UpdateCertificate with GetByID error", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return nil, errors.New("database connection failed")
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.UpdateCertificate(ctx, 1, inscricaoID, "https://example.com/cert.pdf")
		if err == nil {
			t.Error("Expected error when GetByID fails")
		}
		if !strings.Contains(err.Error(), "erro ao verificar inscrição") {
			t.Errorf("Expected verification error message, got: %v", err)
		}
	})

	t.Run("UpdateCertificate when inscricao not found", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return nil, nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.UpdateCertificate(ctx, 1, inscricaoID, "https://example.com/cert.pdf")
		if err == nil {
			t.Error("Expected error when inscricao not found")
		}
		if !strings.Contains(err.Error(), "inscrição não encontrada") {
			t.Errorf("Expected not found error, got: %v", err)
		}
	})

	t.Run("UpdateCertificate for concluded inscricao", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CursoID: 1,
					Status:  models.StatusInscricaoConcluded,
				}, nil
			},
			UpdateCertificateFunc: func(ctx context.Context, id uuid.UUID, certificateURL string) error {
				return nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.UpdateCertificate(ctx, 1, inscricaoID, "https://example.com/cert.pdf")
		if err != nil {
			t.Fatalf("UpdateCertificate should work for concluded status: %v", err)
		}
	})

	t.Run("UpdateCertificate with repository update error", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CursoID: 1,
					Status:  models.StatusInscricaoApproved,
				}, nil
			},
			UpdateCertificateFunc: func(ctx context.Context, id uuid.UUID, certificateURL string) error {
				return errors.New("database write failed")
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.UpdateCertificate(ctx, 1, inscricaoID, "https://example.com/cert.pdf")
		if err == nil {
			t.Error("Expected error when repository update fails")
		}
	})

	t.Run("UpdateCertificate for rejected status should fail", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CursoID: 1,
					Status:  models.StatusInscricaoRejected,
				}, nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.UpdateCertificate(ctx, 1, inscricaoID, "https://example.com/cert.pdf")
		if err == nil {
			t.Error("Expected error for rejected status")
		}
		if !strings.Contains(err.Error(), "certificado só pode ser atribuído") {
			t.Errorf("Expected status validation error, got: %v", err)
		}
	})
}

// TestInscricaoService_Delete_AdditionalCases tests missing branches
func TestInscricaoService_Delete_AdditionalCases(t *testing.T) {
	ctx := context.Background()

	t.Run("Delete with GetByID error", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return nil, errors.New("connection timeout")
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.Delete(ctx, inscricaoID)
		if err == nil {
			t.Error("Expected error when GetByID fails")
		}
		if !strings.Contains(err.Error(), "erro ao verificar inscrição") {
			t.Errorf("Expected verification error, got: %v", err)
		}
	})

	t.Run("Delete with repository error", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{ID: inscricaoID}, nil
			},
			DeleteFunc: func(ctx context.Context, id uuid.UUID) error {
				return errors.New("foreign key constraint violation")
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.Delete(ctx, inscricaoID)
		if err == nil {
			t.Error("Expected error when Delete fails")
		}
		if !strings.Contains(err.Error(), "foreign key constraint") {
			t.Errorf("Expected constraint error, got: %v", err)
		}
	})
}

// TestInscricaoService_UpdateInscricao_AdditionalCases tests missing branches
func TestInscricaoService_UpdateInscricao_AdditionalCases(t *testing.T) {
	ctx := context.Background()

	t.Run("UpdateInscricao with GetByID error", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return nil, errors.New("query failed")
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		updateData := &models.InscricaoUpdateRequest{}
		err := svc.UpdateInscricao(ctx, inscricaoID, 1, updateData)
		if err == nil {
			t.Error("Expected error when GetByID fails")
		}
		if !strings.Contains(err.Error(), "erro ao verificar inscrição") {
			t.Errorf("Expected verification error, got: %v", err)
		}
	})

	t.Run("UpdateInscricao with repository update error", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CursoID: 1,
				}, nil
			},
			UpdateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				return errors.New("update failed")
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		newName := "Test"
		updateData := &models.InscricaoUpdateRequest{
			Name: &newName,
		}
		err := svc.UpdateInscricao(ctx, inscricaoID, 1, updateData)
		if err == nil {
			t.Error("Expected error when Update fails")
		}
	})

	t.Run("UpdateInscricao updates all fields correctly", func(t *testing.T) {
		inscricaoID := uuid.New()
		newName := "New Name"
		newEmail := "new@example.com"
		newPhone := "1234567890"
		newAdminNotes := "Admin note"
		enrolledUnit := &models.EnrolledUnit{ID: "unit-1"}

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CursoID: 1,
					Name:    "Old Name",
					Email:   "old@example.com",
					Phone:   "0000000000",
				}, nil
			},
			UpdateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				if inscricao.Name != newName {
					t.Errorf("Expected name %s, got %s", newName, inscricao.Name)
				}
				if inscricao.Email != newEmail {
					t.Errorf("Expected email %s, got %s", newEmail, inscricao.Email)
				}
				if inscricao.Phone != newPhone {
					t.Errorf("Expected phone %s, got %s", newPhone, inscricao.Phone)
				}
				if inscricao.AdminNotes != newAdminNotes {
					t.Errorf("Expected admin notes %s, got %s", newAdminNotes, inscricao.AdminNotes)
				}
				if inscricao.CustomFieldsData == nil || len(inscricao.CustomFieldsData) == 0 {
					t.Error("Expected custom fields to be updated")
				}
				if inscricao.EnrolledUnit == nil || inscricao.EnrolledUnit.ID != "unit-1" {
					t.Error("Expected enrolled unit to be updated")
				}
				return nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		customFields := []byte(`{"field1":"value1"}`)
		updateData := &models.InscricaoUpdateRequest{
			Name:             &newName,
			Email:            &newEmail,
			Phone:            &newPhone,
			AdminNotes:       &newAdminNotes,
			CustomFieldsData: customFields,
			EnrolledUnit:     enrolledUnit,
		}

		err := svc.UpdateInscricao(ctx, inscricaoID, 1, updateData)
		if err != nil {
			t.Fatalf("UpdateInscricao failed: %v", err)
		}
	})

	t.Run("UpdateInscricao with nil fields does not update", func(t *testing.T) {
		inscricaoID := uuid.New()
		originalName := "Original Name"
		originalEmail := "original@example.com"

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CursoID: 1,
					Name:    originalName,
					Email:   originalEmail,
				}, nil
			},
			UpdateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				if inscricao.Name != originalName {
					t.Errorf("Name should not change, got %s", inscricao.Name)
				}
				if inscricao.Email != originalEmail {
					t.Errorf("Email should not change, got %s", inscricao.Email)
				}
				return nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		updateData := &models.InscricaoUpdateRequest{
			// All fields are nil
		}

		err := svc.UpdateInscricao(ctx, inscricaoID, 1, updateData)
		if err != nil {
			t.Fatalf("UpdateInscricao failed: %v", err)
		}
	})
}

// TestInscricaoService_UpdateStatus_AdditionalCases tests missing branches
func TestInscricaoService_UpdateStatus_AdditionalCases(t *testing.T) {
	ctx := context.Background()

	t.Run("UpdateStatus with curso fetch error during email", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CursoID: 1,
					Status:  models.StatusInscricaoPending,
				}, nil
			},
			UpdateStatusFunc: func(ctx context.Context, id uuid.UUID, status models.StatusInscricao, reason, adminNotes string) error {
				return nil
			},
		}

		cursoRepo := &MockCursoRepository{
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return nil, errors.New("curso fetch failed")
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		// Should succeed despite curso fetch error (email is best-effort)
		err := svc.UpdateStatus(ctx, inscricaoID, models.StatusInscricaoApproved, "", "")
		if err != nil {
			t.Fatalf("UpdateStatus should succeed even if email fails: %v", err)
		}
	})

	t.Run("UpdateStatus to cancelled status", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CursoID: 1,
					Status:  models.StatusInscricaoApproved,
				}, nil
			},
			UpdateStatusFunc: func(ctx context.Context, id uuid.UUID, status models.StatusInscricao, reason, adminNotes string) error {
				if status != models.StatusInscricaoCancelled {
					t.Errorf("Expected status CANCELLED, got %s", status)
				}
				return nil
			},
		}

		cursoRepo := &MockCursoRepository{
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return &models.Curso{ID: 1}, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.UpdateStatus(ctx, inscricaoID, models.StatusInscricaoCancelled, "User requested", "")
		if err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}
	})

	t.Run("UpdateStatus to concluded status", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CursoID: 1,
					Status:  models.StatusInscricaoApproved,
				}, nil
			},
			UpdateStatusFunc: func(ctx context.Context, id uuid.UUID, status models.StatusInscricao, reason, adminNotes string) error {
				if status != models.StatusInscricaoConcluded {
					t.Errorf("Expected status CONCLUDED, got %s", status)
				}
				return nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.UpdateStatus(ctx, inscricaoID, models.StatusInscricaoConcluded, "", "")
		if err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}
	})
}

// TestInscricaoService_UpdateMultipleStatus_AdditionalCases tests missing branches
func TestInscricaoService_UpdateMultipleStatus_AdditionalCases(t *testing.T) {
	ctx := context.Background()

	t.Run("UpdateMultipleStatus to cancelled status", func(t *testing.T) {
		ids := []uuid.UUID{uuid.New(), uuid.New()}

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      id,
					CursoID: 1,
					Status:  models.StatusInscricaoApproved,
				}, nil
			},
			UpdateMultipleStatusFunc: func(ctx context.Context, inscricaoIDs []uuid.UUID, status models.StatusInscricao, reason, adminNotes string) (int, error) {
				return len(inscricaoIDs), nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		// Cancelled status should not trigger email collection
		count, err := svc.UpdateMultipleStatus(ctx, ids, models.StatusInscricaoCancelled, "", "")
		if err != nil {
			t.Fatalf("UpdateMultipleStatus failed: %v", err)
		}
		if count != 2 {
			t.Errorf("Expected count 2, got %d", count)
		}
	})

	t.Run("UpdateMultipleStatus with curso fetch error during email", func(t *testing.T) {
		ids := []uuid.UUID{uuid.New()}

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      id,
					CursoID: 1,
					Status:  models.StatusInscricaoPending,
				}, nil
			},
			UpdateMultipleStatusFunc: func(ctx context.Context, inscricaoIDs []uuid.UUID, status models.StatusInscricao, reason, adminNotes string) (int, error) {
				return len(inscricaoIDs), nil
			},
		}

		cursoRepo := &MockCursoRepository{
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return nil, errors.New("curso not found")
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		// Should succeed despite curso fetch error (email is best-effort)
		count, err := svc.UpdateMultipleStatus(ctx, ids, models.StatusInscricaoApproved, "", "")
		if err != nil {
			t.Fatalf("UpdateMultipleStatus should succeed even if email fails: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected count 1, got %d", count)
		}
	})

	t.Run("UpdateMultipleStatus to rejected with reason", func(t *testing.T) {
		ids := []uuid.UUID{uuid.New()}

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      id,
					CursoID: 1,
					Status:  models.StatusInscricaoPending,
				}, nil
			},
			UpdateMultipleStatusFunc: func(ctx context.Context, inscricaoIDs []uuid.UUID, status models.StatusInscricao, reason, adminNotes string) (int, error) {
				if reason != "Failed requirements" {
					t.Errorf("Expected reason 'Failed requirements', got %s", reason)
				}
				return len(inscricaoIDs), nil
			},
		}

		cursoRepo := &MockCursoRepository{
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return &models.Curso{ID: 1}, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		count, err := svc.UpdateMultipleStatus(ctx, ids, models.StatusInscricaoRejected, "Failed requirements", "Internal notes")
		if err != nil {
			t.Fatalf("UpdateMultipleStatus failed: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected count 1, got %d", count)
		}
	})

	t.Run("UpdateMultipleStatus no email service configured", func(t *testing.T) {
		ids := []uuid.UUID{uuid.New()}

		inscricaoRepo := &MockInscricaoRepository{
			UpdateMultipleStatusFunc: func(ctx context.Context, inscricaoIDs []uuid.UUID, status models.StatusInscricao, reason, adminNotes string) (int, error) {
				return len(inscricaoIDs), nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		// No email service configured
		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		count, err := svc.UpdateMultipleStatus(ctx, ids, models.StatusInscricaoApproved, "", "")
		if err != nil {
			t.Fatalf("UpdateMultipleStatus should work without email service: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected count 1, got %d", count)
		}
	})
}

// TestInscricaoService_ChangeSchedule_AdditionalCases tests missing branches
func TestInscricaoService_ChangeSchedule_AdditionalCases(t *testing.T) {
	ctx := context.Background()

	t.Run("ChangeSchedule with GetByID error", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return nil, errors.New("database timeout")
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		request := &models.ScheduleChangeRequest{
			EnrolledUnit: &models.EnrolledUnit{
				Schedules: []models.EnrolledUnitSchedule{{ClassStartDate: time.Now().Add(72 * time.Hour).Format(time.RFC3339)}},
			},
		}

		_, err := svc.ChangeSchedule(ctx, inscricaoID, "12345678900", request)
		if err == nil {
			t.Error("Expected error when GetByID fails")
		}
		if !strings.Contains(err.Error(), "erro ao buscar inscrição") {
			t.Errorf("Expected fetch error, got: %v", err)
		}
	})

	t.Run("ChangeSchedule for concluded enrollment", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CPF:     "12345678900",
					CursoID: 1,
					Status:  models.StatusInscricaoConcluded,
				}, nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		request := &models.ScheduleChangeRequest{
			EnrolledUnit: &models.EnrolledUnit{
				Schedules: []models.EnrolledUnitSchedule{{ClassStartDate: time.Now().Add(72 * time.Hour).Format(time.RFC3339)}},
			},
		}

		_, err := svc.ChangeSchedule(ctx, inscricaoID, "12345678900", request)
		if err == nil {
			t.Error("Expected error for concluded enrollment")
		}
	})

	t.Run("ChangeSchedule without enrolled unit", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CPF:     "12345678900",
					CursoID: 1,
					Status:  models.StatusInscricaoApproved,
				}, nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		request := &models.ScheduleChangeRequest{
			EnrolledUnit: nil,
		}

		_, err := svc.ChangeSchedule(ctx, inscricaoID, "12345678900", request)
		if err == nil {
			t.Error("Expected error when enrolled unit is nil")
		}
	})

	t.Run("ChangeSchedule with empty schedules", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CPF:     "12345678900",
					CursoID: 1,
					Status:  models.StatusInscricaoApproved,
				}, nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		request := &models.ScheduleChangeRequest{
			EnrolledUnit: &models.EnrolledUnit{
				Schedules: []models.EnrolledUnitSchedule{},
			},
		}

		_, err := svc.ChangeSchedule(ctx, inscricaoID, "12345678900", request)
		if err == nil {
			t.Error("Expected error when schedules are empty")
		}
	})

	t.Run("ChangeSchedule with missing class start date", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CPF:     "12345678900",
					CursoID: 1,
					Status:  models.StatusInscricaoApproved,
				}, nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		request := &models.ScheduleChangeRequest{
			EnrolledUnit: &models.EnrolledUnit{
				Schedules: []models.EnrolledUnitSchedule{{ClassStartDate: ""}},
			},
		}

		_, err := svc.ChangeSchedule(ctx, inscricaoID, "12345678900", request)
		if err == nil {
			t.Error("Expected error when class start date is missing")
		}
	})

	t.Run("ChangeSchedule with enrollment count error", func(t *testing.T) {
		inscricaoID := uuid.New()
		scheduleID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CPF:     "12345678900",
					CursoID: 1,
					Status:  models.StatusInscricaoApproved,
				}, nil
			},
		}

		cursoRepo := &MockCursoRepository{
			CountEnrollmentsByScheduleIDsFunc: func(ctx context.Context, scheduleIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
				return nil, errors.New("count query failed")
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		futureDate := time.Now().Add(72 * time.Hour)
		request := &models.ScheduleChangeRequest{
			ScheduleID: &scheduleID,
			EnrolledUnit: &models.EnrolledUnit{
				Schedules: []models.EnrolledUnitSchedule{{ClassStartDate: futureDate.Format(time.RFC3339)}},
			},
		}

		_, err := svc.ChangeSchedule(ctx, inscricaoID, "12345678900", request)
		if err == nil {
			t.Error("Expected error when enrollment count fails")
		}
	})

	t.Run("ChangeSchedule with GetCourseScheduleByID error", func(t *testing.T) {
		inscricaoID := uuid.New()
		scheduleID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CPF:     "12345678900",
					CursoID: 1,
					Status:  models.StatusInscricaoApproved,
				}, nil
			},
		}

		cursoRepo := &MockCursoRepository{
			CountEnrollmentsByScheduleIDsFunc: func(ctx context.Context, scheduleIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
				return map[uuid.UUID]int64{scheduleID: 5}, nil
			},
			GetCourseScheduleByIDFunc: func(ctx context.Context, sid uuid.UUID) (*models.CourseSchedule, error) {
				return nil, errors.New("schedule fetch failed")
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		futureDate := time.Now().Add(72 * time.Hour)
		request := &models.ScheduleChangeRequest{
			ScheduleID: &scheduleID,
			EnrolledUnit: &models.EnrolledUnit{
				Schedules: []models.EnrolledUnitSchedule{{ClassStartDate: futureDate.Format(time.RFC3339)}},
			},
		}

		_, err := svc.ChangeSchedule(ctx, inscricaoID, "12345678900", request)
		if err == nil {
			t.Error("Expected error when GetCourseScheduleByID fails")
		}
	})

	t.Run("ChangeSchedule with remote schedule not accepting enrollments", func(t *testing.T) {
		inscricaoID := uuid.New()
		scheduleID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CPF:     "12345678900",
					CursoID: 1,
					Status:  models.StatusInscricaoApproved,
				}, nil
			},
		}

		cursoRepo := &MockCursoRepository{
			CountEnrollmentsByScheduleIDsFunc: func(ctx context.Context, scheduleIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
				return map[uuid.UUID]int64{scheduleID: 5}, nil
			},
			GetCourseScheduleByIDFunc: func(ctx context.Context, sid uuid.UUID) (*models.CourseSchedule, error) {
				return nil, nil // Not found in course schedules
			},
			GetRemoteScheduleByIDFunc: func(ctx context.Context, sid uuid.UUID) (*models.RemoteSchedule, error) {
				acceptingEnrollments := false
				return &models.RemoteSchedule{
					ID:                   sid,
					Vacancies:            10,
					AcceptingEnrollments: &acceptingEnrollments,
				}, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		futureDate := time.Now().Add(72 * time.Hour)
		request := &models.ScheduleChangeRequest{
			ScheduleID: &scheduleID,
			EnrolledUnit: &models.EnrolledUnit{
				Schedules: []models.EnrolledUnitSchedule{{ClassStartDate: futureDate.Format(time.RFC3339)}},
			},
		}

		_, err := svc.ChangeSchedule(ctx, inscricaoID, "12345678900", request)
		if err == nil {
			t.Error("Expected error when remote schedule not accepting enrollments")
		}
	})

	t.Run("ChangeSchedule with remote schedule full", func(t *testing.T) {
		inscricaoID := uuid.New()
		scheduleID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CPF:     "12345678900",
					CursoID: 1,
					Status:  models.StatusInscricaoApproved,
				}, nil
			},
		}

		cursoRepo := &MockCursoRepository{
			CountEnrollmentsByScheduleIDsFunc: func(ctx context.Context, scheduleIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
				return map[uuid.UUID]int64{scheduleID: 10}, nil
			},
			GetCourseScheduleByIDFunc: func(ctx context.Context, sid uuid.UUID) (*models.CourseSchedule, error) {
				return nil, nil
			},
			GetRemoteScheduleByIDFunc: func(ctx context.Context, sid uuid.UUID) (*models.RemoteSchedule, error) {
				acceptingEnrollments := true
				return &models.RemoteSchedule{
					ID:                   sid,
					Vacancies:            10,
					AcceptingEnrollments: &acceptingEnrollments,
				}, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		futureDate := time.Now().Add(72 * time.Hour)
		request := &models.ScheduleChangeRequest{
			ScheduleID: &scheduleID,
			EnrolledUnit: &models.EnrolledUnit{
				Schedules: []models.EnrolledUnitSchedule{{ClassStartDate: futureDate.Format(time.RFC3339)}},
			},
		}

		_, err := svc.ChangeSchedule(ctx, inscricaoID, "12345678900", request)
		if err == nil {
			t.Error("Expected error when remote schedule is full")
		}
	})

	t.Run("ChangeSchedule with course schedule not accepting enrollments", func(t *testing.T) {
		inscricaoID := uuid.New()
		scheduleID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CPF:     "12345678900",
					CursoID: 1,
					Status:  models.StatusInscricaoApproved,
				}, nil
			},
		}

		cursoRepo := &MockCursoRepository{
			CountEnrollmentsByScheduleIDsFunc: func(ctx context.Context, scheduleIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
				return map[uuid.UUID]int64{scheduleID: 5}, nil
			},
			GetCourseScheduleByIDFunc: func(ctx context.Context, sid uuid.UUID) (*models.CourseSchedule, error) {
				acceptingEnrollments := false
				return &models.CourseSchedule{
					ID:                   sid,
					Vacancies:            10,
					AcceptingEnrollments: &acceptingEnrollments,
				}, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		futureDate := time.Now().Add(72 * time.Hour)
		request := &models.ScheduleChangeRequest{
			ScheduleID: &scheduleID,
			EnrolledUnit: &models.EnrolledUnit{
				Schedules: []models.EnrolledUnitSchedule{{ClassStartDate: futureDate.Format(time.RFC3339)}},
			},
		}

		_, err := svc.ChangeSchedule(ctx, inscricaoID, "12345678900", request)
		if err == nil {
			t.Error("Expected error when course schedule not accepting enrollments")
		}
	})

	t.Run("ChangeSchedule with repository update error", func(t *testing.T) {
		inscricaoID := uuid.New()
		scheduleID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CPF:     "12345678900",
					CursoID: 1,
					Status:  models.StatusInscricaoApproved,
				}, nil
			},
			UpdateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				return errors.New("update failed")
			},
		}

		cursoRepo := &MockCursoRepository{
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return &models.Curso{
					ID: id,
					LocationClasses: []models.LocationClass{
						{Schedules: []models.CourseSchedule{{}, {}}},
					},
				}, nil
			},
			CountEnrollmentsByScheduleIDsFunc: func(ctx context.Context, scheduleIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
				return map[uuid.UUID]int64{scheduleID: 5}, nil
			},
			GetCourseScheduleByIDFunc: func(ctx context.Context, sid uuid.UUID) (*models.CourseSchedule, error) {
				acceptingEnrollments := true
				return &models.CourseSchedule{
					ID:                   sid,
					Vacancies:            10,
					AcceptingEnrollments: &acceptingEnrollments,
				}, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		futureDate := time.Now().Add(72 * time.Hour)
		request := &models.ScheduleChangeRequest{
			ScheduleID: &scheduleID,
			EnrolledUnit: &models.EnrolledUnit{
				Schedules: []models.EnrolledUnitSchedule{{ClassStartDate: futureDate.Format(time.RFC3339)}},
			},
		}

		_, err := svc.ChangeSchedule(ctx, inscricaoID, "12345678900", request)
		if err == nil {
			t.Error("Expected error when Update fails")
		}
		if !strings.Contains(err.Error(), "erro ao atualizar inscrição") {
			t.Errorf("Expected update error, got: %v", err)
		}
	})

	t.Run("ChangeSchedule with remote schedule success", func(t *testing.T) {
		inscricaoID := uuid.New()
		scheduleID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CPF:     "12345678900",
					CursoID: 1,
					Status:  models.StatusInscricaoApproved,
				}, nil
			},
			UpdateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				if inscricao.ScheduleID == nil || *inscricao.ScheduleID != scheduleID {
					t.Error("Expected schedule ID to be updated")
				}
				return nil
			},
		}

		cursoRepo := &MockCursoRepository{
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return &models.Curso{
					ID: id,
					RemoteClass: &models.RemoteClass{
						Schedules: []models.RemoteSchedule{{}, {}},
					},
				}, nil
			},
			CountEnrollmentsByScheduleIDsFunc: func(ctx context.Context, scheduleIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
				return map[uuid.UUID]int64{scheduleID: 3}, nil
			},
			GetCourseScheduleByIDFunc: func(ctx context.Context, sid uuid.UUID) (*models.CourseSchedule, error) {
				return nil, nil // Not found in course schedules
			},
			GetRemoteScheduleByIDFunc: func(ctx context.Context, sid uuid.UUID) (*models.RemoteSchedule, error) {
				acceptingEnrollments := true
				return &models.RemoteSchedule{
					ID:                   sid,
					Vacancies:            10,
					AcceptingEnrollments: &acceptingEnrollments,
				}, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		futureDate := time.Now().Add(72 * time.Hour)
		request := &models.ScheduleChangeRequest{
			ScheduleID: &scheduleID,
			EnrolledUnit: &models.EnrolledUnit{
				Schedules: []models.EnrolledUnitSchedule{{ClassStartDate: futureDate.Format(time.RFC3339)}},
			},
		}

		result, err := svc.ChangeSchedule(ctx, inscricaoID, "12345678900", request)
		if err != nil {
			t.Fatalf("ChangeSchedule failed: %v", err)
		}
		if result == nil {
			t.Fatal("Expected inscricao result")
		}
	})

	t.Run("ChangeSchedule with GetRemoteScheduleByID error", func(t *testing.T) {
		inscricaoID := uuid.New()
		scheduleID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CPF:     "12345678900",
					CursoID: 1,
					Status:  models.StatusInscricaoApproved,
				}, nil
			},
		}

		cursoRepo := &MockCursoRepository{
			CountEnrollmentsByScheduleIDsFunc: func(ctx context.Context, scheduleIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
				return map[uuid.UUID]int64{scheduleID: 5}, nil
			},
			GetCourseScheduleByIDFunc: func(ctx context.Context, sid uuid.UUID) (*models.CourseSchedule, error) {
				return nil, nil
			},
			GetRemoteScheduleByIDFunc: func(ctx context.Context, sid uuid.UUID) (*models.RemoteSchedule, error) {
				return nil, errors.New("remote schedule fetch failed")
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		futureDate := time.Now().Add(72 * time.Hour)
		request := &models.ScheduleChangeRequest{
			ScheduleID: &scheduleID,
			EnrolledUnit: &models.EnrolledUnit{
				Schedules: []models.EnrolledUnitSchedule{{ClassStartDate: futureDate.Format(time.RFC3339)}},
			},
		}

		_, err := svc.ChangeSchedule(ctx, inscricaoID, "12345678900", request)
		if err == nil {
			t.Error("Expected error when GetRemoteScheduleByID fails")
		}
	})

	t.Run("ChangeSchedule with both schedules not found", func(t *testing.T) {
		inscricaoID := uuid.New()
		scheduleID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CPF:     "12345678900",
					CursoID: 1,
					Status:  models.StatusInscricaoApproved,
				}, nil
			},
		}

		cursoRepo := &MockCursoRepository{
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return &models.Curso{
					ID: id,
					LocationClasses: []models.LocationClass{
						{Schedules: []models.CourseSchedule{{}, {}}},
					},
				}, nil
			},
			CountEnrollmentsByScheduleIDsFunc: func(ctx context.Context, scheduleIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
				return map[uuid.UUID]int64{scheduleID: 5}, nil
			},
			GetCourseScheduleByIDFunc: func(ctx context.Context, sid uuid.UUID) (*models.CourseSchedule, error) {
				return nil, nil
			},
			GetRemoteScheduleByIDFunc: func(ctx context.Context, sid uuid.UUID) (*models.RemoteSchedule, error) {
				return nil, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		futureDate := time.Now().Add(72 * time.Hour)
		request := &models.ScheduleChangeRequest{
			ScheduleID: &scheduleID,
			EnrolledUnit: &models.EnrolledUnit{
				Schedules: []models.EnrolledUnitSchedule{{ClassStartDate: futureDate.Format(time.RFC3339)}},
			},
		}

		_, err := svc.ChangeSchedule(ctx, inscricaoID, "12345678900", request)
		if err == nil {
			t.Error("Expected error when schedule not found")
		}
		if !strings.Contains(err.Error(), "turma não encontrada") {
			t.Errorf("Expected not found error, got: %v", err)
		}
	})
}

func TestNewInscricaoService(t *testing.T) {
	service := services.NewInscricaoService(nil, nil, nil, nil, nil, &config.AppConfig{})

	if service == nil {
		t.Error("NewInscricaoService() returned nil")
	}
}

func TestNewInscricaoServiceWithInterface(t *testing.T) {
	inscricaoRepo := &MockInscricaoRepository{}
	cursoRepo := &MockCursoRepository{}
	service := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

	if service == nil {
		t.Error("NewInscricaoServiceWithInterface() returned nil")
	}
}
